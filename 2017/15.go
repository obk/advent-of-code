// Day 15: Dueling Generators
//
// Two generators each multiply their previous value by a factor (A: 16807,
// B: 48271) modulo 2147483647. The input gives their starting values. A judge
// compares the lowest 16 bits of each pair of values.
//
// Part 1 counts the matching pairs among the first 40 million.
// Part 2 has A hand over only multiples of 4 and B only multiples of 8, and
// counts the matches among the first 5 million pairs handed over.
//
// The products fit in 64 bits, so plain integer arithmetic does.
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
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 10
	modulus       = 2147483647
	factorA       = 16807
	factorB       = 48271
	part1Pairs    = 40_000_000
	part2Pairs    = 5_000_000
)

// readStarts returns the two starting values.
func readStarts(path string) (a, b int, err error) {
	if path == "" {
		return 0, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		return 0, 0, fmt.Errorf("%s: %d lines, expected 2", path, len(lines))
	}
	var starts [2]int
	for i, line := range lines {
		f := strings.Fields(line)
		if len(f) == 0 {
			return 0, 0, fmt.Errorf("%s: empty line", path)
		}
		if starts[i], err = strconv.Atoi(f[len(f)-1]); err != nil {
			return 0, 0, fmt.Errorf("%s: invalid start in %q: %w", path, line, err)
		}
		if starts[i] < 1 || starts[i] >= modulus {
			return 0, 0, fmt.Errorf("%s: start %d is outside 1..%d", path, starts[i], modulus-1)
		}
	}
	return starts[0], starts[1], nil
}

// judge counts matching low 16 bits over the given number of pairs.
func judge(a, b, pairs int) (int, error) {
	if a < 1 || b < 1 || a >= modulus || b >= modulus {
		return 0, fmt.Errorf("starts %d and %d are outside 1..%d", a, b, modulus-1)
	}
	if pairs < 0 || pairs > part1Pairs {
		return 0, fmt.Errorf("%d pairs is outside 0..%d", pairs, part1Pairs)
	}
	matches := 0
	for range pairs {
		a = a * factorA % modulus
		b = b * factorB % modulus
		if a&0xFFFF == b&0xFFFF {
			matches++
		}
	}
	return matches, nil
}

// next steps a generator until its value is a multiple of m (at most a
// little over m*30 steps in practice; capped well beyond that).
func next(v, factor, m int) (int, error) {
	if m < 1 || v < 1 {
		return 0, fmt.Errorf("cannot step %d to a multiple of %d", v, m)
	}
	if factor != factorA && factor != factorB {
		return 0, fmt.Errorf("unknown generator factor %d", factor)
	}
	for range 1 << 20 {
		v = v * factor % modulus
		if v%m == 0 {
			return v, nil
		}
	}
	return 0, fmt.Errorf("no multiple of %d within 2^20 steps", m)
}

// pickyJudge counts matches when each generator only hands over multiples.
func pickyJudge(a, b, pairs int) (int, error) {
	if pairs < 0 || pairs > part2Pairs {
		return 0, fmt.Errorf("%d pairs is outside 0..%d", pairs, part2Pairs)
	}
	if a < 1 || b < 1 || a >= modulus || b >= modulus {
		return 0, fmt.Errorf("starts %d and %d are outside 1..%d", a, b, modulus-1)
	}
	matches := 0
	var err error
	for range pairs {
		if a, err = next(a, factorA, 4); err != nil {
			return 0, err
		}
		if b, err = next(b, factorB, 8); err != nil {
			return 0, err
		}
		if a&0xFFFF == b&0xFFFF {
			matches++
		}
	}
	return matches, nil
}

func day15(path string) error {
	a, b, err := readStarts(path)
	if err != nil {
		return err
	}
	matches, err := judge(a, b, part1Pairs)
	if err != nil {
		return err
	}

	picky, err := pickyJudge(a, b, part2Pairs)
	if err != nil {
		return err
	}
	if matches > part1Pairs || picky > part2Pairs {
		return fmt.Errorf("more matches (%d, %d) than pairs", matches, picky)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", matches, picky); err != nil {
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
