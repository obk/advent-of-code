// Day 5: Supply Stacks
//
// The input is a drawing of crate stacks, then moves "move 3 from 1 to 2".
//
// Part 1 moves crates one at a time, and lists the top crate of each stack.
// Part 2 moves the crates of each step together, keeping their order.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 16
	maxStacks     = 9
	maxCrates     = 128
)

// yard is the stacks, bottom first.
type yard struct {
	stack  [maxStacks][maxCrates]byte
	height [maxStacks]int
	n      int
}

// move is one rearrangement step (stacks counted from 0).
type move struct {
	count, from, to int
}

// load parses the drawing and the moves.
func load(path string) (yard, []move, error) {
	var y yard
	if path == "" {
		return y, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return y, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return y, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	drawing, moveText, found := strings.Cut(strings.TrimRight(string(content), "\n"), "\n\n")
	rows := strings.Split(drawing, "\n")
	if !found || len(rows) < 2 || len(rows) > maxCrates {
		return y, nil, fmt.Errorf("%s: malformed drawing", path)
	}
	y.n = len(strings.Fields(rows[len(rows)-1]))
	if y.n < 1 || y.n > maxStacks {
		return y, nil, fmt.Errorf("%s: %d stacks, expected 1..%d", path, y.n, maxStacks)
	}
	for r := len(rows) - 2; r >= 0; r-- {
		for s := range y.n {
			if c := 1 + 4*s; c < len(rows[r]) && rows[r][c] != ' ' {
				y.stack[s][y.height[s]], y.height[s] = rows[r][c], y.height[s]+1
			}
		}
	}
	lines := strings.Split(moveText, "\n")
	moves := make([]move, len(lines))
	for i, l := range lines {
		m := &moves[i]
		k, err := fmt.Sscanf(l, "move %d from %d to %d", &m.count, &m.from, &m.to)
		m.from, m.to = m.from-1, m.to-1
		if err != nil || k != 3 || m.count < 1 || m.from < 0 || m.to < 0 || m.from >= y.n || m.to >= y.n || m.from == m.to {
			return y, nil, fmt.Errorf("%s: invalid move %q", path, l)
		}
	}
	return y, moves, nil
}

// apply runs the moves on a copy; together keeps each step's crates in
// order. It returns the top crates.
func apply(y yard, moves []move, together bool, out []byte) ([]byte, error) {
	if cap(out) < y.n {
		return nil, errors.New("no room for the top crates")
	}
	for _, m := range moves {
		if m.count > y.height[m.from] || y.height[m.to]+m.count > maxCrates {
			return nil, fmt.Errorf("cannot move %d crates from stack %d", m.count, m.from+1)
		}
		for k := range m.count {
			src := y.height[m.from] - 1 - k // one at a time: the top first
			if together {
				src = y.height[m.from] - m.count + k
			}
			y.stack[m.to][y.height[m.to]+k] = y.stack[m.from][src]
		}
		y.height[m.from] -= m.count
		y.height[m.to] += m.count
	}
	out = out[:0]
	for s := range y.n {
		if y.height[s] == 0 {
			return nil, fmt.Errorf("stack %d ended empty", s+1)
		}
		out = append(out, y.stack[s][y.height[s]-1])
	}
	return out, nil
}

func day5(path string) error {
	y, moves, err := load(path)
	if err != nil {
		return err
	}
	single, err := apply(y, moves, false, make([]byte, 0, maxStacks))
	if err != nil {
		return err
	}
	multi, err := apply(y, moves, true, make([]byte, 0, maxStacks))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", single, multi); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
