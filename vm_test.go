package main

import "testing"

var runParams RunParams = RunParams{
	instCount:     1,
	frameDuration: 1,
}

type TestIO struct {
	drawCalled        bool
	clearScreenCalled bool
}

func (testIO *TestIO) Draw(pixels Pixels) {
	testIO.drawCalled = true
}

func (testIO *TestIO) ClearScreen() {
	testIO.clearScreenCalled = true
}

func (testIO *TestIO) WaitKeyPress() (byte, bool) {
	return 12, true
}

func (testIO *TestIO) GetKeysPressed() [16]bool {
	return [16]bool{true, true}
}

var testIO = &TestIO{}

func Test0nnn(t *testing.T) {
}

func Test00E0(t *testing.T) {
	ram := []byte{0x00, 0xE0}

	vm, _ := NewVm(ram, testIO)

	vm.Run(runParams)
	if !testIO.clearScreenCalled {
		t.Error("Clear screen instruction err; not called")
	}
}

func Test00EE(t *testing.T) {
	ram := []byte{0x00, 0xEE}

	vm, _ := NewVm(ram, testIO)
	vm.Stack[vm.Sp] = 0xABC
	vm.Sp = 1
	vm.Run(runParams)

	if vm.Pc != 0xABE || vm.Sp != 0 {
		t.Errorf("Return instruction err; Pc: %x, Sp: %x", vm.Pc, vm.Sp)
	}
}

func Test1nnn(t *testing.T) {
	ram := []byte{0x1A, 0xBC}

	vm, _ := NewVm(ram, testIO)
	vm.Run(runParams)

	if vm.Pc != 0xABC {
		t.Errorf("Jump instruction err; Pc: %x", vm.Pc)
	}
}

func Test2nnn(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0x20
	ram[1] = 0x04
	ram[4] = 0xFF

	vm, _ := NewVm(ram, testIO)
	vm.Run(runParams)

	if vm.Pc != 4 {
		t.Errorf("Call instruction err; Pc: %x", vm.Pc)
	}

	if vm.Stack[vm.Sp-1] != 0x200 || vm.Stack[vm.Sp] != 0 {
		t.Errorf("Call instruction err; Stack[Sp]: %x", vm.Stack[vm.Sp])
	}
}

func Test3xkk(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0x31
	ram[1] = 0xFF
	ram[4] = 0x32
	ram[5] = 0xAB

	vm, _ := NewVm(ram, testIO)
	vm.Regs[1] = 0xFF
	vm.Regs[2] = 0xCD
	vm.Run(runParams)

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run(runParams)

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

	vm, _ := NewVm(ram, testIO)
	vm.Regs[1] = 0xCD
	vm.Regs[2] = 0xFF
	vm.Run(runParams)

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run(runParams)

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

	vm, _ := NewVm(ram, testIO)
	vm.Regs[1] = 0xAA
	vm.Regs[2] = 0xAA
	vm.Regs[0xA] = 0x12
	vm.Regs[0xB] = 0x34

	vm.Run(runParams)
	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run(runParams)
	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run(runParams)
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func Test6xkk(t *testing.T) {
	ram := []byte{0x6A, 0xFF}
	vm, _ := NewVm(ram, testIO)
	vm.Run(runParams)

	if vm.Regs[0xA] != 0xFF {
		t.Errorf("Load instruction err; Reg value: %x", vm.Regs[0xA])
	}
}

func Test7xkk(t *testing.T) {
	ram := []byte{0x7A, 2}

	vm, _ := NewVm(ram, testIO)
	vm.Regs[0xA] = 3

	vm.Run(runParams)
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
		0x2A}

	vm, _ := NewVm(ram, testIO)
	vm.Regs[0x2] = 128

	vm.Run(runParams)
	if vm.Regs[0x1] != 128 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy0 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b0101
	vm.Regs[0x2] = 0b0010
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b111 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy1 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1101
	vm.Regs[0x2] = 0b1010
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1000 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy2 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x1] = 0b1101
	vm.Regs[0x2] = 0b1010
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b0111 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy3 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	err := vm.Run(runParams)
	if err == nil {
		t.Error("0x8xyA should return an error")
	}
}

func Test8xy4(t *testing.T) {
	ram := []byte{
		0x81,
		0x24,
		0x8E,
		0xF4,
		0x8F,
		0xE4,
		0x81,
		0x24,
		0x8E,
		0xF4,
		0x8F,
		0xE4,
	}

	vm, _ := NewVm(ram, testIO)

	vm.Regs[0x1] = 1
	vm.Regs[0x2] = 7
	vm.Run(runParams)
	if vm.Regs[0x1] != 8 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 1
	vm.Regs[0xF] = 7
	vm.Run(runParams)
	if vm.Regs[0xE] != 8 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xE] = 1
	vm.Regs[0xF] = 7
	vm.Run(runParams)
	if vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0xF])
	}

	vm.Regs[0x1] = 255
	vm.Regs[0x2] = 255
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 255
	vm.Regs[0xF] = 255
	vm.Run(runParams)
	if vm.Regs[0xE] != 0b1111_1110 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xE] = 255
	vm.Regs[0xF] = 255
	vm.Run(runParams)
	if vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy4 instruction err; Reg value: %x", vm.Regs[0xF])
	}
}

func Test8xy5(t *testing.T) {
	ram := []byte{
		0x81,
		0x25,
		0x8E,
		0xF5,
		0x8F,
		0xE5,
		0x81,
		0x25,
		0x8E,
		0xF5,
		0x8F,
		0xE5,
	}

	vm, _ := NewVm(ram, testIO)

	vm.Regs[0x1] = 5
	vm.Regs[0x2] = 7
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 5
	vm.Regs[0xF] = 7
	vm.Run(runParams)
	if vm.Regs[0xE] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xE] = 7
	vm.Regs[0xF] = 5
	vm.Run(runParams)
	if vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0xF])
	}

	vm.Regs[0x1] = 7
	vm.Regs[0x2] = 5
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 7
	vm.Regs[0xF] = 5
	vm.Run(runParams)
	if vm.Regs[0xE] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xF] = 7
	vm.Regs[0xE] = 5
	vm.Run(runParams)
	if vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy5 instruction err; Reg value: %x", vm.Regs[0xF])
	}
}

func Test8xy6(t *testing.T) {
	ram := []byte{
		0x81,
		0x26,
		0x8F,
		0x26,
		0x81,
		0x26,
		0x8F,
		0x26,
	}

	vm, _ := NewVm(ram, testIO)

	vm.Regs[0x2] = 0b1110
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b0111 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x2] = 0b1110
	vm.Run(runParams)
	if vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0xF])
	}

	vm.Regs[0x2] = 0b1111
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b0111 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x2] = 0b1111
	vm.Run(runParams)
	if vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy6 instruction err; Reg value: %x", vm.Regs[0xF])
	}
}

func Test8xy7(t *testing.T) {
	ram := []byte{
		0x81,
		0x27,
		0x8E,
		0xF7,
		0x8F,
		0xE7,
		0x81,
		0x27,
		0x8E,
		0xF7,
		0x8F,
		0xE7,
	}

	vm, _ := NewVm(ram, testIO)

	vm.Regs[0x1] = 7
	vm.Regs[0x2] = 5
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 7
	vm.Regs[0xF] = 5
	vm.Run(runParams)
	if vm.Regs[0xE] != 0b1111_1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xF] = 7
	vm.Regs[0xE] = 5
	vm.Run(runParams)
	if vm.Regs[0xF] != 0 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0xF])
	}

	vm.Regs[0x1] = 5
	vm.Regs[0x2] = 7
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0xE] = 5
	vm.Regs[0xF] = 7
	vm.Run(runParams)
	if vm.Regs[0xE] != 0b10 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0xE])
	}

	vm.Regs[0xF] = 5
	vm.Regs[0xE] = 7
	vm.Run(runParams)
	if vm.Regs[0xF] != 1 {
		t.Errorf("0x8xy7 instruction err; Reg value: %x", vm.Regs[0xF])
	}
}

func Test8xyE(t *testing.T) {
	ram := []byte{
		0x81,
		0x2E,
		0x8F,
		0x2E,
		0x81,
		0x2E,
		0x8F,
		0x2E,
	}

	vm, _ := NewVm(ram, testIO)

	vm.Regs[0x2] = 0b0111
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1110 || vm.Regs[0xF] != 0 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x2] = 0b0111
	vm.Run(runParams)
	if vm.Regs[0xF] != 0 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0xF])
	}

	vm.Regs[0x2] = 0b1111_1111
	vm.Run(runParams)
	if vm.Regs[0x1] != 0b1111_1110 || vm.Regs[0xF] != 1 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0x1])
	}

	vm.Regs[0x2] = 0b1111_1111
	vm.Run(runParams)
	if vm.Regs[0xF] != 1 {
		t.Errorf("0x8xyE instruction err; Reg value: %x", vm.Regs[0xF])
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

	vm, _ := NewVm(ram, testIO)
	vm.Regs[1] = 0x12
	vm.Regs[2] = 0x34
	vm.Regs[0xA] = 0xAA
	vm.Regs[0xB] = 0xAA

	vm.Run(runParams)
	if vm.Pc != 0x200+4 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	vm.Run(runParams)
	if vm.Pc != 0x200+6 {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}

	err := vm.Run(runParams)
	if err == nil {
		t.Errorf("Skip instruction err; Pc: %x", vm.Pc)
	}
}

func TestAnnn(t *testing.T) {
	ram := []byte{0xAA, 0xBC}

	vm, _ := NewVm(ram, testIO)
	vm.Run(runParams)

	if vm.I != 0xABC {
		t.Errorf("Load I instruction err; I: %x", vm.I)
	}
}

func TestBnnn(t *testing.T) {
	ram := []byte{0xBA, 0xBC}

	vm, _ := NewVm(ram, testIO)
	vm.Regs[0] = 0xFF
	vm.Run(runParams)

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
	ram[2] = 0xD3
	ram[3] = 0x43
	ram[4] = 0xD3
	ram[5] = 0x43
	ram[10] = 0xAB
	ram[11] = 0xCD
	ram[12] = 0xEF

	vm, _ := NewVm(ram, testIO)
	vm.I = 0x200 + 10
	vm.Regs[1] = 0
	vm.Regs[2] = 31

	vm.Run(runParams)
	if !testIO.drawCalled {
		t.Error("Draw function not called")
	}
	if vm.Regs[0xF] != 0 {
		t.Error("Draw instruction err; VF != 0")
	}

	row1 := [64]byte{1, 0, 1, 0, 1, 0, 1, 1}
	row2 := [64]byte{1, 1, 0, 0, 1, 1, 0, 1}
	row3 := [64]byte{1, 1, 1, 0, 1, 1, 1, 1}
	emptyRow := [64]byte{}
	if vm.Pixels[31] != row1 {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", row1, vm.Pixels[31])
	}
	// pixels should be clipped
	if vm.Pixels[0] != [64]byte{} {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", emptyRow, vm.Pixels[0])
	}
	if vm.Pixels[1] != [64]byte{} {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", emptyRow, vm.Pixels[1])
	}

	vm.Regs[3] = 64
	vm.Regs[4] = 32
	vm.Run(runParams)
	// pixels should wrap
	if vm.Pixels[0] != row1 {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", row1, vm.Pixels[0])
	}
	if vm.Pixels[1] != row2 {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", row2, vm.Pixels[1])
	}
	if vm.Pixels[2] != row3 {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", row3, vm.Pixels[2])
	}

	vm.Run(runParams)
	// Calling the same instruction again should result in the
	// erasure of the sprite
	if vm.Regs[0xF] != 1 {
		t.Error("Draw instruction err; VF != 1")
	}
	row1 = [64]byte{}
	if vm.Pixels[0] != emptyRow || vm.Pixels[1] != emptyRow || vm.Pixels[2] != emptyRow {
		t.Errorf("Draw instruction err; Expected: %v, Received: %v", emptyRow, vm.Pixels[31])
	}
}

func TestEx9EAndExA1(t *testing.T) {
	ram := make([]byte, 6)
	ram[0] = 0xEA
	ram[1] = 0x9E
	ram[4] = 0xEB
	ram[5] = 0xA1

	vm, _ := NewVm(ram, testIO)
	vm.Regs[0xA] = 0
	vm.Regs[0xB] = 1
	vm.Run(runParams)

	if vm.Pc != 0x200+4 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}

	vm.Run(runParams)

	if vm.Pc != 0x200+6 {
		t.Errorf("Skip on key instruction err; Pc: %x", vm.Pc)
	}
}

func TestFx0A(t *testing.T) {
	ram := []byte{0xF1, 0x0A}

	vm, _ := NewVm(ram, testIO)
	vm.Run(runParams)

	if vm.Regs[1] != 12 {
		t.Errorf("Load on key instruction err; V1: %x", vm.Regs[1])
	}
}

func TestFxInsts(t *testing.T) {
	ram := []byte{
		// 0xF1,
		// 0x07,
		// 0xF2,
		// 0x15,
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

	vm, _ := NewVm(ram, testIO)

	// TO DO: Some other test for these instructions
	// since the Run method decrements DT
	// vm.DT = 0xAA
	// vm.Run(runParams)
	// if vm.Regs[1] != 0xAA {
	// 	t.Errorf("Load Vx, DT err; V1: %x", vm.Regs[1])
	// }

	// vm.Regs[2] = 0xBB
	// vm.Run(runParams)
	// if vm.DT != 0xBB {
	// 	t.Errorf("Load DT, Vx err; DT: %x", vm.DT)
	// }

	vm.Regs[3] = 0xCC
	vm.Run(runParams)
	if vm.ST != 0xCC {
		t.Errorf("Load ST, Vx err; ST: %x", vm.ST)
	}

	vm.Regs[4] = 0xDD
	vm.I = 0x1
	vm.Run(runParams)
	if vm.I != 0xDE {
		t.Errorf("Add I, Vx err; I: %x", vm.I)
	}

	vm.Regs[0] = 0
	vm.Run(runParams)
	if vm.I != 0x00 {
		t.Errorf("Load F, Vx err, I: %x", vm.I)
	}

	vm.Regs[1] = 0xF
	vm.Run(runParams)
	if vm.I != 75 {
		t.Errorf("Load F, Vx err, I: %x", vm.I)
	}

	vm.Regs[7] = 123
	vm.Run(runParams)
	if vm.Mem[vm.I] != 1 || vm.Mem[vm.I+1] != 2 || vm.Mem[vm.I+2] != 3 {
		t.Errorf("Load B, Vx err, B: %d C: %d D: %d", vm.Mem[vm.I], vm.Mem[vm.I+1], vm.Mem[vm.I+2])
	}

	I := vm.I
	vm.Regs[0] = 0xA
	vm.Regs[1] = 0xB
	vm.Regs[2] = 0xC
	vm.Regs[3] = 0xD
	vm.Regs[4] = 0xE
	vm.Regs[5] = 0xF
	vm.Regs[6] = 0x1
	vm.Run(runParams)
	for i, v := range vm.Mem[I : I+7] {
		if vm.Regs[i] != v {
			t.Errorf("Load [I], Vx err, I: %x val: %x", vm.I+uint16(i), v)
		}
	}

	if vm.I != I+7 {
		t.Errorf("Load [I], Vx err, I: %x", vm.I)
	}

	I = vm.I
	for i := 0; i <= 7; i++ {
		vm.Mem[vm.I+uint16(i)] = byte(123 + i)
	}
	vm.Run(runParams)
	for i := 0; i <= 6; i++ {
		if vm.Regs[i] != byte(123+i) {
			t.Errorf("Load Vx, [I] err; Vx: %d, vm.Mem[I]: %d", vm.Regs[i], vm.Mem[vm.I+uint16(i)])
		}
	}

	if vm.I != I+7 {
		t.Errorf("Load Vx, [I] err, I: %x", vm.I)
	}
}
