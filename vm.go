package main

import (
	"errors"
	"fmt"
	"math/rand"
)

type Ram [0xFFF]byte

type Vm struct {
	Mem          *Ram
	Stack        [16]uint16
	Regs         [16]byte
	I            uint16   // register used mostly to store memory addresses
	DT           byte     // delay timer
	ST           byte     // sound timer
	Pc           uint16   // program counter
	Sp           byte     // stack pointer
	Keys         [16]bool // represents the 16-key keypad; a true value means the key corresponding key is pressed
	Draw         func(x, y byte, bytes []byte) bool
	WaitKeyPress func() byte
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

// Increment the program counter
func (vm *Vm) incPc() {
	// Since instructions are 2 bytes long; the next
	// instruction shouldn't be the one after the first
	// byte but the one after that
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
	vm.incPc()

	vm.disassemble(byte1, byte2)

	upper := upperBits(byte1)
	switch upper {
	case 0x0:
		switch byte2 {
		case 0xE0:
			// TODO: implement clearing screen
			fmt.Println("Clearing screen!")
		case 0xEE:
			vm.Pc = vm.Stack[vm.Sp]
			vm.Sp--
		default:
			fmt.Println("Ignoring instruction")
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
		z := byte2 & 0xF
		if z != 0 {
			return fmt.Errorf("Invalid instruction 0x5xy%x", z)
		}
		vm.skipIf(vm.Regs[x] == vm.Regs[y])
	case 0x6:
		x := byte1 & 0x0F
		vm.Regs[x] = byte2
	case 0x7:
		x := byte1 & 0x0F
		vm.Regs[x] = vm.Regs[x] + byte2
		// What about overflow?
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
	case 0x9:
		x := byte1 & 0x0F
		y := (byte2 & 0xF0) >> 4
		z := byte2 & 0xF
		if z != 0 {
			return fmt.Errorf("Invalid instruction 0x9xy%x", z)
		}
		vm.skipIf(vm.Regs[x] != vm.Regs[y])
	case 0xA:
		addr := (uint16(byte1&0x0F) << 8) | uint16(byte2)
		vm.I = addr
	case 0xB:
		addr := (uint16(byte1&0x0F) << 8) | uint16(byte2)
		vm.Pc = addr + uint16(vm.Regs[0])
	case 0xC:
		x := byte1 & 0x0F
		randomByte := byte(rand.Intn(255))
		vm.Regs[x] = randomByte & byte2
	case 0xD:
		x := byte1 & 0x0F
		y := (byte2 & 0xF0) >> 4
		n := byte2 & 0xF
		bytes := vm.Mem[vm.I : vm.I+uint16(n)]
		collision := vm.Draw(vm.Regs[x], vm.Regs[y], bytes)
		vm.Regs[0xF] = 0
		if collision {
			vm.Regs[0xF] = 1
		}
	case 0xE:
		x := byte1 & 0x0F
		switch byte2 {
		case 0x9E:
			vm.skipIf(vm.Keys[x])
		case 0xA1:
			vm.skipIf(!vm.Keys[x])
		default:
			return fmt.Errorf("Invalid instruction 0xEx%x", byte2)
		}
	case 0xF:
		x := byte1 & 0x0F
		switch byte2 {
		case 0x07:
			vm.Regs[x] = vm.DT
		case 0x0A:
			vm.Regs[x] = vm.WaitKeyPress()
		case 0x15:
			vm.DT = vm.Regs[x]
		case 0x18:
			vm.ST = vm.Regs[x]
		case 0x1E:
			vm.I = vm.I + uint16(vm.Regs[x])
		default:
			return fmt.Errorf("Invalid instruction 0xFx%x", byte2)
		}
	default:
		return errors.New("Invalid instruction")
	}

	return nil
}
