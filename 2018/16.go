// Day 16: Chronal Classification
//
// A device has four registers and sixteen opcodes (addr, addi, mulr, ...,
// eqrr), but their numbers are unknown. The input first lists samples -
// registers before, an instruction "op A B C", registers after - then a
// test program.
//
// Part 1 counts the samples that behave like three or more opcodes.
// Part 2 works out every opcode's number and runs the test program; it asks
// for register 0 at the end.
//
// For part 2 each opcode number starts with every operation possible and is
// narrowed to those all its samples fit. Then, repeatedly, an opcode left
// with a single operation claims it, and that operation is struck from the
// others - at most sixteen rounds.
//
// The sixteen operations are one switch over an operation number (in the
// order the puzzle lists them); a sample is checked against each in turn.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	opcodes       = 16
)

// sample is one observed instruction.
type sample struct {
	before, after [4]int
	in            [4]int
}

// execute applies operation op to regs with operands a, b, c. Operands that
// an operation reads as registers must be 0..3.
func execute(op int, regs [4]int, a, b, c int) ([4]int, error) {
	if op < 0 || op >= opcodes || c < 0 || c > 3 {
		return regs, fmt.Errorf("operation %d or output register %d is out of range", op, c)
	}
	// Operations that read A as a register: all but seti(9), gtir(10), eqir(13).
	// Those that read B as a register: the "rr" and "ir" forms.
	readsA := op != 9 && op != 10 && op != 13
	readsB := op == 0 || op == 2 || op == 4 || op == 6 || op == 10 || op == 12 || op == 13 || op == 15
	if (readsA && (a < 0 || a > 3)) || (readsB && (b < 0 || b > 3)) {
		return regs, errors.New("operand is not a register")
	}
	ra, rb := 0, 0
	if readsA {
		ra = regs[a]
	}
	if readsB {
		rb = regs[b]
	}
	var v int
	switch op {
	case 0:
		v = ra + rb
	case 1:
		v = ra + b
	case 2:
		v = ra * rb
	case 3:
		v = ra * b
	case 4:
		v = ra & rb
	case 5:
		v = ra & b
	case 6:
		v = ra | rb
	case 7:
		v = ra | b
	case 8:
		v = ra
	case 9:
		v = a
	case 10, 11, 12:
		if [3]bool{a > rb, ra > b, ra > rb}[op-10] {
			v = 1
		}
	case 13, 14, 15:
		if [3]bool{a == rb, ra == b, ra == rb}[op-13] {
			v = 1
		}
	}
	regs[c] = v
	return regs, nil
}

// parseNumbers reads every integer in a line such as "Before: [3, 2, 1, 1]".
func parseNumbers(line string) ([4]int, error) {
	var out [4]int
	if strings.TrimSpace(line) == "" {
		return out, errors.New("empty line where numbers were expected")
	}
	f := strings.Fields(strings.NewReplacer("Before:", "", "After:", "", "[", "", "]", "", ",", " ").Replace(line))
	if len(f) != 4 {
		return out, fmt.Errorf("line %q does not hold four numbers", line)
	}
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return out, fmt.Errorf("invalid number in %q: %w", line, err)
		}
		out[i] = v
	}
	return out, nil
}

// parseInput splits the input into samples and the test program.
func parseInput(path string) ([]sample, [][4]int, error) {
	if path == "" {
		return nil, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) > maxLines {
		return nil, nil, fmt.Errorf("%s: %d lines, more than %d", path, len(lines), maxLines)
	}
	samples, program := make([]sample, 0, len(lines)/4), make([][4]int, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		switch {
		case line == "":
		case strings.HasPrefix(line, "Before:"):
			if i+2 >= len(lines) {
				return nil, nil, errors.New("sample is cut short")
			}
			var s sample
			if s.before, err = parseNumbers(line); err == nil {
				if s.in, err = parseNumbers(lines[i+1]); err == nil {
					s.after, err = parseNumbers(lines[i+2])
				}
			}
			if err != nil {
				return nil, nil, err
			}
			samples, i = append(samples, s), i+2
		default:
			in, err := parseNumbers(line)
			if err != nil {
				return nil, nil, err
			}
			program = append(program, in)
		}
	}
	return samples, program, nil
}

// candidates returns the set of operations (a bitmask) a sample fits.
func candidates(s sample) (uint16, error) {
	var set uint16
	for op := range opcodes {
		got, err := execute(op, s.before, s.in[1], s.in[2], s.in[3])
		if err != nil {
			continue // this operation cannot read those operands
		}
		if got == s.after {
			set |= 1 << op
		}
	}
	if set == 0 {
		return 0, fmt.Errorf("sample %+v fits no operation", s)
	}
	return set, nil
}

// decode works out which operation each opcode number stands for.
func decode(samples []sample) ([opcodes]int, error) {
	var possible [opcodes]uint16
	var opOf [opcodes]int
	for i := range possible {
		possible[i], opOf[i] = 1<<opcodes-1, -1
	}
	for _, s := range samples {
		set, err := candidates(s)
		if err != nil {
			return opOf, err
		}
		if s.in[0] < 0 || s.in[0] >= opcodes {
			return opOf, fmt.Errorf("opcode %d is outside 0..%d", s.in[0], opcodes-1)
		}
		possible[s.in[0]] &= set
	}
	for range opcodes {
		for code, set := range possible {
			if opOf[code] >= 0 || set&(set-1) != 0 || set == 0 {
				continue // already known, or not yet down to one
			}
			op := bits.TrailingZeros16(set)
			opOf[code] = op
			for other := range possible {
				possible[other] &^= 1 << op
			}
		}
	}
	var claimed uint16
	for code, op := range opOf {
		if op < 0 {
			return opOf, fmt.Errorf("opcode %d could not be pinned down", code)
		}
		if claimed>>op&1 == 1 {
			return opOf, fmt.Errorf("operation %d was given to two opcodes", op)
		}
		claimed |= 1 << op
	}
	return opOf, nil
}

func day16(path string) error {
	samples, program, err := parseInput(path)
	if err != nil {
		return err
	}
	if len(samples) == 0 || len(program) == 0 {
		return fmt.Errorf("%s: need samples and a program (%d, %d)", path, len(samples), len(program))
	}
	ambiguous := 0
	for _, s := range samples {
		set, err := candidates(s)
		if err != nil {
			return err
		}
		n := 0
		for op := range opcodes {
			n += int(set >> op & 1)
		}
		if n >= 3 {
			ambiguous++
		}
	}

	opOf, err := decode(samples)
	if err != nil {
		return err
	}
	var regs [4]int
	for _, in := range program {
		if in[0] < 0 || in[0] >= opcodes {
			return fmt.Errorf("opcode %d is outside 0..%d", in[0], opcodes-1)
		}
		if regs, err = execute(opOf[in[0]], regs, in[1], in[2], in[3]); err != nil {
			return err
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", ambiguous, regs[0]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
