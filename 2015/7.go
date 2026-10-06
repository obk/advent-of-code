// Day 7: Some Assembly Required
//
// Each line of the input wires up one 16-bit signal: a constant, another wire,
// or a gate (AND, OR, LSHIFT, RSHIFT, NOT) of wires and constants, ending in
// "-> wire". A gate produces nothing until all its inputs have a signal.
//
// Part 1 is the signal that ends up on wire a.
// Part 2 overrides wire b with that signal, resets everything else, and asks
// for wire a again.
//
// Rather than recursing from wire a, the circuit is settled in sweeps: each
// sweep evaluates every gate whose inputs are all known. Every sweep but the
// last settles at least one wire, so there are at most as many sweeps as
// wires.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
)

// The gates a wire can be driven by. A plain value or wire is opAssign.
const (
	opAssign = iota
	opAnd
	opOr
	opLshift
	opRshift
	opNot
)

// operand is either a constant or the index of a wire.
type operand struct {
	isWire bool
	value  int // the constant, or the wire index
}

// gate drives one wire.
type gate struct {
	op     int
	a, b   operand // b is unused by opAssign and opNot
	target int
}

// circuit is the parsed wiring: one gate per wire, wires numbered by name.
type circuit struct {
	gates []gate
	ids   map[string]int
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

// wireID returns the number of the named wire, numbering it if it is new.
func (c *circuit) wireID(name string) (int, error) {
	if name == "" || strings.Trim(name, "abcdefghijklmnopqrstuvwxyz") != "" {
		return 0, fmt.Errorf("invalid wire name %q", name)
	}
	if id, ok := c.ids[name]; ok {
		return id, nil
	}
	if len(c.ids) >= maxLines {
		return 0, fmt.Errorf("more than %d wires", maxLines)
	}
	c.ids[name] = len(c.ids)
	return c.ids[name], nil
}

// parseOperand reads a 16-bit constant or a wire name.
func (c *circuit) parseOperand(text string) (operand, error) {
	if text == "" {
		return operand{}, errors.New("missing operand")
	}
	if text[0] < '0' || text[0] > '9' {
		id, err := c.wireID(text)
		return operand{isWire: true, value: id}, err
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return operand{}, fmt.Errorf("invalid constant %q: %w", text, err)
	}
	if value < 0 || value > 0xFFFF {
		return operand{}, fmt.Errorf("constant %d does not fit in 16 bits", value)
	}
	return operand{value: value}, nil
}

// parseGate reads a line such as "x AND y -> z".
func (c *circuit) parseGate(line string) (gate, error) {
	source, target, found := strings.Cut(line, " -> ")
	if !found {
		return gate{}, fmt.Errorf("malformed line %q", line)
	}
	var g gate
	var err error
	if g.target, err = c.wireID(target); err != nil {
		return gate{}, err
	}

	words := strings.Fields(source)
	switch len(words) {
	case 1:
		g.op = opAssign
		g.a, err = c.parseOperand(words[0])
	case 2:
		if words[0] != "NOT" {
			return gate{}, fmt.Errorf("unknown gate in %q", line)
		}
		g.op = opNot
		g.a, err = c.parseOperand(words[1])
	case 3:
		ops := map[string]int{"AND": opAnd, "OR": opOr, "LSHIFT": opLshift, "RSHIFT": opRshift}
		op, ok := ops[words[1]]
		if !ok {
			return gate{}, fmt.Errorf("unknown gate in %q", line)
		}
		g.op = op
		if g.a, err = c.parseOperand(words[0]); err == nil {
			g.b, err = c.parseOperand(words[2])
		}
	default:
		return gate{}, fmt.Errorf("malformed gate in %q", line)
	}
	return g, err
}

func parseInput(path string) (circuit, error) {
	lines, err := readLines(path)
	if err != nil {
		return circuit{}, err
	}

	c := circuit{gates: make([]gate, 0, len(lines)), ids: make(map[string]int, 2*len(lines))}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		g, err := c.parseGate(line)
		if err != nil {
			return circuit{}, err
		}
		c.gates = append(c.gates, g)
	}
	if len(c.gates) != len(c.ids) {
		return circuit{}, fmt.Errorf("%s: %d gates drive %d wires; each needs exactly one", path, len(c.gates), len(c.ids))
	}
	return c, nil
}

// evaluate settles the circuit and returns the signal on wire target. If
// override is non-negative, wire overrideWire is fixed to it instead of its
// gate. signal and known are scratch space, one entry per wire.
func (c *circuit) evaluate(target, overrideWire, override int, signal []int, known []bool) (int, error) {
	if len(signal) != len(c.gates) || len(known) != len(c.gates) {
		return 0, errors.New("scratch space does not match the circuit")
	}
	if target < 0 || target >= len(c.gates) {
		return 0, fmt.Errorf("wire %d is not in the circuit", target)
	}

	clear(known)
	if override >= 0 {
		signal[overrideWire], known[overrideWire] = override, true
	}
	for range len(c.gates) + 1 {
		if known[target] {
			return signal[target], nil
		}
		settled := 0
		for _, g := range c.gates {
			if known[g.target] || (g.a.isWire && !known[g.a.value]) || (g.b.isWire && !known[g.b.value]) {
				continue
			}
			value, err := apply(g, signal)
			if err != nil {
				return 0, err
			}
			signal[g.target], known[g.target] = value, true
			settled++
		}
		if settled == 0 {
			return 0, errors.New("circuit cannot settle: a wire depends on itself or is never driven")
		}
	}
	return 0, errors.New("circuit did not settle")
}

// apply computes one gate's output from signals already known.
func apply(g gate, signal []int) (int, error) {
	a, b := g.a.value, g.b.value
	if g.a.isWire {
		a = signal[a]
	}
	if g.b.isWire {
		b = signal[b]
	}
	if a < 0 || a > 0xFFFF || b < 0 || b > 0xFFFF {
		return 0, fmt.Errorf("gate inputs %d and %d are not 16-bit", a, b)
	}

	switch g.op {
	case opAssign:
		return a, nil
	case opAnd:
		return a & b, nil
	case opOr:
		return a | b, nil
	case opLshift:
		return (a << b) & 0xFFFF, nil
	case opRshift:
		return a >> b, nil
	case opNot:
		return ^a & 0xFFFF, nil
	}
	return 0, fmt.Errorf("unknown gate %d", g.op)
}

func day7(path string) error {
	c, err := parseInput(path)
	if err != nil {
		return err
	}
	a, okA := c.ids["a"]
	b, okB := c.ids["b"]
	if !okA || !okB {
		return errors.New("circuit has no wire a or no wire b")
	}

	signal := make([]int, len(c.gates))
	known := make([]bool, len(c.gates))
	first, err := c.evaluate(a, b, -1, signal, known)
	if err != nil {
		return err
	}
	second, err := c.evaluate(a, b, first, signal, known)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
