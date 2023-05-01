package main

import "testing"

func Test0nnn(t *testing.T) {
	var mem Ram
	mem[0] = 0x01

	vm := NewVm(&mem)
	vm.Run()

	if vm.Pc != 2 {
		t.Errorf("Ignore instruction err; Pc: %x", vm.Pc)
	}
}

func Test00E0(t *testing.T) {
	t.Skip()
}

func Test00EE(t *testing.T) {
	var mem Ram
	mem[0] = 0x00
	mem[1] = 0xEE

	vm := NewVm(&mem)
	vm.Sp = 1
	vm.Stack[vm.Sp] = 0xABC
	vm.Run()

	if vm.Pc != 0xABC || vm.Sp != 0 {
		t.Errorf("Return instruction err; Pc: %x, Sp: %x", vm.Pc, vm.Sp)
	}
}

func Test1nnn(t *testing.T) {
	var mem Ram
	mem[0] = 0x1A
	mem[1] = 0xBC

	vm := NewVm(&mem)
	vm.Run()

	if vm.Pc != 0xABC {
		t.Errorf("Jump instruction err; Pc: %x", vm.Pc)
	}
}

func Test2nnn(t *testing.T) {
	var mem Ram
	mem[0] = 0x20
	mem[1] = 0x04
	mem[4] = 0xFF

	vm := NewVm(&mem)
	vm.Run()

	if vm.Pc != 0x04 {
		t.Errorf("Call instruction err; Pc: %x", vm.Pc)
	}

	if vm.Stack[vm.Sp] != 2 {
		t.Errorf("Call instruction err; Stack[Sp]: %x", vm.Stack[vm.Sp])
	}
}

func Test3xkk(t *testing.T) {
	var mem Ram
	mem[0] = 0x31
	mem[1] = 0xFF
	mem[4] = 0x32
	mem[5] = 0xAB

	vm := NewVm(&mem)
	vm.Regs[1] = 0xFF
	vm.Regs[2] = 0xCD
	vm.Run()

	if vm.Pc != 4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test4xkk(t *testing.T) {
	var mem Ram
	mem[0] = 0x41
	mem[1] = 0xAB
	mem[4] = 0x42
	mem[5] = 0xFF

	vm := NewVm(&mem)
	vm.Regs[1] = 0xCD
	vm.Regs[2] = 0xFF
	vm.Run()

	if vm.Pc != 4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test5xy0(t *testing.T) {
	var mem Ram
	mem[0] = 0x51
	mem[1] = 0x20
	mem[4] = 0x5A
	mem[5] = 0xB0
	mem[6] = 0x5C
	mem[7] = 0xC1

	vm := NewVm(&mem)
	vm.Regs[1] = 0xAA
	vm.Regs[2] = 0xAA
	vm.Regs[0xA] = 0x12
	vm.Regs[0xB] = 0x34

	vm.Run()
	if vm.Pc != 4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()
	if vm.Pc != 6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run()
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test6xkk(t *testing.T) {
	var mem Ram
	mem[0] = 0x6A
	mem[1] = 0xFF

	vm := NewVm(&mem)
	vm.Run()

	if vm.Regs[0xA] != 0xFF {
		t.Errorf("Load instruction err; Reg value: %x", vm.Regs[0xA])
	}
}

func Test7xkk(t *testing.T) {
	var mem Ram
	mem[0] = 0x7A
	mem[1] = 2

	vm := NewVm(&mem)
	vm.Regs[0xA] = 3

	vm.Run()
	if vm.Regs[0xA] != 5 {
		t.Errorf("Add instruction err; Reg value: %x", vm.Regs[0xA])
	}
}

func Test8xyz(t *testing.T) {
	var mem Ram
	mem[0] = 0x81
	mem[1] = 0x20
	mem[2] = 0x81
	mem[3] = 0x21
	mem[4] = 0x81
	mem[5] = 0x22
	mem[6] = 0x81
	mem[7] = 0x23
	mem[8] = 0x81
	mem[9] = 0x24
	mem[10] = 0x81
	mem[11] = 0x24
	mem[12] = 0x81
	mem[13] = 0x25
	mem[14] = 0x81
	mem[15] = 0x25
	mem[16] = 0x81
	mem[17] = 0x26
	mem[18] = 0x81
	mem[19] = 0x26
	mem[20] = 0x81
	mem[21] = 0x27
	mem[22] = 0x81
	mem[23] = 0x27
	mem[24] = 0x81
	mem[25] = 0x2E
	mem[26] = 0x81
	mem[27] = 0x2E
	mem[28] = 0x81
	mem[29] = 0x2A

	vm := NewVm(&mem)
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
	var mem Ram
	mem[0] = 0x91
	mem[1] = 0x20
	mem[4] = 0x9A
	mem[5] = 0xB0
	mem[6] = 0x9C
	mem[7] = 0xC1

	vm := NewVm(&mem)
	vm.Regs[1] = 0x12
	vm.Regs[2] = 0x34
	vm.Regs[0xA] = 0xAA
	vm.Regs[0xB] = 0xAA

	vm.Run()
	if vm.Pc != 4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()
	if vm.Pc != 6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run()
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func TestAnnn(t *testing.T) {
	var mem Ram
	mem[0] = 0xAA
	mem[1] = 0xBC

	vm := NewVm(&mem)
	vm.Run()

	if vm.I != 0xABC {
		t.Errorf("Load I instruction err; I: %x", vm.I)
	}
}

func TestBnnn(t *testing.T) {
	var mem Ram
	mem[0] = 0xBA
	mem[1] = 0xBC

	vm := NewVm(&mem)
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
	var mem Ram
	mem[0] = 0xD1
	mem[1] = 0x23
	mem[10] = 0xAB
	mem[11] = 0xCD
	mem[12] = 0xEF

	var called bool
	vm := NewVm(&mem)
	vm.I = 10
	vm.Regs[1] = 12
	vm.Regs[2] = 34
	vm.Draw = func(x, y byte, bytes []byte) bool {
		called = true
		if x != 12 || y != 34 {
			t.Errorf("Draw instruction err; x: %d, y: %d", x, y)
		}
		for i, v := range bytes {
			if v != mem[10+i] {
				t.Errorf("Draw instruction err; b1: %d, b2: %d", v, mem[10+i])
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
	var mem Ram
	mem[0] = 0xE1
	mem[1] = 0x9E
	mem[4] = 0xE2
	mem[5] = 0xA1

	vm := NewVm(&mem)
	vm.Keys[1] = true
	vm.Keys[2] = true
	vm.Run()

	if vm.Pc != 4 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}

	vm.Run()

	if vm.Pc != 6 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}
}

func TestFx0A(t *testing.T) {
	var mem Ram
	mem[0] = 0xF1
	mem[1] = 0x0A

	vm := NewVm(&mem)
	vm.WaitKeyPress = func() byte {
		return 12
	}
	vm.Run()

	if vm.Regs[1] != 12 {
		t.Errorf("Load on key instruction err; V1: %x", vm.Regs[1])
	}
}

func TestFxInsts(t *testing.T) {
	var mem Ram
	mem[0] = 0xF1
	mem[1] = 0x07
	mem[2] = 0xF2
	mem[3] = 0x15
	mem[4] = 0xF3
	mem[5] = 0x18
	mem[6] = 0xF4
	mem[7] = 0x1E

	vm := NewVm(&mem)

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
}
