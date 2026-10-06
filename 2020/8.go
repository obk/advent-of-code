// Day 8: Handheld Halting
//
// Each line of the input is an instruction: "acc N" adds N to the
// accumulator, "jmp N" jumps N instructions, "nop N" does nothing. The
// program loops forever.
//
// Part 1 is the accumulator just before any instruction runs twice.
// Part 2 changes exactly one jmp to nop or nop to jmp so that the program
// ends by running past its last instruction, and gives the accumulator.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 16
	maxProgram    = 10_000
)

// instruction is one line of the program.
type instruction struct {
	op  byte // 'a' acc, 'j' jmp, 'n' nop
	arg int
}

// readProgram parses the instructions.
func readProgram(path string) ([]instruction, error) {
	if path == "" {
		return nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxProgram {
		return nil, fmt.Errorf("%s: more than %d instructions", path, maxProgram)
	}
	program := make([]instruction, len(lines))
	for i, line := range lines {
		op, arg, found := strings.Cut(line, " ")
		n, err := strconv.Atoi(arg)
		if !found || err != nil || (op != "acc" && op != "jmp" && op != "nop") {
			return nil, fmt.Errorf("%s: malformed instruction %q", path, line)
		}
		program[i] = instruction{op[0], n}
	}
	return program, nil
}

// run executes the program with instruction flip swapped (-1 for none),
// stopping before any instruction runs twice. It reports the accumulator
// and whether the program ended normally.
func run(program []instruction, flip int, seen []bool) (acc int, ended bool, err error) {
	if len(seen) != len(program) || flip < -1 || flip >= len(program) {
		return 0, false, errors.New("wrong scratch size or flip out of range")
	}
	clear(seen)
	pc := 0
	for range len(program) + 1 {
		if pc == len(program) {
			return acc, true, nil
		}
		if pc < 0 || pc > len(program) {
			return 0, false, fmt.Errorf("jumped to %d, outside the program", pc)
		}
		if seen[pc] {
			return acc, false, nil
		}
		seen[pc] = true
		op := program[pc].op
		if pc == flip && op != 'a' {
			op ^= 'j' ^ 'n'
		}
		switch op {
		case 'a':
			acc, pc = acc+program[pc].arg, pc+1
		case 'j':
			pc += program[pc].arg
		default:
			pc++
		}
	}
	return 0, false, errors.New("the program ran longer than it has instructions")
}

func day8(path string) error {
	program, err := readProgram(path)
	if err != nil {
		return err
	}
	seen := make([]bool, len(program))
	loopAcc, ended, err := run(program, -1, seen)
	if err != nil {
		return err
	}
	if ended {
		return errors.New("the program ended without a change")
	}
	fixed := -1
	for i, in := range program {
		if in.op == 'a' {
			continue
		}
		acc, ended, err := run(program, i, seen)
		if err == nil && ended {
			if fixed >= 0 {
				return errors.New("more than one change ends the program")
			}
			fixed = acc
		}
	}
	if fixed < 0 {
		return errors.New("no single change ends the program")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", loopAcc, fixed); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
