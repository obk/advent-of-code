// Day 21: Chronal Conversion
//
// The input is a program for the device of day 19 (six registers, "#ip"
// binding). It loops forever unless register 0 matches a value it computes;
// you choose register 0.
//
// Part 1 is the value of register 0 that makes it halt after the fewest
// instructions.
// Part 2 is the value that makes it halt after the most instructions.
//
// The program reads register 0 in exactly one place, an "eqrr" that compares
// it with another register and halts on equality. The value that other
// register holds the first time that instruction runs is the earliest
// possible halt. The values compared there eventually repeat; the last new
// value before the first repeat is the latest possible halt.
//
// Finding that takes thousands of passes through a slow inner loop that
// divides by 256 by counting up. That loop is recognised by its exact shape
// (checked instruction by instruction) and replaced by one division; the
// values seen are marked in a fixed table, as they are all below 2^24.
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
	inputPath     = "inputs/21.txt"
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

// findCheck returns the index of the instruction comparing register 0 and
// the other register it compares with.
func findCheck(program []instruction) (at, other int, err error) {
	at = -1
	for i, in := range program {
		if in.op != 15 || (in.a != 0 && in.b != 0) {
			continue // not "eqrr" with register 0
		}
		if at >= 0 {
			return 0, 0, errors.New("register 0 is compared more than once")
		}
		at, other = i, in.a+in.b // the operand that is not 0
	}
	if at < 0 || other < 1 || other >= registers {
		return 0, 0, errors.New("no instruction compares register 0 with another register")
	}
	return at, other, nil
}

// firstValue runs the program until it reaches the check and returns the
// value it compares register 0 with.
func firstValue(ip int, program []instruction, at, other int) (int, error) {
	if ip < 0 || ip >= registers || at < 0 || at >= len(program) {
		return 0, errors.New("invalid ip binding or check position")
	}
	var regs [registers]int
	pc := 0
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return 0, errors.New("the program halted before reaching the check")
		}
		if pc == at {
			return regs[other], nil
		}
		regs[ip] = pc
		if err := execute(program[pc], &regs); err != nil {
			return 0, err
		}
		pc = regs[ip] + 1
	}
	return 0, fmt.Errorf("the check was not reached within %d steps", maxSteps)
}

// divisionLoop reports whether a counting division loop starts at i, and if
// so its counter register, dividend register, divisor and exit address.
func divisionLoop(program []instruction, ip, i int) (counter, dividend, divisor, exit int, ok bool) {
	if i < 0 || i+9 > len(program) {
		return 0, 0, 0, 0, false
	}
	p := program[i : i+9]
	a, b := p[0].c, p[1].c
	ok = p[0].op == 9 && p[0].a == 0 && // seti 0 _ A
		p[1].op == 1 && p[1].a == a && p[1].b == 1 && // addi A 1 B
		p[2].op == 3 && p[2].a == b && p[2].c == b && p[2].b > 0 && // muli B N B
		p[3].op == 12 && p[3].a == b && p[3].c == b && // gtrr B C B
		p[4].op == 0 && p[4].a == b && p[4].b == ip && p[4].c == ip && // addr B ip ip
		p[5].op == 1 && p[5].a == ip && p[5].b == 1 && p[5].c == ip && // addi ip 1 ip
		p[6].op == 9 && p[6].c == ip && // seti X _ ip
		p[7].op == 1 && p[7].a == a && p[7].b == 1 && p[7].c == a && // addi A 1 A
		p[8].op == 9 && p[8].a == i && p[8].c == ip // seti i _ ip
	return a, p[3].b, p[2].b, p[6].a + 1, ok
}

// lastValue runs the program, short-cutting the division loop, and returns
// the last new value compared at the check before the values repeat. seen
// is scratch space for every 24-bit value.
func lastValue(ip int, program []instruction, at, other int, seen []bool) (first, last int, err error) {
	if ip < 0 || ip >= registers || at < 0 || at >= len(program) || len(seen) != 1<<24 {
		return 0, 0, errors.New("invalid ip binding, check position or scratch space")
	}
	var regs [registers]int
	pc, first, last := 0, -1, -1
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return 0, 0, errors.New("the program halted on its own")
		}
		if pc == at {
			v := regs[other]
			if v < 0 || v >= len(seen) {
				return 0, 0, fmt.Errorf("compared value %d is not 24-bit", v)
			}
			if seen[v] {
				return first, last, nil
			}
			if first < 0 {
				first = v
			}
			seen[v], last = true, v
		}
		if counter, dividend, divisor, exit, ok := divisionLoop(program, ip, pc); ok {
			regs[counter] = regs[dividend] / divisor
			regs[program[pc+1].c] = 1 // the comparison that ends the loop
			pc = exit
			continue
		}
		regs[ip] = pc
		if err := execute(program[pc], &regs); err != nil {
			return 0, 0, err
		}
		pc = regs[ip] + 1
	}
	return 0, 0, fmt.Errorf("the values did not repeat within %d steps", maxSteps)
}

func day21(path string) error {
	ip, program, err := parseProgram(path)
	if err != nil {
		return err
	}
	at, other, err := findCheck(program)
	if err != nil {
		return err
	}
	first, err := firstValue(ip, program, at, other)
	if err != nil {
		return err
	}

	check, last, err := lastValue(ip, program, at, other, make([]bool, 1<<24))
	if err != nil {
		return err
	}
	if check != first {
		return fmt.Errorf("the short-cut run first compares %d, the plain run %d", check, first)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, last); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
