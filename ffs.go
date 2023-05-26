package main

import (
	"fmt"
	"syscall/js"
)

type RunState struct {
	romLoaded     bool
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
	rsObj["inDebug"] = rs.inDebug
	rsObj["waitingForKey"] = rs.waitingForKey
	return rsObj
}

type JsIO struct {
	runState     *RunState
	keysPressed  *[16]bool
	lastPressed  chan byte
	lastReleased chan byte
}

func (jsIO JsIO) Draw(pixels Pixels) {
	// convert Pixels type to []any type which JS can only support
	pixelsJs := []any{}
	for y, row := range pixels {
		for x, col := range row {
			if col == 1 {
				pixelsJs = append(pixelsJs, []any{x, y})
			}
		}
	}
	js.Global().Get("Chip8").Call("draw", pixelsJs)
}

func (jsIO JsIO) ClearScreen() {
	js.Global().Get("Chip8").Call("clearDisplay")
}

func (jsIO JsIO) WaitKeyPress() (byte, bool) {
	select {
	case <-jsIO.lastPressed:
		jsCall := js.Global().Get("Chip8").Call("waitForKeyPress")
		jsIO.runState.setState(func(rs *RunState) { rs.waitingForKey = true })
		lastReleased := <-jsIO.lastReleased
		jsCall.Call("onRelease", lastReleased)
		return lastReleased, true
	case <-jsIO.lastReleased: // get rid of any previous key that was pressed
		return 0, false
	default:
		return 0, false
	}
}

func (jsIO JsIO) GetKeysPressed() [16]bool {
	return *jsIO.keysPressed
}

func setup() {
	runState := newRunState()
	runParams := RunParams{}
	step := make(chan any, 1)
	jsIO := JsIO{
		runState:     &runState,
		lastPressed:  make(chan byte, 1),
		lastReleased: make(chan byte, 1),
	}
	var vm Vm

	js.Global().Set("toggleDebug", js.FuncOf(func(this js.Value, args []js.Value) any {
		runState.setState(func(rs *RunState) {
			rs.inDebug = !rs.inDebug
			if !rs.inDebug {
				step <- true
			}
		})
		return runState.inDebug
	}))

	js.Global().Set("nextInst", js.FuncOf(func(this js.Value, args []js.Value) any {
		if runState.waitingForKey || !runState.romLoaded || vm.Done {
			return nil
		}

		step <- true
		return nil
	}))

	js.Global().Set("setKey", js.FuncOf(func(this js.Value, args []js.Value) any {
		key := args[0].Int()
		pressed := args[1].Bool()
		jsIO.keysPressed[key] = pressed
		// discard the value of either channel if they are already filled
		select {
		case <-jsIO.lastPressed:
		case <-jsIO.lastReleased:
		default:
		}
		if pressed {
			jsIO.lastPressed <- byte(key)
		} else {
			jsIO.lastReleased <- byte(key)
		}
		return nil
	}))

	js.Global().Set("setInstsPerFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		runParams.instCount = args[0].Int()
		return nil
	}))

	// `createNewVm` will be called from JS-space and thus from another goroutine;
	// better to keep everything in a single goroutine as much as possible so let's
	// make a channel that will expect bytes coming from JS. In effect, this
	// function will be a blocking one until `createNewVm` is called from JS.
	rom := make(chan []byte, 1)
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

	loop := make(chan bool, 1)
	for {
		select {
		case newRom := <-rom:
			runState = newRunState()
			jsIO.keysPressed = &[16]bool{}
			newVm, err := NewVm(newRom, jsIO)
			vm = newVm
			if err != nil {
				panic(err)
			}
			runState.setState(func(rs *RunState) { rs.romLoaded = true })

			// start the run loop
			select {
			case loop <- true:
			default: // if the loop channel has already been filled, do nothing
			}
		case <-loop:
			commVmState := vmState(&vm)
			if runState.inDebug {
				<-step
				runParams.instCount = 1
				runParams.frameDuration = 1
				commVmState()
			}

			if vm.Done {
				fmt.Println("Program executed.")
				continue
			}

			err := vm.Run(runParams)
			if err != nil {
				fmt.Println(err)
			}
			loop <- true
		}
	}
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
