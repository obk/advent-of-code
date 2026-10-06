// Day 3: Squares With Three Sides
//
// Each line of the input lists three side lengths. A triangle is possible
// when every two sides add up to more than the third.
//
// Part 1 counts the possible triangles, reading each line as one triangle.
// Part 2 reads them down the columns instead: each column, three lines at a
// time, gives one triangle.
//
// Sorting the three sides means only one comparison is needed: the two
// shortest must add up to more than the longest.
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
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxSide       = 1_000_000
)

type triangle [3]int

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

func parseInput(path string) ([]triangle, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	triangles := make([]triangle, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %q does not hold three sides", line)
		}
		var t triangle
		for i, field := range fields {
			if t[i], err = strconv.Atoi(field); err != nil {
				return nil, fmt.Errorf("invalid side in %q: %w", line, err)
			}
			if t[i] < 1 || t[i] > maxSide {
				return nil, fmt.Errorf("side in %q is outside 1..%d", line, maxSide)
			}
		}
		triangles = append(triangles, t)
	}
	if len(triangles) == 0 {
		return nil, fmt.Errorf("%s: no triangles found", path)
	}
	return triangles, nil
}

// possible reports whether three sides can form a triangle.
func possible(t triangle) (bool, error) {
	if t[0] < 1 || t[1] < 1 || t[2] < 1 {
		return false, fmt.Errorf("triangle %v has a side shorter than 1", t)
	}
	if t[0] > maxSide || t[1] > maxSide || t[2] > maxSide {
		return false, fmt.Errorf("triangle %v has a side longer than %d", t, maxSide)
	}
	// Three elements: a fixed sorting network.
	for _, pair := range [3][2]int{{0, 1}, {1, 2}, {0, 1}} {
		if t[pair[0]] > t[pair[1]] {
			t[pair[0]], t[pair[1]] = t[pair[1]], t[pair[0]]
		}
	}
	if t[0] > t[1] || t[1] > t[2] {
		return false, fmt.Errorf("sides %v did not sort", t)
	}
	return t[0]+t[1] > t[2], nil
}

// countPossible counts the possible triangles, by rows or by columns.
func countPossible(rows []triangle, byColumn bool) (int, error) {
	if len(rows) == 0 {
		return 0, errors.New("no triangles")
	}
	if byColumn && len(rows)%3 != 0 {
		return 0, fmt.Errorf("%d rows do not group into threes", len(rows))
	}

	count := 0
	for i := range len(rows) {
		t := rows[i]
		if byColumn {
			// Row i stands for column i%3 of the group starting at i-i%3.
			g, c := i-i%3, i%3
			t = triangle{rows[g][c], rows[g+1][c], rows[g+2][c]}
		}
		ok, err := possible(t)
		if err != nil {
			return 0, err
		}
		if ok {
			count++
		}
	}
	if count > len(rows) {
		return 0, fmt.Errorf("%d possible triangles out of %d", count, len(rows))
	}
	return count, nil
}

func day3(path string) error {
	triangles, err := parseInput(path)
	if err != nil {
		return err
	}
	byRow, err := countPossible(triangles, false)
	if err != nil {
		return err
	}
	byColumn, err := countPossible(triangles, true)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", byRow, byColumn); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
