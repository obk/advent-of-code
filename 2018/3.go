// Day 3: No Matter How You Slice It
//
// Each line of the input is a claim on a square of fabric:
// "#123 @ 3,2: 5x4" - a rectangle 3 from the left, 2 from the top, 5 wide
// and 4 tall.
//
// Part 1 counts the square inches inside two or more claims.
// Part 2 is the ID of the only claim that overlaps no other.
//
// The fabric is a fixed 1000x1000 array of claim counts.
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
	side          = 1000
)

// claim is one rectangle.
type claim struct {
	id, x, y, w, h int
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

// parseClaim reads "#123 @ 3,2: 5x4".
func parseClaim(line string) (claim, error) {
	f := strings.Fields(strings.NewReplacer("#", " ", "@", " ", ",", " ", ":", " ", "x", " ").Replace(line))
	if len(f) != 5 {
		return claim{}, fmt.Errorf("malformed claim %q", line)
	}
	var n [5]int
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return claim{}, fmt.Errorf("invalid number in %q: %w", line, err)
		}
		n[i] = v
	}
	if n[0] < 1 {
		return claim{}, fmt.Errorf("claim ID in %q must be positive", line)
	}
	c := claim{n[0], n[1], n[2], n[3], n[4]}
	if c.x < 0 || c.y < 0 || c.w < 1 || c.h < 1 || c.x+c.w > side || c.y+c.h > side {
		return claim{}, fmt.Errorf("claim %q does not fit on the fabric", line)
	}
	return c, nil
}

// overlaps lays every claim on the fabric and counts the square inches
// claimed more than once.
func overlaps(claims []claim, fabric *[side * side]int) (int, error) {
	if len(claims) == 0 || fabric == nil {
		return 0, errors.New("no claims or no fabric")
	}
	for _, c := range claims {
		if c.x+c.w > side || c.y+c.h > side {
			return 0, fmt.Errorf("claim #%d does not fit on the fabric", c.id)
		}
		for r := c.y; r < c.y+c.h; r++ {
			for col := c.x; col < c.x+c.w; col++ {
				fabric[r*side+col]++
			}
		}
	}
	shared := 0
	for _, n := range fabric {
		if n >= 2 {
			shared++
		}
	}
	if shared > side*side {
		return 0, fmt.Errorf("%d shared inches on a fabric of %d", shared, side*side)
	}
	return shared, nil
}

// intact returns the ID of the one claim whose every square inch is claimed
// only once, after overlaps has filled the fabric.
func intact(claims []claim, fabric *[side * side]int) (int, error) {
	if fabric == nil {
		return 0, errors.New("no fabric")
	}
	if len(claims) == 0 {
		return 0, errors.New("no claims")
	}
	found, count := 0, 0
	for _, c := range claims {
		alone := true
		for r := c.y; r < c.y+c.h && alone; r++ {
			for col := c.x; col < c.x+c.w && alone; col++ {
				alone = fabric[r*side+col] == 1
			}
		}
		if alone {
			found, count = c.id, count+1
		}
	}
	if count != 1 {
		return 0, fmt.Errorf("%d claims overlap nothing, expected 1", count)
	}
	return found, nil
}

func day3(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	claims := make([]claim, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		c, err := parseClaim(line)
		if err != nil {
			return err
		}
		claims = append(claims, c)
	}
	fabric := new([side * side]int)
	shared, err := overlaps(claims, fabric)
	if err != nil {
		return err
	}

	id, err := intact(claims, fabric)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", shared, id); err != nil {
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
