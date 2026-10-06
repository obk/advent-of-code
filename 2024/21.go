// Day 21: Keypad Conundrum
//
// The input is codes like "029A" to type on a numeric keypad. A robot
// types them, driven from a directional keypad (^ v < > A), which another
// robot presses, driven from another directional keypad, and so on; you
// press the last one. Robots' arms start on A and must never aim at the
// keypads' gaps.
//
// Part 1 uses two robot-held directional keypads; part 2 uses 25. Each
// answer sums, over codes, the presses you make times the code's number.
//
// cost[L][a][b] is how many of your presses it takes to have the robot L
// layers away move its arm from key a to key b and press it (layer 0 is
// you: one press). A move is all its horizontal steps then all vertical,
// or the other way round, whichever avoids the gap and costs less; each is
// priced with the layer below, starting and ending on A.
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
	maxInputBytes = 1 << 8
	maxLayers     = 25
	dirKeys       = "^A<v>"
	numKeys       = "789456123 0A"
)

// pos returns a key's column and row on its keypad (3 keys per row).
func pos(keys string, k byte) (int, int, error) {
	i := strings.IndexByte(keys, k)
	if i < 0 || k == ' ' {
		return 0, 0, fmt.Errorf("no key %q", k)
	}
	if keys == dirKeys {
		i++ // the directional pad's gap is its top-left corner
	}
	return i % 3, i / 3, nil
}

// pathCost prices the cheapest way to move from key a to key b on a keypad
// (with the gap at gx, gy) and press it, using the given costs of the
// directional keys one layer below.
func pathCost(keys string, a, b byte, gx, gy int, below *[5][5]int) (int, error) {
	if below == nil {
		return 0, errors.New("no costs for the layer below")
	}
	if gx < 0 || gx > 2 || gy < 0 || gy > 3 {
		return 0, fmt.Errorf("gap %d,%d is off the keypad", gx, gy)
	}
	ax, ay, errA := pos(keys, a)
	bx, by, errB := pos(keys, b)
	if errA != nil || errB != nil {
		return 0, fmt.Errorf("unknown keys %q and %q", a, b)
	}
	h, v := byte('>'), byte('v')
	if bx < ax {
		h = '<'
	}
	if by < ay {
		v = '^'
	}
	nh, nv := max(bx-ax, ax-bx), max(by-ay, ay-by)
	best := -1
	for _, horizontalFirst := range [2]bool{true, false} {
		if (horizontalFirst && bx == gx && ay == gy) || (!horizontalFirst && ax == gx && by == gy) {
			continue // that corner is the gap
		}
		first, second, n1, n2 := h, v, nh, nv
		if !horizontalFirst {
			first, second, n1, n2 = v, h, nv, nh
		}
		cost, at := 0, strings.IndexByte(dirKeys, 'A')
		for _, step := range [3][2]int{{int(first), n1}, {int(second), n2}, {'A', 1}} {
			k := strings.IndexByte(dirKeys, byte(step[0]))
			for range step[1] {
				cost, at = cost+below[at][k], k
			}
		}
		if best < 0 || cost < best {
			best = cost
		}
	}
	if best < 0 {
		return 0, errors.New("every path crosses the gap")
	}
	return best, nil
}

// layers fills cost for every layer of directional keypads.
func layers(cost *[maxLayers + 1][5][5]int) error {
	if cost == nil {
		return errors.New("no cost table")
	}
	for a := range 5 {
		for b := range 5 {
			cost[0][a][b] = 1
		}
	}
	for l := 1; l <= maxLayers; l++ {
		for a := range 5 {
			for b := range 5 {
				c, err := pathCost(dirKeys, dirKeys[a], dirKeys[b], 0, 0, &cost[l-1])
				if err != nil {
					return err
				}
				cost[l][a][b] = c
			}
			// Pressing the key the arm is on takes one press of A below.
			if cost[l][a][a] != 1 {
				return fmt.Errorf("pressing %c again costs %d at layer %d", dirKeys[a], cost[l][a][a], l)
			}
		}
	}
	return nil
}

// complexity prices typing a code with the given layers of robots.
func complexity(code string, robots int, cost *[maxLayers + 1][5][5]int) (int, error) {
	if robots < 1 || robots > maxLayers || len(code) == 0 || code[len(code)-1] != 'A' {
		return 0, fmt.Errorf("cannot type %q with %d robots", code, robots)
	}
	n, err := strconv.Atoi(strings.TrimLeft(code[:len(code)-1], "0"))
	if err != nil {
		return 0, fmt.Errorf("code %q has no number", code)
	}
	if n < 1 || n > 999 {
		return 0, fmt.Errorf("code %q has number %d", code, n)
	}
	presses, at := 0, byte('A')
	for k := range len(code) {
		c, err := pathCost(numKeys, at, code[k], 0, 3, &cost[robots])
		if err != nil {
			return 0, err
		}
		presses, at = presses+c, code[k]
	}
	return presses * n, nil
}

func day21(path string) error {
	if path == "" {
		return errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	var cost [maxLayers + 1][5][5]int
	if err := layers(&cost); err != nil {
		return err
	}
	codes := strings.Fields(string(content))
	if len(codes) == 0 {
		return fmt.Errorf("%s: no codes", path)
	}
	var answers [2]int
	for _, code := range codes {
		for k, robots := range [2]int{2, maxLayers} {
			c, err := complexity(code, robots, &cost)
			if err != nil {
				return err
			}
			answers[k] += c
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
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
