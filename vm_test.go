package main

import "testing"

func Test0nnn(t *testing.T) {
	ram := []byte{0x01}

	vm, _ := NewVm(ram)
	vm.Run()

	if vm.Pc != 0x200+2 {
		t.Errorf("Ignore instruction err; Pc: %x", vm.Pc)
	}
}

func Test00E0(t *testing.T) {
	ram := []byte{0x00, 0xE0}

	vm, _ := NewVm(ram)
	var called bool
	vm.ClearScreen = func() {
		called = true
	}

	vm.Run()
	if !called {
		t.Error("Clear screen instruction err; not called")
	}
}

func Test00EE(t *testing.T) {
	ram := []byte{0x00, 0xEE}

	vm, _ := NewVm(ram)
	vm.Sp = 1
	vm.Stack[vm.Sp] = 0xABC
	vm.Run()

	if vm.Pc != 0xABC || vm.Sp != 0 {
		t.Errorf("Return instruction err; Pc: %x, Sp: %x", vm.Pc, vm.Sp)
	}
}

func Test1nnn(t *testing.T) {
	ram := []byte{0x1A, 0xBC}

	vm, _ := NewVm(ram)
	vm.Run()

	if vm.Pc != 0xABC {
		t.Errorf("Jump instruction err; Pc: %x", vm.Pc)
	}
}

func Test2nnn(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0x20
	ram[1] = 0x04
	ram[4] = 0xFF

	vm, _ := NewVm(ram)
	vm.Run()

	if vm.Pc != 4 {
		t.Errorf("Call instruction err; Pc: %x", vm.Pc)
	}

	if vm.Stack[vm.Sp] != 0x200+2 {
		t.Errorf("Call instruction err; Stack[Sp]: %x", vm.Stack[vm.Sp])
	}
}

func Test3xkk(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0x31
	ram[1] = 0xFF
	ram[4] = 0x32
	ram[5] = 0xAB

	vm, _ := NewVm(ram)
	vm.Regs[1] = 0xFF
	vm.Regs[2] = 0xCD
	vm.Run()

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test4xkk(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0x41
	ram[1] = 0xAB
	ram[4] = 0x42
	ram[5] = 0xFF

	vm, _ := NewVm(ram)
	vm.Regs[1] = 0xCD
	vm.Regs[2] = 0xFF
	vm.Run()

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test5xy0(t *testing.T) {
	ram := make([]byte, 8)
	ram[0] = 0x51
	ram[1] = 0x20
	ram[4] = 0x5A
	ram[5] = 0xB0
	ram[6] = 0x5C
	ram[7] = 0xC1

	vm, _ := NewVm(ram)
	vm.Regs[1] = 0xAA
	vm.Regs[2] = 0xAA
	vm.Regs[0xA] = 0x12
	vm.Regs[0xB] = 0x34

	vm.Run()
	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()
	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run()
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test6xkk(t *testing.T) {
	ram := []byte{0x6A, 0xFF}
	vm, _ := NewVm(ram)
	vm.Run()

	if vm.Regs[0xA] != 0xFF {
		t.Errorf("Load instruction err; Reg value: %x", vm.Regs[0xA])
	}
}

func Test7xkk(t *testing.T) {
	ram := []byte{0x7A, 2}

	vm, _ := NewVm(ram)
	vm.Regs[0xA] = 3

	vm.Run()
	if vm.Regs[0xA] != 5 {
		t.Errorf("Add instruction err; Reg value: %x", vm.Regs[0xA])
	}
}

func Test8xyz(t *testing.T) {
	ram := []byte{
		0x81,
		0x20,
		0x81,
		0x21,
		0x81,
		0x22,
		0x81,
		0x23,
		0x81,
		0x24,
		0x81,
		0x24,
		0x81,
		0x25,
		0x81,
		0x25,
		0x81,
		0x26,
		0x81,
		0x26,
		0x81,
		0x27,
		0x81,
		0x27,
		0x81,
		0x2E,
		0x81,
		0x2E,
		0x81,
		0x2A}

	vm, _ := NewVm(ram)
	vm.Regs[0x2] = 128

	vm.Run()
	if vm.Regs[0x1] != 128 {
		t.Errorf("0x8xy0 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b0101
	vm.Regs[0x2] = 0b0010
	vm.Run()
	if vm.Regs[0x1] != 0b111 {
		t.Errorf("0x8xy1 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1101
	vm.Regs[0x2] = 0b1010
	vm.Run()
	if vm.Regs[0x1] != 0b1000 {
		t.Errorf("0x8xy2 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1101
	vm.Regs[0x2] = 0b1010
	vm.Run()
	if vm.Regs[0x1] != 0b0111 {
		t.Errorf("0x8xy3 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 1
	vm.Regs[0x2] = 7
	vm.Run()
	if vm.Regs[0x1] != 8 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 255
	vm.Regs[0x2] = 255
	vm.Run()
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 5
	vm.Regs[0x2] = 7
	vm.Run()
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 7
	vm.Regs[0x2] = 5
	vm.Run()
	if vm.Regs[0x1] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1110
	vm.Run()
	if vm.Regs[0x1] != 0b0111 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1111
	vm.Run()
	if vm.Regs[0x1] != 0b0111 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 7
	vm.Regs[0x2] = 5
	vm.Run()
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 5
	vm.Regs[0x2] = 7
	vm.Run()
	if vm.Regs[0x1] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b0111
	vm.Run()
	if vm.Regs[0x1] != 0b1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1111_1111
	vm.Run()
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0x1])
	}

	err := vm.Run()
	if err == nil {
		t.Error("0x8xyA should return an error")
	}
}

func Test9xy0(t *testing.T) {
	ram := make([]byte, 8)
	ram[0] = 0x91
	ram[1] = 0x20
	ram[4] = 0x9A
	ram[5] = 0xB0
	ram[6] = 0x9C
	ram[7] = 0xC1

	vm, _ := NewVm(ram)
	vm.Regs[1] = 0x12
	vm.Regs[2] = 0x34
	vm.Regs[0xA] = 0xAA
	vm.Regs[0xB] = 0xAA

	vm.Run()
	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()
	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run()
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func TestAnnn(t *testing.T) {
	ram := []byte{0xAA, 0xBC}

	vm, _ := NewVm(ram)
	vm.Run()

	if vm.I != 0xABC {
		t.Errorf("Load I instruction err; I: %x", vm.I)
	}
}

func TestBnnn(t *testing.T) {
	ram := []byte{0xBA, 0xBC}

	vm, _ := NewVm(ram)
	vm.Regs[0] = 0xFF
	vm.Run()

	if vm.Pc != 0xBBB {
		t.Errorf("Jump V0, addr instruction err; Pc: %x", vm.Pc)
	}
}

func TestCnnn(t *testing.T) {
	t.Skip()
}

func TestDxyn(t *testing.T) {
	ram := make([]byte, 14)
	ram[0] = 0xD1
	ram[1] = 0x23
	ram[10] = 0xAB
	ram[11] = 0xCD
	ram[12] = 0xEF

	var called bool
	vm, _ := NewVm(ram)
	vm.I = 0x200 + 10
	vm.Regs[1] = 12
	vm.Regs[2] = 34
	vm.Draw = func(x, y byte, bytes []byte) bool {
		called = true
		if x != 12 || y != 34 {
			t.Errorf("Draw instruction err; x: %d, y: %d", x, y)
		}
		for i, v := range bytes {
			if v != ram[10+i] {
				t.Errorf("Draw instruction err; b1: %d, b2: %d", v, ram[10+i])
			}
		}
		return true
	}

	vm.Run()
	if !called {
		t.Error("Draw function not called")
	}
	if vm.Regs[0xF] != 1 {
		t.Errorf("Draw instruction err; VF: %x", vm.Regs[0xF])
	}
}

func TestEx9EAndExA1(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0xE1
	ram[1] = 0x9E
	ram[4] = 0xE2
	ram[5] = 0xA1

	vm, _ := NewVm(ram)
	vm.Keys[1] = true
	vm.Keys[2] = true
	vm.Run()

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 0x200+6 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}
}

func TestFx0A(t *testing.T) {
	ram := []byte{0xF1, 0x0A}

	vm, _ := NewVm(ram)
	vm.WaitKeyPress = func() byte {
		return 12
	}
	vm.Run()

	if vm.Regs[1] != 12 {
		t.Errorf("Load on key instruction err; V1: %x", vm.Regs[1])
	}
}

func TestFxInsts(t *testing.T) {
	ram := []byte{
		0xF1,
		0x07,
		0xF2,
		0x15,
		0xF3,
		0x18,
		0xF4,
		0x1E,
		0xF0,
		0x29,
		0xF1,
		0x29,
		0xF7,
		0x33,
		0xF6,
		0x55,
		0xF6,
		0x65,
	}

	vm, _ := NewVm(ram)

	vm.DT = 0xAA
	vm.Run()
	if vm.Regs[1] != 0xAA {
		t.Errorf("Load Vx, DT err; V1: %x", vm.Regs[1])
	}

	vm.Regs[2] = 0xBB
	vm.Run()
	if vm.DT != 0xBB {
		t.Errorf("Load DT, Vx err; DT: %x", vm.DT)
	}

	vm.Regs[3] = 0xCC
	vm.Run()
	if vm.ST != 0xCC {
		t.Errorf("Load ST, Vx err; ST: %x", vm.ST)
	}

	vm.Regs[4] = 0xDD
	vm.I = 0x1
	vm.Run()
	if vm.I != 0xDE {
		t.Errorf("Add I, Vx err; I: %x", vm.I)
	}

	vm.Regs[0] = 0
	vm.Run()
	if vm.I != 0x00 {
		t.Errorf("Load F, Vx err, I: %x", vm.I)
	}

	vm.Regs[1] = 0xF
	vm.Run()
	if vm.I != 75 {
		t.Errorf("Load F, Vx err, I: %x", vm.I)
	}

	vm.Regs[7] = 123
	vm.Run()
	if vm.Mem[vm.I] != 1 || vm.Mem[vm.I+1] != 2 || vm.Mem[vm.I+2] != 3 {
		t.Errorf("Load B, Vx err, B: %d C: %d D: %d", vm.Mem[vm.I], vm.Mem[vm.I+1], vm.Mem[vm.I+2])
	}

	vm.Regs[0] = 0xA
	vm.Regs[1] = 0xB
	vm.Regs[2] = 0xC
	vm.Regs[3] = 0xD
	vm.Regs[4] = 0xE
	vm.Regs[5] = 0xF
	vm.Regs[6] = 0x1
	vm.Run()
	for i, v := range vm.Mem[vm.I : vm.I+6] {
		if vm.Regs[i] != v {
			t.Errorf("Load [I], Vx err, I: %x val: %x", vm.I+uint16(i), v)
		}
	}

	for i := 0; i <= 6; i++ {
		vm.Mem[vm.I+uint16(i)] = byte(123 + i)
	}
	vm.Run()
	for i := 0; i <= 6; i++ {
		if vm.Regs[i] != byte(123+i) {
			t.Errorf("Load Vx, [I] err; Vx: %d, vm.Mem[I]: %d", vm.Regs[i], vm.Mem[vm.I+uint16(i)])
		}
	}
}
