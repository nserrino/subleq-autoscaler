package main

import (
	"testing"
)

func TestParseProgram(t *testing.T) {
	program := CreateSubleqProgram("10x-1000x20")
	expected := []int{10, -1000, 20}
	for i := 0; i < len(expected); i++ {
		if program.Instructions[i] != expected[i] {
			t.Errorf("Instructions[%d] should be %d, got %d", i, expected[i], program.Instructions[i])
		}
	}
}

func TestParseProgram2(t *testing.T) {
	program := CreateSubleqProgram("9x-1x3x10x-1x6x0x0x-1x72x105x0")
	expected := []int{
		9, -1, 3,
		10, -1, 6,
		0, 0, -1,
		72, 105, 0,
	}
	for i := 0; i < len(expected); i++ {
		if program.Instructions[i] != expected[i] {
			t.Errorf("Instructions[%d] should be %d, got %d", i, expected[i], program.Instructions[i])
		}
	}
}

func TestParseProgramBadInput(t *testing.T) {
	program := CreateSubleqProgram("fooxbarx123")
	expected := []int{
		9, -1, 3,
		10, -1, 6,
		0, 0, -1,
		72, 105, 0,
	}
	for i := 0; i < len(expected); i++ {
		if program.Instructions[i] != expected[i] {
			t.Errorf("Instructions[%d] should be %d, got %d", i, expected[i], program.Instructions[i])
		}
	}
}

func TestHiProgram(t *testing.T) {
	subleq := &SubleqProgram{
		Step:               0,
		InstructionPointer: 0,
		Instructions: []int{
			9, -1, 3,
			10, -1, 6,
			0, 0, -1,
			72, 105, 0,
		},
	}
	expected := []int{72, 105, 0, -1}
	for i := 0; i < len(expected); i++ {
		res := subleq.GetNextOutputValue()
		if res != expected[i] {
			t.Errorf("result[%d] should be %d, got %d", i, expected[i], res)
		}
	}
}

func TestHelloWorldProgram(t *testing.T) {
	subleq := &SubleqProgram{
		Step:               0,
		InstructionPointer: 0,
		Instructions: []int{
			12, 12, 3,
			36, 37, 6,
			37, 12, 9,
			37, 37, 12,
			0, -1, 15,
			38, 36, 18,
			12, 12, 21,
			53, 37, 24,
			37, 12, 27,
			37, 37, 30,
			36, 12, -1,
			37, 37, 0,
			39, 0, -1,
			72, 101, 108,
			108, 111, 44,
			32, 87, 111,
			114, 108, 100,
			33, 10, 53,
		},
	}
	expected := []int{
		0, 0, 0, 0, 72,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 101,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 108,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 108,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 111,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 44,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 87,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 111,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 114,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 108,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 33,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10,
		0, 0, 0, 0, 0, 0, -1,
	}
	for i := 0; i < len(expected); i++ {
		res := subleq.GetNextOutputValue()
		if res != expected[i] {
			t.Errorf("result[%d] should be %d, got %d", i, expected[i], res)
		}
	}
}
