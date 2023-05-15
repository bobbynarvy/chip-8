package main

import (
	"fmt"
	"syscall/js"
)

type RunState struct {
	romLoaded     bool
	paused        bool
	inDebug       bool
	waitingForKey bool
}

func newRunState() RunState {
	rs := RunState{}
	js.Global().Get("Chip8").Call("onRunStateInit", rs.toJsObj())
	return rs
}

func (rs *RunState) setState(setter func(rs *RunState)) {
	setter(rs)
	js.Global().Get("Chip8").Call("onRunStateUpdate", rs.toJsObj())
}

func (rs *RunState) toJsObj() map[string]any {
	rsObj := make(map[string]any)
	rsObj["romLoaded"] = rs.romLoaded
	rsObj["paused"] = rs.paused
	rsObj["inDebug"] = rs.inDebug
	rsObj["waitingForKey"] = rs.waitingForKey
	return rsObj
}

type JsIO struct {
	runState    *RunState
	keysPressed *[16]bool
}

func (jsIO JsIO) Draw(pixels Pixels) {
	// convert Pixels type to [][]any type which JS can only support
	pixelsJs := []any{}
	for _, row := range pixels {
		cols := []any{}
		for _, col := range row {
			cols = append(cols, col)
		}
		pixelsJs = append(pixelsJs, cols)
	}
	js.Global().Get("Chip8").Call("draw", pixelsJs)
}

func (jsIO JsIO) ClearScreen() {
	js.Global().Get("Chip8").Call("clearDisplay")
}

func (jsIO JsIO) WaitKeyPress() byte {
	wait := make(chan byte)
	jsIO.runState.setState(func(rs *RunState) { rs.waitingForKey = true })
	js.Global().Get("Chip8").Call("waitForKeyPress").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		wait <- byte(args[0].Int())
		jsIO.runState.setState(func(rs *RunState) { rs.waitingForKey = false })
		return nil
	}))
	return <-wait
}

func (jsIO JsIO) GetKeysPressed() [16]bool {
	return *jsIO.keysPressed
}

func setup() Vm {
	// Set up run state and the functions that will be used client-side to
	// manipulate the run state
	runState := newRunState()
	step := make(chan any, 1)
	paused := make(chan bool)
	keysPressed := [16]bool{}

	js.Global().Set("toggleDebug", js.FuncOf(func(this js.Value, args []js.Value) any {
		runState.setState(func(rs *RunState) {
			rs.inDebug = !rs.inDebug
			if !rs.inDebug {
				step <- true
			}
		})
		return runState.inDebug
	}))

	js.Global().Set("togglePause", js.FuncOf(func(this js.Value, args []js.Value) any {
		paused <- !runState.paused
		runState.setState(func(rs *RunState) { rs.paused = !runState.paused })
		return runState.paused
	}))

	js.Global().Set("nextInst", js.FuncOf(func(this js.Value, args []js.Value) any {
		if runState.waitingForKey || runState.paused || !runState.romLoaded {
			return nil
		}

		step <- true
		return nil
	}))

	js.Global().Set("setKey", js.FuncOf(func(this js.Value, args []js.Value) any {
		key := args[0].Int()
		keysPressed[key] = args[1].Bool()
		return nil
	}))

	// `createNewVm` will be called from JS-space and thus from another goroutine;
	// better to keep everything in a single goroutine as much as possible so let's
	// make a channel that will expect bytes coming from JS. In effect, this
	// function will be a blocking one until `createNewVm` is called from JS.
	rom := make(chan []byte)
	js.Global().Set("createNewVm", js.FuncOf(func(this js.Value, args []js.Value) any {
		// args should be a Uint8Array in the JS-space;
		// let's convert them to bytes that can be used by the VM
		bytes := make([]byte, len(args))
		for i, num := range args {
			b := byte(num.Int())
			bytes[i] = b
		}
		rom <- bytes
		return nil
	}))

	jsIO := JsIO{
		runState:    &runState,
		keysPressed: &keysPressed,
	}
	// Wait until the ROM is loaded then
	// initialize the VM
	vm, err := NewVm(<-rom, jsIO)
	if err != nil {
		panic(err)
	}
	runState.setState(func(rs *RunState) { rs.romLoaded = true })

	commVmState := vmState(&vm)
	for {
		select {
		case <-paused:
			<-paused
		default:
			if runState.inDebug {
				<-step
				commVmState()
			}

			if vm.Done {
				fmt.Println("Program executed.")
				return vm
			}

			err := vm.Run()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return vm
}

func vmState(vm *Vm) func() {
	stack := make([]any, len(vm.Stack))
	regs := make([]any, len(vm.Regs))
	return func() {
		for i, v := range vm.Stack {
			stack[i] = v
		}
		for i, v := range vm.Regs {
			regs[i] = v
		}
		state := make(map[string]any)
		state["I"] = vm.I
		state["Pc"] = vm.Pc
		state["Sp"] = vm.Sp
		state["DT"] = vm.DT
		state["Stack"] = stack
		state["Done"] = vm.Done
		state["Regs"] = regs
		byte1, byte2 := vm.Mem[vm.Pc], vm.Mem[vm.Pc+1]
		inst, _ := getInstruction(byte1, byte2)
		state["Assembly"] = vm.trace(byte1, byte2)(inst.assembly)
		js.Global().Get("Chip8").Call("onVmUpdate", state)
	}
}
