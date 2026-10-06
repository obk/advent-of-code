// Day 2: I Was Told There Would Be No Math
//
// Each line of the input is a present's dimensions, "LxWxH".
//
// Part 1 is the wrapping paper needed: each box's surface area plus the area
// of its smallest side as slack.
// Part 2 is the ribbon needed: the smallest perimeter of any one face, plus
// the box's volume for the bow.
//
// Sorting each box's three sides makes the smallest face simply the product
// (or perimeter) of the two shortest.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxSide       = 10_000 // keeps every area and volume well inside int
)

// box is a present's three side lengths, shortest first.
type box [3]int

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

// parseBox reads dimensions written as "2x3x4" and sorts them.
func parseBox(line string) (box, error) {
	fields := strings.Split(line, "x")
	if len(fields) != 3 {
		return box{}, fmt.Errorf("malformed dimensions %q", line)
	}

	var b box
	for i, field := range fields {
		side, err := strconv.Atoi(field)
		if err != nil {
			return box{}, fmt.Errorf("invalid dimensions %q: %w", line, err)
		}
		if side < 1 || side > maxSide {
			return box{}, fmt.Errorf("side in %q is outside 1..%d", line, maxSide)
		}
		b[i] = side
	}

	// Three elements: a fixed sorting network.
	for _, pair := range [3][2]int{{0, 1}, {1, 2}, {0, 1}} {
		if b[pair[0]] > b[pair[1]] {
			b[pair[0]], b[pair[1]] = b[pair[1]], b[pair[0]]
		}
	}
	if b[0] > b[1] || b[1] > b[2] {
		return box{}, fmt.Errorf("sides of %q did not sort", line)
	}
	return b, nil
}

func parseInput(path string) ([]box, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	boxes := make([]box, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b, err := parseBox(line)
		if err != nil {
			return nil, err
		}
		boxes = append(boxes, b)
	}
	if len(boxes) == 0 {
		return nil, fmt.Errorf("%s: no presents found", path)
	}
	return boxes, nil
}

// paperAndRibbon returns what one sorted box needs of each.
func paperAndRibbon(b box) (paper, ribbon int, err error) {
	if b[0] < 1 || b[2] > maxSide {
		return 0, 0, fmt.Errorf("box %v is outside 1..%d", b, maxSide)
	}
	if b[0] > b[1] || b[1] > b[2] {
		return 0, 0, fmt.Errorf("box %v is not sorted", b)
	}

	l, w, h := b[0], b[1], b[2]
	paper = 2*l*w + 2*w*h + 2*h*l + l*w
	ribbon = 2*l + 2*w + l*w*h
	// A 1x1x1 box is the smallest possible: 7 square feet and 5 feet.
	if paper < 7 {
		return 0, 0, fmt.Errorf("box %v needs only %d square feet of paper", b, paper)
	}
	if ribbon < 5 {
		return 0, 0, fmt.Errorf("box %v needs only %d feet of ribbon", b, ribbon)
	}
	return paper, ribbon, nil
}

func day2(path string) error {
	boxes, err := parseInput(path)
	if err != nil {
		return err
	}

	totalPaper, totalRibbon := 0, 0
	for _, b := range boxes {
		paper, ribbon, err := paperAndRibbon(b)
		if err != nil {
			return err
		}
		totalPaper += paper
		totalRibbon += ribbon
	}

	if totalPaper < 7*len(boxes) {
		return fmt.Errorf("%d square feet is too little paper for %d presents", totalPaper, len(boxes))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", totalPaper, totalRibbon); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
