// Day 23: Safe Cracking
//
// The input is an assembunny program (cpy, inc, dec, jnz) with one new
// instruction, "tgl x", which toggles the instruction x places away:
// inc <-> dec and any other one-argument instruction becomes inc; jnz <-> cpy
// and any other two-argument instruction becomes jnz. A toggled instruction
// that ends up invalid (such as cpy into a constant) is skipped.
//
// Part 1 is register a at the end, starting with a = 7.
// Part 2 starts with a = 12, which takes billions of steps one increment at a
// time.
//
// The program is parsed into a mutable slice of instructions, so tgl just
// rewrites an opcode in place. Its slow part is multiplication by repeated
// addition, in the six-instruction shape
//
//	cpy s t / inc r / dec t / jnz t -2 / dec u / jnz u -5
//
// which leaves r += s*u and t = u = 0. Whenever the running program reaches
// that exact shape (checked afresh each time, since tgl may have rewritten
// it), the six instructions are done in one step.
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
	maxSteps      = 10_000_000_000
	maxConstant   = 1_000_000
)

// operand is a register (0-3) or a constant.
type operand struct {
	isReg bool
	value int
}

// instruction is one line; y is unused by one-argument instructions.
type instruction struct {
	op   string
	args int
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
		if len(f) < 2 || len(f) > 3 {
			return nil, fmt.Errorf("malformed instruction %q", line)
		}
		in := instruction{op: f[0], args: len(f) - 1}
		if !strings.Contains(" cpy inc dec jnz tgl ", " "+in.op+" ") {
			return nil, fmt.Errorf("unknown instruction %q", line)
		}
		if (in.args == 2) != (in.op == "cpy" || in.op == "jnz") {
			return nil, fmt.Errorf("wrong number of arguments in %q", line)
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
	return program, nil
}

// toggle rewrites the instruction at i as tgl describes.
func toggle(program []instruction, i int) error {
	if len(program) == 0 {
		return errors.New("empty program")
	}
	if i < 0 || i >= len(program) {
		return nil // toggling outside the program does nothing
	}
	in := &program[i]
	switch {
	case in.args == 1 && in.op == "inc":
		in.op = "dec"
	case in.args == 1:
		in.op = "inc"
	case in.args == 2 && in.op == "jnz":
		in.op = "cpy"
	case in.args == 2:
		in.op = "jnz"
	default:
		return fmt.Errorf("instruction %d has %d arguments", i, in.args)
	}
	return nil
}

// run executes a copy of the program (in work) from the given registers and
// returns register a.
func run(program, work []instruction, regs [4]int) (int, error) {
	if len(work) != len(program) || len(program) == 0 {
		return 0, errors.New("work space does not match the program")
	}
	for i, in := range program {
		if (in.x.isReg && in.x.value > 3) || (in.y.isReg && in.y.value > 3) {
			return 0, fmt.Errorf("instruction %d names a register outside a-d", i)
		}
	}
	copy(work, program)

	pc := 0
	for range maxSteps {
		if pc < 0 || pc >= len(work) {
			return regs[0], nil
		}
		done, err := multiply(work, pc, &regs)
		if err != nil {
			return 0, err
		}
		if done {
			pc += 6
			continue
		}
		in := work[pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		if in.y.isReg {
			y = regs[in.y.value]
		}
		switch {
		case in.op == "cpy" && in.args == 2 && in.y.isReg:
			regs[in.y.value] = x
		case in.op == "inc" && in.args == 1 && in.x.isReg:
			regs[in.x.value]++
		case in.op == "dec" && in.args == 1 && in.x.isReg:
			regs[in.x.value]--
		case in.op == "jnz" && in.args == 2 && x != 0:
			pc += y - 1
		case in.op == "tgl" && in.args == 1:
			if err := toggle(work, pc+x); err != nil {
				return 0, err
			}
		}
		pc++ // anything that matched no case above is invalid and skipped
	}
	return 0, fmt.Errorf("program did not halt within %d steps", int64(maxSteps))
}

// multiply applies the multiplication shortcut at pc if the program there has
// exactly the right shape, reporting whether it did.
func multiply(work []instruction, pc int, regs *[4]int) (bool, error) {
	if pc < 0 || pc+6 > len(work) {
		return false, nil
	}
	if regs == nil {
		return false, errors.New("no registers")
	}
	w := work[pc : pc+6]
	r, t, u := w[1].x, w[0].y, w[4].x
	shape := w[0].op == "cpy" && w[0].args == 2 && t.isReg &&
		w[1].op == "inc" && r.isReg && r.value != t.value && r.value != u.value &&
		w[2].op == "dec" && w[2].x == t &&
		w[3].op == "jnz" && w[3].x == t && w[3].y == operand{value: -2} &&
		w[4].op == "dec" && u.isReg && u.value != t.value &&
		w[5].op == "jnz" && w[5].x == u && w[5].y == operand{value: -5}
	if !shape {
		return false, nil
	}
	s := w[0].x.value
	if w[0].x.isReg {
		s = regs[w[0].x.value]
	}
	if s != 0 && (regs[u.value] > (1<<62)/max(s, -s) || regs[u.value] < -(1<<62)/max(s, -s)) {
		return false, fmt.Errorf("multiplying %d by %d overflows", s, regs[u.value])
	}
	regs[r.value] += s * regs[u.value]
	regs[t.value], regs[u.value] = 0, 0
	return true, nil
}

func day23(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}

	work := make([]instruction, len(program))
	a, err := run(program, work, [4]int{7, 0, 0, 0})
	if err != nil {
		return err
	}
	b, err := run(program, work, [4]int{12, 0, 0, 0})
	if err != nil {
		return err
	}
	if a < 0 || b < 0 {
		return fmt.Errorf("safe codes %d and %d should not be negative", a, b)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", a, b); err != nil {
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
