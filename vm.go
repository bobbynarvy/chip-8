package main

import (
	"errors"
	"fmt"
)

type Pixels [32][64]byte

type Vm struct {
	Mem          []byte
	Stack        [16]uint16
	Regs         [16]byte
	I            uint16   // register used mostly to store memory addresses
	DT           byte     // delay timer
	ST           byte     // sound timer
	Pc           uint16   // program counter
	Sp           byte     // stack pointer
	Keys         [16]bool // represents the 16-key keypad; a true value means the key corresponding key is pressed
	Pixels       Pixels
	ClearScreen  func()
	Draw         func(bytes Pixels)
	WaitKeyPress func() byte
	Done         bool
	repeatCnt    byte
}

func NewVm(rom []byte) (Vm, error) {
	if 0x200+len(rom) > 0xFFF {
		return Vm{}, errors.New("ROM size exceeds RAM limit")
	}

	// The first 0x200 bytes in RAM are reserved for
	// the CHIP-8 Interpreter.
	// The first 80 locations (16 chars x 5 bytes) in mem are used
	// to store the sprites representing the hex digits 0 to F.
	hexSprites := []byte{
		0xF1, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	mem := make([]byte, 0xFFF) // initialize RAM with the first reserved 0x200 bytes
	copy(mem, hexSprites)
	copy(mem[0x200:], rom) // copy the ROM into RAM

	return Vm{
		Mem: mem,
		Pc:  0x200,
	}, nil
}

func (vm *Vm) trace(b1, b2 byte) func(string) string {
	instInfo := fmt.Sprintf("%3x %2x %2x   ", vm.Pc, b1, b2)
	return func(instDesc string) string {
		assembly := fmt.Sprintf(instInfo + instDesc)
		return assembly
	}
}

// Increment the program counter
func (vm *Vm) incPc() {
	// Since instructions are 2 bytes long; the next instruction
	// shouldn't be the byte after the first byte but the one after that
	vm.Pc += 2
}

func (vm *Vm) skipIf(cond bool) {
	if cond {
		vm.incPc()
	}
}

func (vm *Vm) setVF1If(cond bool) {
	vm.Regs[0xF] = 0
	if cond {
		vm.Regs[0xF] = 1
	}
}

func (vm *Vm) Run() error {
	byte1, byte2 := vm.Mem[vm.Pc], vm.Mem[vm.Pc+1]
	trace := vm.trace(byte1, byte2)
	vm.incPc()

	inst, err := getInstruction(byte1, byte2)
	if err != nil {
		return err
	}

	trace(inst.assembly)
	inst.execFn(vm)
	return nil
}
