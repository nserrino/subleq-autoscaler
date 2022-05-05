package main

import (
	"strconv"
	"strings"
)

// SubleqProgram represents a program made of `subleq` instructions.
type SubleqProgram struct {
	// The current instruction index to be executed
	InstructionPointer int
	// The set of instructions for the program
	Instructions []int
	// Number of instruction executions so far. This is different from InstructionPointer,
	// since the same instruction may be re-executed.
	Step int
	// The last value outputted by the program
	LastValue int
}

// CreateSubleqProgram creates a new Subleq program from an input string.
func CreateSubleqProgram(input string) *SubleqProgram {
	// Use this if the input string doesn't match the expected format.
	defaultProgram := &SubleqProgram{
		Step:               0,
		InstructionPointer: 0,
		Instructions: []int{
			9, -1, 3,
			10, -1, 6,
			0, 0, -1,
			72, 105, 0,
		},
	}

	split := strings.Split(input, "x")
	if len(split) < 3 {
		return defaultProgram
	}

	var ints []int
	for _, item := range split {
		i, err := strconv.Atoi(item)
		if err != nil {
			return defaultProgram
		}
		ints = append(ints, i)
	}

	return &SubleqProgram{
		Step:               0,
		InstructionPointer: 0,
		Instructions:       ints,
	}
}

// Get the value at memory address i, returning 0 for memory addresses that are
// out of bounds.
func (s *SubleqProgram) getValueAtAddress(i int) int {
	// For out of bounds memory accesses, output 0
	if i < 0 || i >= len(s.Instructions) {
		return 0
	}
	return s.Instructions[i]
}

// Set the value at memory address i, returning for memory addresses that are
// out of bounds.
func (s *SubleqProgram) setValueAtAddress(i int, val int) {
	// For out of bounds memory accesses, output 0
	if i < 0 || i >= len(s.Instructions) {
		return
	}
	s.Instructions[i] = val
}

// Invoke the program and get the target number of output pods to produce.
// Corresponds to evaluating a single `subleq` instruction.
// -1 means program is done
// 0 is the 0 value for an executing program
func (s *SubleqProgram) GetNextOutputValue() int {
	// If execution is complete, return 0
	if s.InstructionPointer < 0 {
		s.LastValue = -1
		return s.LastValue
	}

	a := s.Instructions[s.InstructionPointer+0]
	b := s.Instructions[s.InstructionPointer+1]
	c := s.Instructions[s.InstructionPointer+2]

	registerA := s.getValueAtAddress(a)
	registerB := s.getValueAtAddress(b)
	s.setValueAtAddress(b, registerB-registerA)

	// Increment the step
	s.Step += 1

	if registerB-registerA > 0 {
		// Increment the instruction pointer by 3 (one full instruction) wyeshen B - A > 0
		s.InstructionPointer += 3
	} else {
		// Jump to the value at C when B - A <= 0
		s.InstructionPointer = c
	}

	// Return 0 if registerA is not -1, otherwise return register A.
	if b == -1 {
		s.LastValue = registerA
	} else {
		s.LastValue = 0
	}
	return s.LastValue
}
