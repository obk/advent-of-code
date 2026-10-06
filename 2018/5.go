// Day 5: Alchemical Reduction
//
// The input is a polymer: a string of letters. Two adjacent units of the
// same letter in opposite cases ("aA" or "Bb") react and both vanish, which
// may bring new pairs together.
//
// Part 1 is the length of the polymer once nothing more can react.
// Part 2 removes every unit of one letter (both cases) first, choosing the
// letter that leaves the shortest result, and gives that length.
//
// A stack does it in one pass: each unit either cancels the unit on top of
// the stack or is pushed onto it.
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
	maxInputBytes = 1 << 20
)

// readPolymer returns the polymer.
func readPolymer(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	polymer := strings.TrimSpace(string(content))
	if polymer == "" || strings.Trim(polymer, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return "", fmt.Errorf("%s: polymer must be letters only", path)
	}
	return polymer, nil
}

// react returns the length of the fully reacted polymer, skipping units of
// the letter skip (0 to skip none). stack is scratch space.
func react(polymer string, skip byte, stack []byte) (int, error) {
	if cap(stack) < len(polymer) {
		return 0, fmt.Errorf("stack of %d cannot hold %d units", cap(stack), len(polymer))
	}
	if skip != 0 && (skip < 'a' || skip > 'z') {
		return 0, fmt.Errorf("cannot skip %q", skip)
	}
	stack = stack[:0]
	kept := 0
	for i := range len(polymer) {
		u := polymer[i]
		if u|0x20 == skip {
			continue
		}
		kept++
		if n := len(stack); n > 0 && stack[n-1] != u && stack[n-1]|0x20 == u|0x20 {
			stack = stack[:n-1]
			continue
		}
		stack = append(stack, u)
	}
	if len(stack) > len(polymer) {
		return 0, errors.New("the polymer grew")
	}
	// Every reaction removes two units.
	if (kept-len(stack))%2 != 0 {
		return 0, fmt.Errorf("%d units reacted, an odd number", kept-len(stack))
	}
	return len(stack), nil
}

func day5(path string) error {
	polymer, err := readPolymer(path)
	if err != nil {
		return err
	}
	stack := make([]byte, 0, len(polymer))
	length, err := react(polymer, 0, stack)
	if err != nil {
		return err
	}

	shortest := length
	for letter := byte('a'); letter <= 'z'; letter++ {
		n, err := react(polymer, letter, stack)
		if err != nil {
			return err
		}
		shortest = min(shortest, n)
	}
	if shortest > length {
		return fmt.Errorf("removing a letter made the polymer longer (%d > %d)", shortest, length)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", length, shortest); err != nil {
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
