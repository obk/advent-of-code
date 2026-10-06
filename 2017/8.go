// Day 8: I Heard You Like Registers
//
// Each line of the input is a conditional register update:
// "b inc 5 if a > 1". All registers start at 0.
//
// Part 1 is the largest register value after every instruction has run.
// Part 2 is the largest value any register held at any point.
//
// Register names are numbered while parsing, so running the program only
// touches a slice of integers.
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
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxAmount     = 1_000_000
)

// instruction is one parsed line.
type instruction struct {
	target, delta int
	reg, compare  int
	op            string
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

// regID returns the number of a register, numbering it if new.
func regID(ids map[string]int, name string) (int, error) {
	if name == "" || strings.Trim(name, "abcdefghijklmnopqrstuvwxyz") != "" {
		return 0, fmt.Errorf("invalid register name %q", name)
	}
	if id, ok := ids[name]; ok {
		return id, nil
	}
	if len(ids) >= 2*maxLines {
		return 0, fmt.Errorf("more than %d registers", 2*maxLines)
	}
	ids[name] = len(ids)
	return ids[name], nil
}

// parseInstruction reads "b inc 5 if a > 1".
func parseInstruction(line string, ids map[string]int) (instruction, error) {
	w := strings.Fields(line)
	if len(w) != 7 || (w[1] != "inc" && w[1] != "dec") || w[3] != "if" {
		return instruction{}, fmt.Errorf("malformed instruction %q", line)
	}
	switch w[5] {
	case "<", ">", "<=", ">=", "==", "!=":
	default:
		return instruction{}, fmt.Errorf("unknown comparison in %q", line)
	}
	in := instruction{op: w[5]}
	var err error
	if in.target, err = regID(ids, w[0]); err != nil {
		return instruction{}, err
	}
	if in.reg, err = regID(ids, w[4]); err != nil {
		return instruction{}, err
	}
	if in.delta, err = strconv.Atoi(w[2]); err != nil {
		return instruction{}, fmt.Errorf("invalid amount in %q: %w", line, err)
	}
	if in.compare, err = strconv.Atoi(w[6]); err != nil {
		return instruction{}, fmt.Errorf("invalid value in %q: %w", line, err)
	}
	if in.delta < -maxAmount || in.delta > maxAmount {
		return instruction{}, fmt.Errorf("amount in %q is outside ±%d", line, maxAmount)
	}
	if in.compare < -maxAmount || in.compare > maxAmount {
		return instruction{}, fmt.Errorf("value in %q is outside ±%d", line, maxAmount)
	}
	if w[1] == "dec" {
		in.delta = -in.delta
	}
	return in, nil
}

// holds evaluates a condition.
func holds(v int, op string, compare int) (bool, error) {
	if op == "" {
		return false, errors.New("missing comparison")
	}
	switch op {
	case "<":
		return v < compare, nil
	case ">":
		return v > compare, nil
	case "<=":
		return v <= compare, nil
	case ">=":
		return v >= compare, nil
	case "==":
		return v == compare, nil
	case "!=":
		return v != compare, nil
	}
	return false, fmt.Errorf("unknown comparison %q", op)
}

// run executes the program on regs (all zero) and returns the largest value
// at the end and the largest ever held.
func run(program []instruction, regs []int) (final, highest int, err error) {
	if len(regs) == 0 || len(program) == 0 {
		return 0, 0, errors.New("no registers or no instructions")
	}
	for _, in := range program {
		if in.target >= len(regs) || in.reg >= len(regs) {
			return 0, 0, fmt.Errorf("instruction %+v names an unknown register", in)
		}
		ok, err := holds(regs[in.reg], in.op, in.compare)
		if err != nil {
			return 0, 0, err
		}
		if ok {
			regs[in.target] += in.delta
			highest = max(highest, regs[in.target])
		}
	}
	final = regs[0]
	for _, v := range regs {
		final = max(final, v)
	}
	if final > highest && final > 0 {
		return 0, 0, fmt.Errorf("final value %d beats the highest seen %d", final, highest)
	}
	return final, highest, nil
}

func day8(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	ids := make(map[string]int, 2*len(lines))
	program := make([]instruction, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		in, err := parseInstruction(line, ids)
		if err != nil {
			return err
		}
		program = append(program, in)
	}

	if len(program) == 0 {
		return fmt.Errorf("%s: no instructions", path)
	}
	largest, highest, err := run(program, make([]int, len(ids)))
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", largest, highest); err != nil {
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
