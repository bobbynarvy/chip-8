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
	ClearScreen  func()
	Draw         func(x, y byte, bytes []byte) bool
	WaitKeyPress func() byte
}

func upperBits(b byte) byte {
	return (b & 0xF0) >> 4
}

func NewVm(rom []byte) (Vm, error) {
	// The first 0x1FF locations in RAM are reserved for
	// the CHIP-8 Interpreter.
	// The first 80 locations (16 chars x 5 bytes) in mem are used
	// to store the sprites representing the hex digits 0 to F.
	mem := Ram{
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

	// copy the ROM data into the RAM
	for j, v := range rom {
		if 0x200+j > 0xFFF {
			return Vm{}, errors.New("ROM size exceeds RAM limit")
		}
		mem[0x200+j] = v
	}

	return Vm{
		Mem: &mem,
		Pc:  0x200,
	}, nil
}

func (vm *Vm) disassemble(b1, b2 byte) {
	upper := upperBits(b1)
	fmt.Printf("PC: %x; Inst: %x\n", vm.Pc, upper)
	fmt.Println("==========")
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
	vm.incPc()

	vm.disassemble(byte1, byte2)

	upper := upperBits(byte1)
	switch upper {
	case 0x0:
		switch byte2 {
		case 0xE0:
			vm.ClearScreen()
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
		vm.Pc = (uint16(byte1&0x0F) << 8) | uint16(byte2)
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
		case 0x29:
			// The location of the hex digit sprites start at location 0 of vm.Mem;
			// vm.I is set with the location of the first byte of the sprite
			vm.I = uint16(vm.Regs[x]) * 5
		case 0x33:
			num := vm.Regs[x]
			vm.Mem[vm.I+2] = num % 10 // ones place
			num /= 10
			vm.Mem[vm.I+1] = num % 10 // tens place
			num /= 10
			vm.Mem[vm.I] = num % 10 // hundreds place
		case 0x55:
			for i := 0; i < int(x); i++ {
				vm.Mem[vm.I+uint16(i)] = vm.Regs[i]
			}
		case 0x65:
			for i := 0; i < int(x); i++ {
				vm.Regs[i] = vm.Mem[vm.I+uint16(i)]
			}
		default:
			return fmt.Errorf("Invalid instruction 0xFx%x", byte2)
		}
	default:
		return errors.New("Invalid instruction")
	}

	return nil
}
