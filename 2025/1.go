// Day 1: Secret Entrance
//
// The input is one dial rotation per line, written as a direction and a click
// count: "R23" turns clockwise 23 clicks, "L5" anticlockwise 5. The dial has
// positions 0..99, wraps around at either end, and starts at 50.
//
// Part 1 counts the rotations that come to rest on 0.
// Part 2 counts every time the dial reaches 0, including the times it merely
// sweeps past mid-rotation.
//
// Only part 2 needs the dial walked click by click, and even then not all of
// it: a rotation of n clicks makes n/100 full revolutions, each passing 0
// exactly once, so only the n%100 leftover clicks have to be simulated.
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
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
	maxLines      = 50_000
	maxClicks     = 1_000_000
)

// The dial has positions 0..99 and starts pointing at 50.
const (
	dialPositions = 100
	dialStart     = 50
)

// rotation is one instruction: which way to turn the dial, and by how many
// clicks.
type rotation struct {
	clockwise bool
	clicks    int
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

// parseRotation reads an instruction such as "R23" - a direction character
// followed by a click count.
func parseRotation(line string) (rotation, error) {
	if len(line) < 2 {
		return rotation{}, fmt.Errorf("malformed instruction %q", line)
	}

	clicks, err := strconv.Atoi(line[1:])
	if err != nil {
		return rotation{}, fmt.Errorf("invalid click count in %q: %w", line, err)
	}
	if clicks < 0 || clicks > maxClicks {
		return rotation{}, fmt.Errorf("click count in %q is outside 0..%d", line, maxClicks)
	}

	if line[0] != 'R' && line[0] != 'L' {
		return rotation{}, fmt.Errorf("unknown direction in %q", line)
	}
	return rotation{clockwise: line[0] == 'R', clicks: clicks}, nil
}

func parseInput(path string) ([]rotation, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	// Sized up front, so the appends below never reallocate.
	rotations := make([]rotation, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r, err := parseRotation(line)
		if err != nil {
			return nil, err
		}
		rotations = append(rotations, r)
	}
	if len(rotations) == 0 {
		return nil, fmt.Errorf("%s: no rotations found", path)
	}
	return rotations, nil
}

// turn applies one rotation and reports where the dial ends up together with
// how many times it landed on 0 along the way.
//
// Each full revolution passes 0 exactly once, so only the clicks left over
// after the whole revolutions have to be walked one at a time.
func turn(position int, r rotation) (end, zeroTouches int, err error) {
	if position < 0 || position >= dialPositions {
		return 0, 0, fmt.Errorf("dial position %d is outside 0..%d", position, dialPositions-1)
	}
	if r.clicks < 0 || r.clicks > maxClicks {
		return 0, 0, fmt.Errorf("click count %d is outside 0..%d", r.clicks, maxClicks)
	}

	zeroTouches = r.clicks / dialPositions
	for range r.clicks % dialPositions {
		if r.clockwise {
			position = (position + 1) % dialPositions
		} else {
			position = (position + dialPositions - 1) % dialPositions
		}
		if position == 0 {
			zeroTouches++
		}
	}
	// The leftover clicks are fewer than a full revolution, so they can reach
	// 0 at most once more.
	if zeroTouches > r.clicks/dialPositions+1 {
		return 0, 0, fmt.Errorf("%d clicks reached 0 %d times", r.clicks, zeroTouches)
	}
	return position, zeroTouches, nil
}

func day1(path string) error {
	rotations, err := parseInput(path)
	if err != nil {
		return err
	}

	position := dialStart
	stopsOnZero := 0 // part 1: rotations that finish on 0
	touchesZero := 0 // part 2: every time the dial passes or lands on 0

	for _, r := range rotations {
		var touches int
		position, touches, err = turn(position, r)
		if err != nil {
			return err
		}
		touchesZero += touches

		if position == 0 {
			stopsOnZero++
		}
	}

	if stopsOnZero > len(rotations) {
		return fmt.Errorf("%d rotations stopped on 0 %d times", len(rotations), stopsOnZero)
	}
	if stopsOnZero > touchesZero {
		return fmt.Errorf("landed on 0 %d times but only reached it %d times", stopsOnZero, touchesZero)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", stopsOnZero, touchesZero); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day1(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
