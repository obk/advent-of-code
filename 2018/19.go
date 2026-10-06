// Day 19: Go With The Flow
//
// The input is a program for the device of day 16, now with six registers
// and a "#ip r" line binding the instruction pointer to register r: the
// pointer is written to r before each instruction and read back after it,
// then advanced by one. The program halts when the pointer leaves it.
//
// Part 1 is register 0 when the program halts.
// Part 2 starts with register 0 at 1, which makes it run for far too long.
//
// Instructions are parsed into operation numbers (day 16's order) once and
// interpreted with a cap on the steps. The program sums the divisors of a
// number it first builds: it sets that number up, then jumps back to
// instruction 1 to start its slow double loop. So for part 2 it is run only
// until the pointer first returns to 1, the largest register is taken as the
// number, and its divisors are summed directly. The same shortcut is applied
// to part 1, where it must agree with running the program in full.
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
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	maxSteps      = 100_000_000
	registers     = 6
)

// names lists the operations in day 16's order.
const names = "addr addi mulr muli banr bani borr bori setr seti gtir gtri gtrr eqir eqri eqrr"

// instruction is one parsed line.
type instruction struct {
	op, a, b, c int
}

// parseProgram reads the "#ip" line and the instructions.
func parseProgram(path string) (ip int, program []instruction, err error) {
	if path == "" {
		return 0, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 || len(lines) > maxLines {
		return 0, nil, fmt.Errorf("%s: %d lines, expected 2..%d", path, len(lines), maxLines)
	}
	if ip, err = strconv.Atoi(strings.TrimPrefix(lines[0], "#ip ")); err != nil || ip < 0 || ip >= registers {
		return 0, nil, fmt.Errorf("%s: invalid ip binding %q", path, lines[0])
	}
	program = make([]instruction, 0, len(lines)-1)
	opNames := strings.Fields(names)
	for _, line := range lines[1:] {
		f := strings.Fields(line)
		if len(f) != 4 {
			return 0, nil, fmt.Errorf("%s: malformed instruction %q", path, line)
		}
		in := instruction{op: -1}
		for i, n := range opNames {
			if n == f[0] {
				in.op = i
			}
		}
		var n [3]int
		for i := range n {
			if n[i], err = strconv.Atoi(f[i+1]); err != nil {
				return 0, nil, fmt.Errorf("%s: invalid operand in %q: %w", path, line, err)
			}
		}
		in.a, in.b, in.c = n[0], n[1], n[2]
		if in.op < 0 || in.c < 0 || in.c >= registers {
			return 0, nil, fmt.Errorf("%s: invalid instruction %q", path, line)
		}
		program = append(program, in)
	}
	return ip, program, nil
}

// execute applies one instruction to the registers.
func execute(in instruction, regs *[registers]int) error {
	readsA := in.op != 9 && in.op != 10 && in.op != 13
	readsB := in.op == 0 || in.op == 2 || in.op == 4 || in.op == 6 || in.op == 10 || in.op == 12 || in.op == 13 || in.op == 15
	if (readsA && (in.a < 0 || in.a >= registers)) || (readsB && (in.b < 0 || in.b >= registers)) {
		return fmt.Errorf("instruction %+v reads a register that does not exist", in)
	}
	ra, rb, a, b := 0, 0, in.a, in.b
	if readsA {
		ra = regs[a]
	}
	if readsB {
		rb = regs[b]
	}
	results := [16]int{ra + rb, ra + b, ra * rb, ra * b, ra & rb, ra & b, ra | rb, ra | b, ra, a}
	switch {
	case in.op <= 9:
		regs[in.c] = results[in.op]
	case in.op <= 12:
		regs[in.c] = 0
		if [3]bool{a > rb, ra > b, ra > rb}[in.op-10] {
			regs[in.c] = 1
		}
	case in.op <= 15:
		regs[in.c] = 0
		if [3]bool{a == rb, ra == b, ra == rb}[in.op-13] {
			regs[in.c] = 1
		}
	default:
		return fmt.Errorf("unknown operation %d", in.op)
	}
	return nil
}

// run executes the program from the given registers until it halts.
func run(ip int, program []instruction, regs [registers]int) ([registers]int, error) {
	if ip < 0 || ip >= registers || len(program) == 0 {
		return regs, errors.New("invalid ip binding or empty program")
	}
	pc := 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return regs, nil
		}
		regs[ip] = pc
		if err := execute(program[pc], &regs); err != nil {
			return regs, err
		}
		pc = regs[ip] + 1
	}
	return regs, fmt.Errorf("the program did not halt within %d steps", maxSteps)
}

// divisorSum runs the program's set-up from regs and sums the divisors of
// the number it builds.
func divisorSum(ip int, program []instruction, regs [registers]int) (int, error) {
	if ip < 0 || ip >= registers || len(program) < 2 {
		return 0, errors.New("invalid ip binding or program")
	}
	pc := 0
	for steps := 0; steps == 0 || pc != 1; steps++ {
		if steps > 1000 || pc < 0 || pc >= len(program) {
			return 0, errors.New("the set-up never jumps back to instruction 1")
		}
		regs[ip] = pc
		if err := execute(program[pc], &regs); err != nil {
			return 0, err
		}
		pc = regs[ip] + 1
	}
	n := 0
	for _, r := range regs {
		n = max(n, r)
	}
	if n < 1 || n > 1<<40 {
		return 0, fmt.Errorf("target %d is out of range", n)
	}
	sum := 0
	for d := 1; d*d <= n; d++ {
		if n%d == 0 {
			sum += d
			if d*d != n {
				sum += n / d
			}
		}
	}
	return sum, nil
}

func day19(path string) error {
	ip, program, err := parseProgram(path)
	if err != nil {
		return err
	}
	regs, err := run(ip, program, [registers]int{})
	if err != nil {
		return err
	}

	check, err := divisorSum(ip, program, [registers]int{})
	if err != nil {
		return err
	}
	if check != regs[0] {
		return fmt.Errorf("the divisor-sum reading gives %d for part 1, not %d", check, regs[0])
	}
	sum, err := divisorSum(ip, program, [registers]int{1})
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", regs[0], sum); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
