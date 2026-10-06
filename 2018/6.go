// Day 6: Chronal Coordinates
//
// Each line of the input is a coordinate "x, y". Every location of the grid
// belongs to the coordinate it is strictly closest to by taxicab distance
// (ties belong to none).
//
// Part 1 is the size of the largest area that is not infinite.
// Part 2 counts the locations whose total distance to all coordinates is
// below 10000. A location more than 10000/n beyond the bounding box is
// already too far from every coordinate, so the box widened by that much
// holds the whole region.
//
// Only the bounding box of the coordinates needs examining. An area that
// reaches the edge of the box goes on forever beyond it, so it is
// infinite; every other area lies wholly inside the box.
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
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 16
	maxPoints     = 100
	maxCoord      = 1_000
	safeTotal     = 10_000
)

// point is one coordinate.
type point struct {
	x, y int
}

// readPoints parses the coordinates.
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || len(lines) > maxPoints {
		return nil, fmt.Errorf("%s: %d coordinates, expected 1..%d", path, len(lines), maxPoints)
	}
	points := make([]point, len(lines))
	for i, line := range lines {
		xs, ys, found := strings.Cut(line, ",")
		x, errX := strconv.Atoi(strings.TrimSpace(xs))
		y, errY := strconv.Atoi(strings.TrimSpace(ys))
		if !found || errX != nil || errY != nil || x < 0 || y < 0 || x > maxCoord || y > maxCoord {
			return nil, fmt.Errorf("%s: invalid coordinate %q", path, line)
		}
		points[i] = point{x, y}
	}
	for i := range points {
		for j := i + 1; j < len(points); j++ {
			if points[i] == points[j] {
				return nil, fmt.Errorf("%s: coordinate %v appears twice", path, points[i])
			}
		}
	}
	return points, nil
}

// owner returns the coordinate strictly closest to (x, y), or -1 on a tie,
// along with the total distance to all coordinates.
func owner(points []point, x, y int) (int, int) {
	best, bestDist, tie, total := -1, 0, false, 0
	for i, p := range points {
		d := max(p.x-x, x-p.x) + max(p.y-y, y-p.y)
		total += d
		switch {
		case best < 0 || d < bestDist:
			best, bestDist, tie = i, d, false
		case d == bestDist:
			tie = true
		}
	}
	if tie {
		return -1, total
	}
	return best, total
}

// largestArea measures every area inside the bounding box. area and
// infinite are scratch space, one entry per coordinate.
func largestArea(points []point, area []int, infinite []bool) (int, error) {
	if len(points) == 0 || len(area) != len(points) || len(infinite) != len(points) {
		return 0, errors.New("scratch space does not match the coordinates")
	}
	if len(points) > maxPoints {
		return 0, fmt.Errorf("%d coordinates, more than %d", len(points), maxPoints)
	}
	lo, hi := points[0], points[0]
	for _, p := range points {
		lo = point{min(lo.x, p.x), min(lo.y, p.y)}
		hi = point{max(hi.x, p.x), max(hi.y, p.y)}
	}
	for y := lo.y; y <= hi.y; y++ {
		for x := lo.x; x <= hi.x; x++ {
			o, _ := owner(points, x, y)
			if o < 0 {
				continue
			}
			area[o]++
			if x == lo.x || x == hi.x || y == lo.y || y == hi.y {
				infinite[o] = true
			}
		}
	}
	best := -1
	for i := range points {
		if !infinite[i] {
			best = max(best, area[i])
		}
	}
	if best < 0 {
		return 0, errors.New("every area is infinite")
	}
	if best > (hi.x-lo.x+1)*(hi.y-lo.y+1) {
		return 0, fmt.Errorf("area %d is larger than the bounding box", best)
	}
	return best, nil
}

// safeRegion counts the locations with total distance below limit.
func safeRegion(points []point, limit int) (int, error) {
	if len(points) == 0 || limit < 1 {
		return 0, errors.New("no coordinates or no limit")
	}
	lo, hi := points[0], points[0]
	for _, p := range points {
		lo = point{min(lo.x, p.x), min(lo.y, p.y)}
		hi = point{max(hi.x, p.x), max(hi.y, p.y)}
	}
	margin := limit/len(points) + 1
	if margin > maxCoord*10 {
		return 0, fmt.Errorf("margin %d is too wide to search", margin)
	}
	count := 0
	for y := lo.y - margin; y <= hi.y+margin; y++ {
		for x := lo.x - margin; x <= hi.x+margin; x++ {
			if _, total := owner(points, x, y); total < limit {
				count++
			}
		}
	}
	return count, nil
}

func day6(path string) error {
	points, err := readPoints(path)
	if err != nil {
		return err
	}
	largest, err := largestArea(points, make([]int, len(points)), make([]bool, len(points)))
	if err != nil {
		return err
	}

	region, err := safeRegion(points, safeTotal)
	if err != nil {
		return err
	}
	if largest < 1 {
		return fmt.Errorf("largest area %d is empty", largest)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", largest, region); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
