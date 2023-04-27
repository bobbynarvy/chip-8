package main

import (
	"errors"
	"fmt"
)

type Ram [0xFFF]byte

type Vm struct {
	Mem   *Ram
	Stack [16]uint16
	Regs  [16]byte
	Delay uint16
	Sound uint16
	Pc    uint16 // program counter
	Sp    byte   // stack pointer
}

func upperBits(b byte) byte {
	return (b & 0xF0) >> 4
}

func mergeBytePair(b1, b2 byte) uint16 {
	return (uint16(b1) << 8) | uint16(b2)
}

func NewVm(mem *Ram) Vm {
	return Vm{
		Mem: mem,
	}
}

func (vm *Vm) disassemble(b1, b2 byte) {
	upper := upperBits(b1)
	fmt.Printf("PC: %x; Inst: %x\n", vm.Pc, upper)
	fmt.Println("==========")
}

func (vm *Vm) skipIf(cond bool) {
	if cond {
		vm.Pc += 2
	} else {
		vm.Pc++
	}
}

func (vm *Vm) setVF1If(cond bool) {
	vm.Regs[0xF] = 0
	if cond {
		vm.Regs[0xF] = 1
	}
}

func (vm *Vm) Run() error {
	byte1 := vm.Mem[vm.Pc]
	vm.Pc++
	byte2 := vm.Mem[vm.Pc]

	vm.disassemble(byte1, byte2)

	upper := upperBits(byte1)
	switch upper {
	case 0x0:
		switch byte2 {
		case 0xE0:
			// TODO: implement clearing screen
			fmt.Println("Clearing screen!")
			vm.Pc++
		case 0xEE:
			vm.Pc = vm.Stack[vm.Sp]
			vm.Sp--
		default:
			fmt.Println("Ignoring instruction")
			vm.Pc++
		}
	case 0x1:
		addr := (uint16(byte1&0x0F) << 8) | uint16(byte2)
		vm.Pc = addr
	case 0x2:
		vm.Sp++
		vm.Stack[vm.Sp] = vm.Pc
		vm.Pc = mergeBytePair(byte1&0x0F, byte2)
	case 0x3:
		x := byte1 & 0x0F
		vm.skipIf(vm.Regs[x] == byte2)
	case 0x4:
		x := byte1 & 0x0F
		vm.skipIf(vm.Regs[x] != byte2)
	case 0x5:
		x := byte1 & 0x0F
		y := (byte2 & 0xF0) >> 4
		vm.skipIf(vm.Regs[x] == vm.Regs[y])
	case 0x6:
		x := byte1 & 0x0F
		vm.Regs[x] = byte2
		vm.Pc++
	case 0x7:
		x := byte1 & 0x0F
		vm.Regs[x] = vm.Regs[x] + byte2
		// What about overflow?
		vm.Pc++
	case 0x8:
		x := byte1 & 0x0F
		y := (byte2 & 0xF0) >> 4
		z := byte2 & 0xF
		switch z {
		case 0x0:
			vm.Regs[x] = vm.Regs[y]
		case 0x1:
			vm.Regs[x] = vm.Regs[x] | vm.Regs[y]
		case 0x2:
			vm.Regs[x] = vm.Regs[x] & vm.Regs[y]
		case 0x3:
			vm.Regs[x] = vm.Regs[x] ^ vm.Regs[y]
		case 0x4:
			vm.setVF1If(vm.Regs[y] > 255-vm.Regs[x])
			vm.Regs[x] = vm.Regs[x] + vm.Regs[y]
		case 0x5:
			vm.setVF1If(vm.Regs[x] > vm.Regs[y])
			vm.Regs[x] = vm.Regs[x] - vm.Regs[y]
		case 0x6:
			bit := vm.Regs[x] & 1
			vm.setVF1If(bit == 1)
			vm.Regs[x] = vm.Regs[x] >> 1
		case 0x7:
			vm.setVF1If(vm.Regs[y] > vm.Regs[x])
			vm.Regs[x] = vm.Regs[y] - vm.Regs[x]
		case 0xE:
			bit := vm.Regs[x] & 0x80
			vm.setVF1If(bit == 0x80)
			vm.Regs[x] = vm.Regs[x] << 1
		default:
			return fmt.Errorf("Invalid instruction 0x8xyz; z: %x", z)
		}
		vm.Pc++
	default:
		return errors.New("Invalid instruction")
	}

	return nil
}
