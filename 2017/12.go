// Day 12: Digital Plumber
//
// Each line of the input lists a program and the programs it has pipes to:
// "2 <-> 0, 3, 4". Pipes work both ways.
//
// Part 1 is the size of the group of programs connected to program 0.
// Part 2 is the number of groups.
//
// A union-find joins the two ends of every pipe; a group's size is then the
// size of its root's set.
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
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
)

// sets is a union-find over the programs.
type sets struct {
	parent []int
	size   []int
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

// find returns i's root, halving the path on the way.
func (s *sets) find(i int) (int, error) {
	if i < 0 || i >= len(s.parent) {
		return 0, fmt.Errorf("program %d does not exist", i)
	}
	for range len(s.parent) {
		if s.parent[i] < 0 || s.parent[i] >= len(s.parent) {
			return 0, fmt.Errorf("program %d has parent %d outside the sets", i, s.parent[i])
		}
		if s.parent[i] == i {
			return i, nil
		}
		s.parent[i] = s.parent[s.parent[i]]
		i = s.parent[i]
	}
	return 0, errors.New("union-find has a cycle")
}

// union joins the groups of a and b.
func (s *sets) union(a, b int) error {
	if len(s.parent) != len(s.size) {
		return errors.New("union-find tables differ in length")
	}
	ra, err := s.find(a)
	if err != nil {
		return err
	}
	rb, err := s.find(b)
	if err != nil {
		return err
	}
	if ra == rb {
		return nil
	}
	if s.size[ra] < 1 || s.size[rb] < 1 {
		return fmt.Errorf("groups %d and %d have no members", ra, rb)
	}
	if s.size[ra] < s.size[rb] {
		ra, rb = rb, ra
	}
	s.parent[rb] = ra
	s.size[ra] += s.size[rb]
	return nil
}

// build reads every line's pipes into a union-find over len(lines) programs.
func build(lines []string) (*sets, error) {
	n := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	if n == 0 {
		return nil, errors.New("no programs")
	}
	s := &sets{parent: make([]int, n), size: make([]int, n)}
	for i := range n {
		s.parent[i], s.size[i] = i, 1
	}
	next := 0 // programs are listed in order 0, 1, 2, ...
	for _, line := range lines {
		from, list, found := strings.Cut(strings.TrimSpace(line), " <-> ")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if from != strconv.Itoa(next) {
			return nil, fmt.Errorf("line %q is out of order; expected program %d", line, next)
		}
		next++
		if !found {
			return nil, fmt.Errorf("malformed line %q", line)
		}
		a, err := strconv.Atoi(from)
		if err != nil {
			return nil, fmt.Errorf("invalid program in %q: %w", line, err)
		}
		for _, to := range strings.Split(list, ", ") {
			b, err := strconv.Atoi(to)
			if err != nil {
				return nil, fmt.Errorf("invalid pipe in %q: %w", line, err)
			}
			if err := s.union(a, b); err != nil {
				return nil, err
			}
		}
	}
	return s, nil
}

func day12(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	s, err := build(lines)
	if err != nil {
		return err
	}
	root, err := s.find(0)
	if err != nil {
		return err
	}

	groups := 0
	for i := range s.parent {
		if s.parent[i] == i {
			groups++
		}
	}
	if groups < 1 || groups > len(s.parent) {
		return fmt.Errorf("%d groups among %d programs", groups, len(s.parent))
	}
	if s.size[root] < 1 {
		return errors.New("program 0's group is empty")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", s.size[root], groups); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
