package main

import (
	"errors"
	"fmt"
	"math/rand"
)

type Instruction struct {
	assembly string
	execFn   func(*Vm)
}

func newInst(assembly string, execFn func(*Vm)) Instruction {
	return Instruction{
		assembly: assembly,
		execFn:   execFn,
	}
}

var Sprintf = fmt.Sprintf

func getInstruction(byte1, byte2 byte) (Instruction, error) {
	addr := (uint16(byte1&0x0F) << 8) | uint16(byte2)
	x := byte1 & 0x0F
	y := (byte2 & 0xF0) >> 4
	z := byte2 & 0xF
	switch (byte1 & 0xF0) >> 4 {
	case 0x0:
		switch byte2 {
		case 0xE0:
			return newInst("CLS", func(vm *Vm) {
				vm.Pixels = Pixels{}
				vm.ClearScreen()
			}), nil
		case 0xEE:
			return newInst("RET", func(vm *Vm) {
				vm.Sp--
				vm.Pc = vm.Stack[vm.Sp]
				vm.incPc()
			}), nil
		default:
			return newInst("Ignored", func(vm *Vm) {
				fmt.Println("Ignoring instruction")
			}), nil
		}
	case 0x1:
		return newInst(Sprintf("%-4v %-3x", "JP", addr), func(vm *Vm) {
			// Check to see if the VM is jumping to the same address over and over.
			// If it is then the program is probably done.
			if vm.Pc-2 == addr {
				if vm.repeatCnt == 10 {
					vm.Done = true
				}
				vm.repeatCnt++
			}

			vm.Pc = addr
		}), nil
	case 0x2:
		return newInst(Sprintf("%-4v %-3x", "CALL", addr), func(vm *Vm) {
			vm.Stack[vm.Sp] = vm.Pc - 2 // at this point, vm.Pc will have been incremented
			vm.Sp++
			vm.Pc = addr
		}), nil
	case 0x3:
		return newInst(Sprintf("%-4v V%-2x %-3x", "SE", x, byte2), func(vm *Vm) {
			vm.skipIf(vm.Regs[x] == byte2)
		}), nil
	case 0x4:
		return newInst(Sprintf("%-4v V%-2x %-3x", "SNE", x, byte2), func(vm *Vm) {
			vm.skipIf(vm.Regs[x] != byte2)
		}), nil
	case 0x5:
		if z != 0 {
			return Instruction{}, fmt.Errorf("Invalid instruction 0x5xy%x", z)
		}
		return newInst(Sprintf("%-4v V%-2x V%-2x", "SE", x, y), func(vm *Vm) {
			vm.skipIf(vm.Regs[x] == vm.Regs[y])
		}), nil
	case 0x6:
		return newInst(Sprintf("%-4v V%-2x %-3x", "LD", x, byte2), func(vm *Vm) {
			vm.Regs[x] = byte2
		}), nil
	case 0x7:
		return newInst(Sprintf("%-4v V%-2x %-3x", "ADD", x, byte2), func(vm *Vm) {
			vm.Regs[x] = vm.Regs[x] + byte2
			// What about overflow?
		}), nil
	case 0x8:
		switch z {
		case 0x0:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "LD", x, y), func(vm *Vm) {
				vm.Regs[x] = vm.Regs[y]
			}), nil
		case 0x1:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "OR", x, y), func(vm *Vm) {
				vm.Regs[x] = vm.Regs[x] | vm.Regs[y]
			}), nil
		case 0x2:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "AND", x, y), func(vm *Vm) {
				vm.Regs[x] = vm.Regs[x] & vm.Regs[y]
			}), nil
		case 0x3:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "XOR", x, y), func(vm *Vm) {
				vm.Regs[x] = vm.Regs[x] ^ vm.Regs[y]
			}), nil
		case 0x4:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "ADD", x, y), func(vm *Vm) {
				vm.setVF1If(vm.Regs[y] > 255-vm.Regs[x])
				vm.Regs[x] = vm.Regs[x] + vm.Regs[y]
			}), nil
		case 0x5:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "SUB", x, y), func(vm *Vm) {
				vm.setVF1If(vm.Regs[x] > vm.Regs[y])
				vm.Regs[x] = vm.Regs[x] - vm.Regs[y]
			}), nil
		case 0x6:
			return newInst(Sprintf("%-4v V%-2x", "SHR", x), func(vm *Vm) {
				bit := vm.Regs[x] & 1
				vm.setVF1If(bit == 1)
				vm.Regs[x] = vm.Regs[x] >> 1
			}), nil
		case 0x7:
			return newInst(Sprintf("%-4v V%-2x V%-2x", "SUBN", x, y), func(vm *Vm) {
				vm.setVF1If(vm.Regs[y] > vm.Regs[x])
				vm.Regs[x] = vm.Regs[y] - vm.Regs[x]
			}), nil
		case 0xE:
			return newInst(Sprintf("%-4v V%-2x", "SHL", x), func(vm *Vm) {
				bit := vm.Regs[x] & 0x80
				vm.setVF1If(bit == 0x80)
				vm.Regs[x] = vm.Regs[x] << 1
			}), nil
		default:
			return Instruction{}, fmt.Errorf("Invalid instruction 0x8xyz; z: %x", z)
		}
	case 0x9:
		if z != 0 {
			return Instruction{}, fmt.Errorf("Invalid instruction 0x9xy%x", z)
		}
		return newInst(Sprintf("%-4v V%-2x V%-2x", "SNE", x, y), func(vm *Vm) {
			vm.skipIf(vm.Regs[x] != vm.Regs[y])
		}), nil
	case 0xA:
		return newInst(Sprintf("%-4v %-3v %-3x", "LD", "I", addr), func(vm *Vm) {
			vm.I = addr
		}), nil
	case 0xB:
		return newInst(Sprintf("%-4v %-3v %-3x", "JP", "V0", addr), func(vm *Vm) {
			vm.Pc = addr + uint16(vm.Regs[0])
		}), nil
	case 0xC:
		return newInst(Sprintf("%-4v V%-2x %-3x", "RND", x, byte2), func(vm *Vm) {
			randomByte := byte(rand.Intn(255))
			vm.Regs[x] = randomByte & byte2
		}), nil
	case 0xD:
		n := byte2 & 0xF
		return newInst(Sprintf("%-4v V%-2x V%-2x %-3x", "DRW", x, y, n), func(vm *Vm) {
			spriteGroup := vm.Mem[vm.I : vm.I+uint16(n)]
			vm.Regs[0xF] = 0
			for yOffset, sprite := range spriteGroup {
				for xOffset := 0; xOffset < 8; xOffset++ {
					col := (int(vm.Regs[x]) + xOffset) % 64 // mod is for wrapping
					row := (int(vm.Regs[y]) + yOffset) % 32
					pixel := &vm.Pixels[row][col]

					// check if the current bit in the sprite is to be drawn
					if sprite&0x80 > 1 {
						// check if pixel has already been drawn on
						if *pixel == 1 {
							*pixel = 0
							vm.Regs[0xF] = 1
						} else {
							*pixel = 1
						}
					}
					sprite <<= 1
				}
			}
			vm.Draw(vm.Pixels)
		}), nil
	case 0xE:
		switch byte2 {
		case 0x9E:
			return newInst(Sprintf("%-4v V%-2x", "SKP", x), func(vm *Vm) {
				vm.Keys = vm.GetKeysPressed()
				vm.skipIf(vm.Keys[vm.Regs[x]])
			}), nil
		case 0xA1:
			return newInst(Sprintf("%-4v V%-2x", "SKNP", x), func(vm *Vm) {
				vm.Keys = vm.GetKeysPressed()
				vm.skipIf(!vm.Keys[vm.Regs[x]])
			}), nil
		default:
			return Instruction{}, fmt.Errorf("Invalid instruction 0xEx%x", byte2)
		}
	case 0xF:
		switch byte2 {
		case 0x07:
			return newInst(Sprintf("%-4v V%-2x %-3v", "LD", x, "DT"), func(vm *Vm) {
				vm.Regs[x] = vm.DT
			}), nil
		case 0x0A:
			return newInst(Sprintf("%-4v V%-2x %-3v", "LD", x, "K"), func(vm *Vm) {
				vm.Regs[x] = vm.WaitKeyPress()
			}), nil
		case 0x15:
			return newInst(Sprintf("%-4v %-3v V%-2x", "LD", "DT", x), func(vm *Vm) {
				vm.DT = vm.Regs[x]
			}), nil
		case 0x18:
			return newInst(Sprintf("%-4v %-3v V%-2x", "LD", "ST", x), func(vm *Vm) {
				vm.ST = vm.Regs[x]
			}), nil
		case 0x1E:
			return newInst(Sprintf("%-4v %-3v V%-2x", "ADD", "I", x), func(vm *Vm) {
				vm.I = vm.I + uint16(vm.Regs[x])
			}), nil
		case 0x29:
			// The location of the hex digit sprites start at location 0 of vm.Mem;
			// vm.I is set with the location of the first byte of the sprite
			return newInst(Sprintf("%-4v %-3v V%-2x", "LD", "F", x), func(vm *Vm) {
				vm.I = uint16(vm.Regs[x]) * 5
			}), nil
		case 0x33:
			return newInst(Sprintf("%-4v %-3v V%-2x", "LD", "B", x), func(vm *Vm) {
				num := vm.Regs[x]
				vm.Mem[vm.I+2] = num % 10 // ones place
				num /= 10
				vm.Mem[vm.I+1] = num % 10 // tens place
				num /= 10
				vm.Mem[vm.I] = num % 10 // hundreds place
			}), nil
		case 0x55:
			return newInst(Sprintf("%-4v %-3v V%-2x", "LD", "[I]", x), func(vm *Vm) {
				for i := 0; i <= int(x); i++ {
					vm.Mem[vm.I+uint16(i)] = vm.Regs[i]
				}
			}), nil
		case 0x65:
			return newInst(Sprintf("%-4v V%-2x %-3v", "LD", x, "[I]"), func(vm *Vm) {
				for i := 0; i <= int(x); i++ {
					vm.Regs[i] = vm.Mem[vm.I+uint16(i)]
				}
			}), nil
		default:
			return Instruction{}, fmt.Errorf("Invalid instruction 0xFx%x", byte2)
		}
	default:
		return Instruction{}, errors.New("Invalid instruction")
	}
}
