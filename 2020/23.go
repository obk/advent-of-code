// Day 23: Crab Cups
//
// The input is the labels of cups in a circle. Each move, the crab picks
// up the three cups after the current one, places them after the cup
// labelled one less than the current (skipping picked-up labels, wrapping
// to the highest), and moves to the next cup.
//
// Part 1 runs 100 moves on the nine cups and lists the labels after cup 1.
// Part 2 adds cups up to one million, runs ten million moves, and
// multiplies the two labels after cup 1.
//
// The circle is an array next[label] = the label clockwise of it, so each
// move rewires three links.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 6
	labels        = 9
	million       = 1_000_000
)

// readLabels parses the nine digits, which must be 1..9 once each.
func readLabels(path string) ([labels]int, error) {
	var out [labels]int
	if path == "" {
		return out, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return out, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	text := strings.TrimSpace(string(content))
	var seen [labels + 1]bool
	if len(text) != labels {
		return out, fmt.Errorf("%s: %d cups, expected %d", path, len(text), labels)
	}
	for i := range labels {
		d := int(text[i] - '0')
		if d < 1 || d > labels || seen[d] {
			return out, fmt.Errorf("%s: cups are not 1..9 once each", path)
		}
		seen[d], out[i] = true, d
	}
	return out, nil
}

// play links the cups (the given labels, then labels+1..total) in next
// and runs the moves.
func play(start [labels]int, total, moves int, next []int32) error {
	if total < labels || len(next) != total+1 || moves < 0 {
		return fmt.Errorf("cannot play %d moves with %d cups", moves, total)
	}
	prev := start[0]
	for _, c := range start[1:] {
		next[prev], prev = int32(c), c
	}
	for c := labels + 1; c <= total; c++ {
		next[prev], prev = int32(c), c
	}
	next[prev] = int32(start[0])
	cur := start[0]
	for range moves {
		a := int(next[cur])
		b := int(next[a])
		c := int(next[b])
		next[cur] = next[c]
		dest := cur
		for k := 0; k < 4 && (dest == cur || dest == a || dest == b || dest == c); k++ {
			dest--
			if dest == 0 {
				dest = total
			}
		}
		if dest == a || dest == b || dest == c || dest == cur {
			return errors.New("no destination cup")
		}
		next[c], next[dest] = next[dest], int32(a)
		cur = int(next[cur])
	}
	// The cups must still form one circle of every label.
	c := 1
	for k := 1; k < total; k++ {
		c = int(next[c])
		if c == 1 {
			return fmt.Errorf("the circle closed after %d cups", k)
		}
	}
	if next[c] != 1 {
		return errors.New("the circle does not close")
	}
	return nil
}

func day23(path string) error {
	start, err := readLabels(path)
	if err != nil {
		return err
	}
	small := make([]int32, labels+1)
	if err := play(start, labels, 100, small); err != nil {
		return err
	}
	var order [labels - 1]byte
	for i, c := 0, small[1]; i < labels-1; i, c = i+1, small[c] {
		order[i] = byte('0' + c)
	}
	big := make([]int32, million+1)
	if err := play(start, million, 10*million, big); err != nil {
		return err
	}
	product := int(big[1]) * int(big[big[1]])
	if product < 1 {
		return fmt.Errorf("product %d is not positive", product)
	}

	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", order[:], product); err != nil {
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
