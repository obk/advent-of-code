// Day 23: Opening the Turing Lock
//
// The input is a program for a computer with two registers, a and b, and six
// instructions: hlf, tpl and inc change a register, jmp jumps by an offset,
// jie jumps if a register is even and jio if it is exactly one. The program
// stops when it jumps outside itself.
//
// Part 1 is register b at the end, starting with both registers at 0.
// Part 2 is the same with register a starting at 1.
//
// The program is parsed once and then interpreted, with a cap on the steps.
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
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxSteps      = 100_000_000
	maxRegister   = 1 << 62
)

// instruction is one parsed line: an opcode, a register (0 for a, 1 for b)
// and a jump offset.
type instruction struct {
	op       string
	register int
	offset   int
}

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// parseInstruction reads a line such as "jio a, +2".
func parseInstruction(line string) (instruction, error) {
	fields := strings.Fields(strings.ReplaceAll(line, ",", " "))
	if len(fields) < 2 || len(fields) > 3 {
		return instruction{}, fmt.Errorf("malformed instruction %q", line)
	}

	in := instruction{op: fields[0]}
	args := fields[1:]
	switch in.op {
	case "hlf", "tpl", "inc", "jie", "jio":
		if args[0] != "a" && args[0] != "b" {
			return instruction{}, fmt.Errorf("unknown register in %q", line)
		}
		in.register = int(args[0][0] - 'a')
		args = args[1:]
	case "jmp":
	default:
		return instruction{}, fmt.Errorf("unknown instruction %q", line)
	}

	jumps := in.op == "jmp" || in.op == "jie" || in.op == "jio"
	if jumps != (len(args) == 1) {
		return instruction{}, fmt.Errorf("wrong number of arguments in %q", line)
	}
	if jumps {
		offset, err := strconv.Atoi(args[0])
		if err != nil {
			return instruction{}, fmt.Errorf("invalid offset in %q: %w", line, err)
		}
		in.offset = offset
	}
	return in, nil
}

func parseInput(path string) ([]instruction, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	program := make([]instruction, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		in, err := parseInstruction(line)
		if err != nil {
			return nil, err
		}
		program = append(program, in)
	}
	if len(program) == 0 {
		return nil, fmt.Errorf("%s: empty program", path)
	}
	return program, nil
}

// run executes the program from the given value of register a and returns b.
func run(program []instruction, a int) (int, error) {
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	if a < 0 {
		return 0, fmt.Errorf("register a cannot start at %d", a)
	}

	regs := [2]int{a, 0}
	pc := 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return regs[1], nil
		}
		in := program[pc]
		r := &regs[in.register]
		next := pc + 1
		switch in.op {
		case "hlf":
			*r /= 2
		case "tpl":
			*r *= 3
		case "inc":
			*r++
		case "jmp":
			next = pc + in.offset
		case "jie":
			if *r%2 == 0 {
				next = pc + in.offset
			}
		case "jio":
			if *r == 1 {
				next = pc + in.offset
			}
		default:
			return 0, fmt.Errorf("unknown instruction %q at %d", in.op, pc)
		}
		if *r < 0 || *r > maxRegister {
			return 0, fmt.Errorf("register overflowed at instruction %d", pc)
		}
		pc = next
	}
	return 0, fmt.Errorf("program did not stop within %d steps", maxSteps)
}

func day23(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}
	fromZero, err := run(program, 0)
	if err != nil {
		return err
	}
	fromOne, err := run(program, 1)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", fromZero, fromOne); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day23(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
