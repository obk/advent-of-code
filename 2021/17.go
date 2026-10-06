// Day 17: Trick Shot
//
// The input is a target area "target area: x=20..30, y=-10..-5". A probe
// starts at 0,0 with an integer velocity; each step it moves by its
// velocity, drag brings x velocity towards 0 and gravity lowers y velocity
// by 1.
//
// Part 1 is the highest y any probe reaches while still ending up in the
// target after some step.
// Part 2 counts the initial velocities that hit the target.
//
// The target lies right of and below the start, so x velocities beyond
// its far edge and y velocities beyond its depth (either way) overshoot on
// the first step, or on the way back down past y=0. That bounds the search.
package main

import (
	"errors"
	"fmt"
	"os"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 8
	maxCoord      = 1_000
)

// target is the area to hit.
type target struct {
	x1, x2, y1, y2 int
}

// fly launches a probe and reports whether it hits, and its peak.
func fly(t target, vx, vy int) (hit bool, peak int, err error) {
	if vx < 1 || vx > t.x2 || vy < t.y1 || vy > -t.y1 {
		return false, 0, fmt.Errorf("velocity %d,%d is outside the search", vx, vy)
	}
	x, y := 0, 0
	for step := 0; step < 4*maxCoord && x <= t.x2 && y >= t.y1; step++ {
		x, y = x+vx, y+vy
		vx, vy = max(vx-1, 0), vy-1
		peak = max(peak, y)
		if x >= t.x1 && x <= t.x2 && y >= t.y1 && y <= t.y2 {
			return true, peak, nil
		}
	}
	return false, 0, nil
}

func day17(path string) error {
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
	var t target
	n, err := fmt.Sscanf(string(content), "target area: x=%d..%d, y=%d..%d", &t.x1, &t.x2, &t.y1, &t.y2)
	if err != nil || n != 4 {
		return fmt.Errorf("%s: malformed target", path)
	}
	if t.x1 < 1 || t.x2 < t.x1 || t.x2 > maxCoord || t.y2 >= 0 || t.y1 > t.y2 || t.y1 < -maxCoord {
		return fmt.Errorf("%s: target must be right of and below the start", path)
	}
	best, count := -1, 0
	for vx := 1; vx <= t.x2; vx++ {
		for vy := t.y1; vy <= -t.y1; vy++ {
			hit, peak, err := fly(t, vx, vy)
			if err != nil {
				return err
			}
			if hit {
				best, count = max(best, peak), count+1
			}
		}
	}
	if count == 0 {
		return errors.New("no velocity hits the target")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", best, count); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
