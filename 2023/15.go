// Day 15: Lens Library
//
// The input is comma-separated steps. HASH of a string: for each byte,
// add its code, multiply by 17, keep the remainder mod 256.
//
// Part 1 sums the HASH of every step.
// Part 2 runs the steps on 256 boxes: "label=f" puts a lens of focal
// length f in box HASH(label) (replacing one with that label, else at the
// back); "label-" removes it. The answer sums box+1 times slot times focal
// length over all lenses.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 16
	boxes         = 256
	maxLenses     = 64
)

// hash is the HASH algorithm.
func hash(s string) int {
	h := 0
	for k := range len(s) {
		h = (h + int(s[k])) * 17 % 256
	}
	return h
}

// box is a row of lenses.
type box struct {
	label [maxLenses]string
	focal [maxLenses]int
	n     int
}

// apply runs one step on the boxes.
func apply(b *[boxes]box, step string) error {
	if step == "" || b == nil {
		return errors.New("empty step or no boxes")
	}
	label, focal := strings.TrimSuffix(step, "-"), -1
	if i := strings.IndexByte(step, '='); i >= 0 {
		label, focal = step[:i], int(step[len(step)-1]-'0')
		if i != len(step)-2 || focal < 1 || focal > 9 {
			return fmt.Errorf("malformed step %q", step)
		}
	}
	if label == "" || strings.Trim(label, "abcdefghijklmnopqrstuvwxyz") != "" {
		return fmt.Errorf("step %q has no lowercase label", step)
	}
	bx := &b[hash(label)]
	slot := -1
	for k := range bx.n {
		if bx.label[k] == label {
			slot = k
		}
	}
	switch {
	case focal < 0 && slot >= 0:
		copy(bx.label[slot:bx.n-1], bx.label[slot+1:bx.n])
		copy(bx.focal[slot:bx.n-1], bx.focal[slot+1:bx.n])
		bx.n--
	case focal > 0 && slot >= 0:
		bx.focal[slot] = focal
	case focal > 0:
		if bx.n == maxLenses {
			return errors.New("a box is full")
		}
		bx.label[bx.n], bx.focal[bx.n], bx.n = label, focal, bx.n+1
	}
	return nil
}

func day15(path string) error {
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
	steps := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(steps) > maxInputBytes/2 {
		return fmt.Errorf("%s: too many steps", path)
	}
	b := new([boxes]box)
	sum := 0
	for _, s := range steps {
		sum += hash(s)
		if err := apply(b, s); err != nil {
			return err
		}
	}
	power := 0
	for i := range b {
		for k := range b[i].n {
			power += (i + 1) * (k + 1) * b[i].focal[k]
		}
	}

	if sum < 1 {
		return errors.New("every step hashes to 0")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, power); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
