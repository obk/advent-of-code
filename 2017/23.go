// Day 23: Coprocessor Conflagration
//
// The input is a program for registers a-h: set, sub and mul change a
// register, and jnz X Y jumps by Y if X is not zero.
//
// Part 1 counts how many times mul runs.
// Part 2 starts with a = 1 and asks for h at the end, which the program
// would take far too long to compute.
//
// The program is parsed once and interpreted with a cap on the steps. For
// part 2 it is read rather than run: it walks b from its start up to c in
// steps of k, and adds one to h for every b that is not prime (its inner
// loops try every d*e == b). So the program is run only until it reaches
// "set f 1", where b and c are set, k is read from the "sub b -k" near the
// end, and the composites are counted directly. The shape this relies on is
// checked first.
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
	maxSteps      = 1_000_000_000
	registers     = 8
)

// operand is a register (0-7) or a constant.
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

// parseOperand reads a register letter a-h or an integer.
func parseOperand(text string) (operand, error) {
	if text == "" {
		return operand{}, errors.New("empty operand")
	}
	if len(text) == 1 && text[0] >= 'a' && text[0] < 'a'+registers {
		return operand{isReg: true, value: int(text[0] - 'a')}, nil
	}
	v, err := strconv.Atoi(text)
	if err != nil {
		return operand{}, fmt.Errorf("invalid operand %q: %w", text, err)
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
		if len(f) != 3 || !strings.Contains(" set sub mul jnz ", " "+f[0]+" ") {
			return nil, fmt.Errorf("invalid instruction %q", line)
		}
		in := instruction{op: f[0]}
		if in.x, err = parseOperand(f[1]); err != nil {
			return nil, err
		}
		if in.y, err = parseOperand(f[2]); err != nil {
			return nil, err
		}
		if in.op != "jnz" && !in.x.isReg {
			return nil, fmt.Errorf("instruction %q must name a register", line)
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

// countMuls runs the program from the given registers and counts the muls.
func countMuls(program []instruction, regs [registers]int) (int, error) {
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	if len(program) > maxLines {
		return 0, fmt.Errorf("%d instructions, more than %d", len(program), maxLines)
	}
	pc, muls := 0, 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return muls, nil
		}
		in := program[pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		if in.y.isReg {
			y = regs[in.y.value]
		}
		switch in.op {
		case "set":
			regs[in.x.value] = y
		case "sub":
			regs[in.x.value] -= y
		case "mul":
			regs[in.x.value] *= y
			muls++
		case "jnz":
			if x != 0 {
				pc += y - 1
			}
		}
		pc++
	}
	return 0, fmt.Errorf("the program did not halt within %d steps", maxSteps)
}

// composites counts the non-primes among b, b+k, ..., c.
func composites(b, c, k int) (int, error) {
	if b < 2 || c < b || k < 1 || (c-b)/k > maxSteps {
		return 0, fmt.Errorf("cannot walk %d..%d in steps of %d", b, c, k)
	}
	total, count := 0, 0
	for n := b; n <= c; n += k {
		total++
		for d := 2; d*d <= n; d++ {
			if n%d == 0 {
				count++
				break
			}
		}
	}
	if count > total {
		return 0, fmt.Errorf("%d composites among %d numbers", count, total)
	}
	return count, nil
}

// readH works out what part 2's h would be, as the package comment explains.
func readH(program []instruction) (int, error) {
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	loop, step := -1, -1
	for i, in := range program {
		if in.op == "set" && in.x == (operand{isReg: true, value: 'f' - 'a'}) && in.y == (operand{value: 1}) && loop < 0 {
			loop = i
		}
		if in.op == "sub" && in.x == (operand{isReg: true, value: 'b' - 'a'}) && !in.y.isReg && in.y.value < 0 {
			step = -in.y.value
		}
	}
	if loop < 0 || step < 1 {
		return 0, errors.New("the program does not have the expected composite-counting shape")
	}
	var regs [registers]int
	regs[0] = 1
	pc := 0
	for steps := 0; steps < len(program) && pc != loop; steps++ {
		if pc < 0 || pc > loop {
			return 0, fmt.Errorf("the set-up jumped to %d, past the main loop at %d", pc, loop)
		}
		in := program[pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		if in.y.isReg {
			y = regs[in.y.value]
		}
		switch in.op {
		case "set":
			regs[in.x.value] = y
		case "sub":
			regs[in.x.value] -= y
		case "mul":
			regs[in.x.value] *= y
		case "jnz":
			if x != 0 {
				pc += y - 1
			}
		}
		pc++
	}
	if pc != loop {
		return 0, errors.New("the set-up never reached the main loop")
	}
	return composites(regs[1], regs[2], step)
}

func day23(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}
	muls, err := countMuls(program, [registers]int{})
	if err != nil {
		return err
	}

	h, err := readH(program)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", muls, h); err != nil {
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
