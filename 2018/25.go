// Day 25: Four-Dimensional Adventure
//
// Each line of the input is a point in four dimensions, "x,y,z,w". Points
// within taxicab distance 3 of each other belong to the same constellation,
// and so does anything chained to them that way.
//
// Part 1 is the number of constellations. (Day 25 has no second puzzle.)
//
// A union-find joins every close pair; the constellations are its roots.
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
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 20
	maxPoints     = 5_000
	maxCoord      = 1_000_000
	joinDistance  = 3
)

// point is one fixed point.
type point [4]int

// readPoints parses the points.
func readPoints(path string) ([]point, error) {
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
	lines := strings.Fields(string(content))
	if len(lines) == 0 || len(lines) > maxPoints {
		return nil, fmt.Errorf("%s: %d points, expected 1..%d", path, len(lines), maxPoints)
	}
	points := make([]point, len(lines))
	for i, line := range lines {
		f := strings.Split(line, ",")
		if len(f) != 4 {
			return nil, fmt.Errorf("%s: malformed point %q", path, line)
		}
		for k, s := range f {
			if points[i][k], err = strconv.Atoi(s); err != nil {
				return nil, fmt.Errorf("%s: invalid number in %q: %w", path, line, err)
			}
			if points[i][k] < -maxCoord || points[i][k] > maxCoord {
				return nil, fmt.Errorf("%s: coordinate in %q is out of range", path, line)
			}
		}
	}
	return points, nil
}

// find returns i's root, halving the path on the way.
func find(parent []int, i int) (int, error) {
	if i < 0 || i >= len(parent) {
		return 0, fmt.Errorf("point %d does not exist", i)
	}
	for range len(parent) {
		if parent[i] == i {
			return i, nil
		}
		parent[i] = parent[parent[i]]
		i = parent[i]
		if i < 0 || i >= len(parent) {
			return 0, fmt.Errorf("parent %d is outside the points", i)
		}
	}
	return 0, errors.New("union-find has a cycle")
}

// constellations joins every pair within joinDistance and counts the roots.
func constellations(points []point, parent []int) (int, error) {
	if len(parent) != len(points) || len(points) == 0 {
		return 0, errors.New("scratch space does not match the points")
	}
	if len(points) > maxPoints {
		return 0, fmt.Errorf("%d points, more than %d", len(points), maxPoints)
	}
	for i := range parent {
		parent[i] = i
	}
	for i := range points {
		for j := i + 1; j < len(points); j++ {
			d := 0
			for k := range 4 {
				d += max(points[i][k]-points[j][k], points[j][k]-points[i][k])
			}
			if d > joinDistance {
				continue
			}
			ri, err := find(parent, i)
			if err != nil {
				return 0, err
			}
			rj, err := find(parent, j)
			if err != nil {
				return 0, err
			}
			parent[ri] = rj
		}
	}
	count := 0
	for i, p := range parent {
		if p == i {
			count++
		}
	}
	return count, nil
}

func day25(path string) error {
	points, err := readPoints(path)
	if err != nil {
		return err
	}
	count, err := constellations(points, make([]int, len(points)))
	if err != nil {
		return err
	}

	if count < 1 || count > len(points) {
		return fmt.Errorf("%d constellations among %d points", count, len(points))
	}
	if _, err := fmt.Printf("Part 1: %d\n", count); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
