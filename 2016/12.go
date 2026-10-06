// Day 12: Leonardo's Monorail
//
// The input is an assembunny program for four registers a-d, all starting at
// 0: "cpy x y" copies a value or register into a register, "inc x" and
// "dec x" add or subtract one, and "jnz x y" jumps by y if x is not zero.
//
// Part 1 is register a when the program halts.
// Part 2 is the same with register c starting at 1.
//
// The program is parsed once into instructions whose operands are already
// resolved into register numbers or constants, then interpreted with a cap on
// the steps.
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
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxSteps      = 1_000_000_000
)

// operand is a register (0-3) or a constant.
type operand struct {
	isReg bool
	value int
}

// instruction is one parsed line.
type instruction struct {
	op   string
	x, y operand
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

// parseOperand reads a register name a-d or an integer.
func parseOperand(text string) (operand, error) {
	if text == "" {
		return operand{}, errors.New("empty operand")
	}
	if len(text) == 1 && text[0] >= 'a' && text[0] <= 'd' {
		return operand{isReg: true, value: int(text[0] - 'a')}, nil
	}
	v, err := strconv.Atoi(text)
	if err != nil {
		return operand{}, fmt.Errorf("invalid operand %q: %w", text, err)
	}
	if v < -maxLines*maxLines || v > maxLines*maxLines {
		return operand{}, fmt.Errorf("constant %d is out of range", v)
	}
	return operand{value: v}, nil
}

// parseInstruction reads one line and checks its operands fit its opcode.
func parseInstruction(line string) (instruction, error) {
	f := strings.Fields(line)
	if len(f) < 2 || len(f) > 3 {
		return instruction{}, fmt.Errorf("malformed instruction %q", line)
	}
	in := instruction{op: f[0]}
	var err error
	if in.x, err = parseOperand(f[1]); err != nil {
		return instruction{}, err
	}
	if len(f) == 3 {
		if in.y, err = parseOperand(f[2]); err != nil {
			return instruction{}, err
		}
	}

	switch {
	case (in.op == "inc" || in.op == "dec") && len(f) == 2 && in.x.isReg:
	case in.op == "cpy" && len(f) == 3 && in.y.isReg:
	case in.op == "jnz" && len(f) == 3:
	default:
		return instruction{}, fmt.Errorf("invalid instruction %q", line)
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
	if len(program) > maxLines {
		return nil, fmt.Errorf("%s: %d instructions, more than %d", path, len(program), maxLines)
	}
	return program, nil
}

// run executes the program from the given registers and returns register a.
func run(program []instruction, regs [4]int) (int, error) {
	if len(program) == 0 || len(program) > maxLines {
		return 0, fmt.Errorf("program of %d instructions is outside 1..%d", len(program), maxLines)
	}
	for i, in := range program {
		if (in.x.isReg && (in.x.value < 0 || in.x.value > 3)) || (in.y.isReg && (in.y.value < 0 || in.y.value > 3)) {
			return 0, fmt.Errorf("instruction %d names a register outside a-d", i)
		}
	}

	pc := 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return regs[0], nil
		}
		in := program[pc]
		x := in.x.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		switch in.op {
		case "cpy":
			regs[in.y.value] = x
		case "inc":
			regs[in.x.value]++
		case "dec":
			regs[in.x.value]--
		case "jnz":
			y := in.y.value
			if in.y.isReg {
				y = regs[in.y.value]
			}
			if x != 0 && y == 0 {
				return 0, fmt.Errorf("jump by 0 at %d would loop forever", pc)
			}
			if x != 0 {
				pc += y - 1
			}
		default:
			return 0, fmt.Errorf("unknown instruction %q at %d", in.op, pc)
		}
		pc++
	}
	return 0, fmt.Errorf("program did not halt within %d steps", maxSteps)
}

func day12(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}
	if len(program) == 0 {
		return errors.New("empty program")
	}
	a, err := run(program, [4]int{})
	if err != nil {
		return err
	}
	withKey, err := run(program, [4]int{0, 0, 1, 0})
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", a, withKey); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
