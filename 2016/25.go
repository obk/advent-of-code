// Day 25: Clock Signal
//
// The input is an assembunny program (cpy, inc, dec, jnz) with one new
// instruction, "out x", which transmits x. The program runs forever.
//
// Part 1 is the lowest positive starting value of register a that makes the
// program transmit 0, 1, 0, 1, ... forever. (Day 25 has no second puzzle.)
//
// "Forever" is checked as far as it practically can be: each candidate a is
// run until it has sent a long run of alternating bits, and rejected at the
// first bit out of place. These programs repeat their output with a short
// period, so a long run settles the question.
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
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxConstant   = 1_000_000
	maxCandidate  = 1_000_000
	signalLength  = 1_000       // alternating bits needed to accept
	maxSteps      = 100_000_000 // per candidate
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
	if v < -maxConstant || v > maxConstant {
		return operand{}, fmt.Errorf("constant %d is out of range", v)
	}
	return operand{value: v}, nil
}

func parseInput(path string) ([]instruction, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	program := make([]instruction, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		in := instruction{op: f[0]}
		if !strings.Contains(" cpy inc dec jnz out ", " "+in.op+" ") {
			return nil, fmt.Errorf("unknown instruction %q", line)
		}
		if len(f) < 2 || len(f) > 3 || (len(f) == 3) != (in.op == "cpy" || in.op == "jnz") {
			return nil, fmt.Errorf("malformed instruction %q", line)
		}
		if in.x, err = parseOperand(f[1]); err != nil {
			return nil, err
		}
		if len(f) == 3 {
			if in.y, err = parseOperand(f[2]); err != nil {
				return nil, err
			}
		}
		program = append(program, in)
	}
	if len(program) == 0 {
		return nil, fmt.Errorf("%s: empty program", path)
	}
	sends := false
	for _, in := range program {
		sends = sends || in.op == "out"
	}
	if !sends {
		return nil, fmt.Errorf("%s: the program never transmits anything", path)
	}
	return program, nil
}

// alternates reports whether the program, started with a, sends signalLength
// bits of 0, 1, 0, 1, ...
func alternates(program []instruction, a int) (bool, error) {
	if len(program) == 0 || a < 0 {
		return false, errors.New("nothing to run")
	}
	if len(program) > maxLines {
		return false, fmt.Errorf("program of %d instructions is too long", len(program))
	}
	regs := [4]int{a, 0, 0, 0}
	pc, sent := 0, 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return false, nil // the signal stopped
		}
		in := program[pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		if in.y.isReg {
			y = regs[in.y.value]
		}
		switch {
		case in.op == "cpy" && in.y.isReg:
			regs[in.y.value] = x
		case in.op == "inc" && in.x.isReg:
			regs[in.x.value]++
		case in.op == "dec" && in.x.isReg:
			regs[in.x.value]--
		case in.op == "jnz":
			if x != 0 {
				pc += y - 1
			}
		case in.op == "out":
			if x != sent%2 {
				return false, nil
			}
			if sent++; sent == signalLength {
				return true, nil
			}
		default:
			return false, fmt.Errorf("invalid instruction %+v at %d", in, pc)
		}
		pc++
	}
	return false, fmt.Errorf("start %d sent only %d bits within %d steps", a, sent, maxSteps)
}

func day25(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}

	for a := 1; a <= maxCandidate; a++ {
		ok, err := alternates(program, a)
		if err != nil {
			return err
		}
		if ok {
			if _, err := fmt.Printf("Part 1: %d\n", a); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("no start value up to %d sends a clock signal", maxCandidate)
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
