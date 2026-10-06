// Day 9: Movie Theater
//
// The input lists the corners of one closed polygon in order, as "x,y". The
// walls between consecutive corners are always axis-aligned. Any two corners
// span a rectangle, whose area counts both end cells: (|dx|+1) * (|dy|+1).
//
// Part 1 takes the largest rectangle spanned by any two corners, wherever it
// happens to lie.
// Part 2 only accepts rectangles that lie wholly inside the polygon.
//
// A rectangle is inside when two things hold: no wall crosses its interior,
// and its centre is inside the outline. The first is a direct check against
// every wall; the second is ray casting, counting how often a ray drawn to the
// right crosses the outline, which is odd exactly for points inside. Neither
// check alone is enough - a rectangle can clear every wall while sitting
// entirely outside the polygon, or contain the centre while a wall slices
// through it. The centre can fall halfway between two cells, so the ray
// casting works on doubled coordinates, which keeps it in exact integers.
//
// Both parts are quadratic in the number of corners. Part 2 keeps that
// affordable by testing the cheap area first and only paying for the
// containment check when a rectangle would actually beat the current best.
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
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxCoord      = 1_000_000 // keeps every area well inside int
)

// point is one corner of the polygon.
type point struct {
	x, y int
}

// edge is one wall of the polygon, always axis-aligned.
type edge struct {
	from, to point
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

// parseCorner reads a corner written as "x,y".
func parseCorner(line string) (point, error) {
	xText, yText, found := strings.Cut(line, ",")
	if !found {
		return point{}, fmt.Errorf("malformed coordinate %q", line)
	}
	x, err := strconv.Atoi(strings.TrimSpace(xText))
	if err != nil {
		return point{}, fmt.Errorf("invalid coordinate %q: %w", line, err)
	}
	y, err := strconv.Atoi(strings.TrimSpace(yText))
	if err != nil {
		return point{}, fmt.Errorf("invalid coordinate %q: %w", line, err)
	}
	if x < 0 || x > maxCoord || y < 0 || y > maxCoord {
		return point{}, fmt.Errorf("coordinate %q is outside 0..%d", line, maxCoord)
	}
	return point{x: x, y: y}, nil
}

func parseInput(path string) ([]point, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	corners := make([]point, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p, err := parseCorner(line)
		if err != nil {
			return nil, err
		}
		corners = append(corners, p)
	}

	// Walls alternate between horizontal and vertical, so a closed outline
	// has an even number of corners, and at least four.
	if len(corners) < 4 || len(corners)%2 != 0 {
		return nil, fmt.Errorf("%s: %d corners cannot close an axis-aligned polygon", path, len(corners))
	}
	return corners, nil
}

// polygonEdges walks the corners in order, closing the loop at the end, and
// writes the walls into edges, which must have one slot per corner.
func polygonEdges(corners []point, edges []edge) error {
	if len(corners) < 4 {
		return fmt.Errorf("%d corners cannot close a polygon", len(corners))
	}
	if len(edges) != len(corners) {
		return fmt.Errorf("%d edge slots for %d corners", len(edges), len(corners))
	}
	for i, corner := range corners {
		next := corners[(i+1)%len(corners)]
		if corner == next {
			return fmt.Errorf("corner %v is listed twice in a row", corner)
		}
		if corner.x != next.x && corner.y != next.y {
			return fmt.Errorf("wall %v-%v is not axis-aligned", corner, next)
		}
		edges[i] = edge{from: corner, to: next}
	}
	return nil
}

// area is the number of cells in the rectangle spanned by two corners, both
// ends included.
func area(a, b point) int {
	return (max(a.x, b.x) - min(a.x, b.x) + 1) * (max(a.y, b.y) - min(a.y, b.y) + 1)
}

// largestArea returns the biggest rectangle spanned by any two corners,
// ignoring where the polygon's walls run (part 1).
func largestArea(corners []point) (int, error) {
	if len(corners) < 2 {
		return 0, fmt.Errorf("%d corners span no rectangle", len(corners))
	}
	if len(corners) > maxLines {
		return 0, fmt.Errorf("%d corners, more than the %d allowed", len(corners), maxLines)
	}

	largest := 0
	for i := range corners {
		for j := i + 1; j < len(corners); j++ {
			largest = max(largest, area(corners[i], corners[j]))
		}
	}

	if largest < 1 {
		return 0, fmt.Errorf("largest rectangle has area %d", largest)
	}
	return largest, nil
}

// largestInteriorArea returns the biggest rectangle spanned by two corners that
// also lies wholly inside the polygon (part 2).
func largestInteriorArea(corners []point, edges []edge) (int, error) {
	if len(corners) < 4 || len(corners) > maxLines {
		return 0, fmt.Errorf("%d corners, expected 4..%d", len(corners), maxLines)
	}
	if len(edges) != len(corners) {
		return 0, fmt.Errorf("%d edges for %d corners", len(edges), len(corners))
	}

	largest := 0
	for i := range corners {
		for j := i + 1; j < len(corners); j++ {
			a, b := corners[i], corners[j]

			// Checking containment is expensive, so skip any rectangle that
			// could not beat the best one found so far.
			if area(a, b) <= largest {
				continue
			}
			inside, err := isInside(a, b, edges)
			if err != nil {
				return 0, err
			}
			if inside {
				largest = area(a, b)
			}
		}
	}

	if largest < 1 {
		return 0, errors.New("no rectangle lies inside the polygon")
	}
	return largest, nil
}

// isInside reports whether the rectangle spanned by a and b lies inside the
// polygon.
func isInside(a, b point, edges []edge) (bool, error) {
	if len(edges) < 4 {
		return false, fmt.Errorf("%d walls cannot enclose anything", len(edges))
	}
	minX, maxX := min(a.x, b.x), max(a.x, b.x)
	minY, maxY := min(a.y, b.y), max(a.y, b.y)
	if minX < 0 || minY < 0 || maxX > maxCoord || maxY > maxCoord {
		return false, fmt.Errorf("rectangle %v-%v is off the floor", a, b)
	}

	// A wall cutting through the rectangle means part of it is outside.
	for _, e := range edges {
		loX, hiX := min(e.from.x, e.to.x), max(e.from.x, e.to.x)
		loY, hiY := min(e.from.y, e.to.y), max(e.from.y, e.to.y)
		vertical := loX == hiX
		if vertical && loX > minX && loX < maxX && max(loY, minY) < min(hiY, maxY) {
			return false, nil
		}
		if !vertical && loY > minY && loY < maxY && max(loX, minX) < min(hiX, maxX) {
			return false, nil
		}
	}

	// No wall crosses it, so the whole rectangle is on the same side of the
	// outline as its centre. Doubling keeps a half-cell centre exact.
	return containsDoubled(edges, minX+maxX, minY+maxY)
}

// containsDoubled reports whether the point (x/2, y/2) lies inside the polygon
// or on its outline, using ray casting: a ray drawn to the right crosses the
// outline an odd number of times exactly when the point is inside. The walls
// are axis-aligned, so only vertical ones can cross a horizontal ray.
func containsDoubled(edges []edge, x, y int) (bool, error) {
	if len(edges) == 0 {
		return false, errors.New("no walls to test against")
	}
	if x < 0 || y < 0 || x > 2*maxCoord || y > 2*maxCoord {
		return false, fmt.Errorf("point (%d, %d)/2 is off the floor", x, y)
	}

	inside := false
	for _, e := range edges {
		x1, y1, x2, y2 := 2*e.from.x, 2*e.from.y, 2*e.to.x, 2*e.to.y

		// On the outline counts as inside.
		onVertical := x1 == x2 && x == x1 && y >= min(y1, y2) && y <= max(y1, y2)
		onHorizontal := y1 == y2 && y == y1 && x >= min(x1, x2) && x <= max(x1, x2)
		if onVertical || onHorizontal {
			return true, nil
		}

		// Count the crossing if the wall spans the ray's height. Comparing
		// with > on both ends keeps a corner from counting twice.
		if x1 == x2 && (y1 > y) != (y2 > y) && x < x1 {
			inside = !inside
		}
	}
	return inside, nil
}

func day9(path string) error {
	corners, err := parseInput(path)
	if err != nil {
		return err
	}
	if len(corners) < 4 {
		return fmt.Errorf("%d corners cannot close a polygon", len(corners))
	}
	edges := make([]edge, len(corners))
	if err := polygonEdges(corners, edges); err != nil {
		return err
	}

	part1, err := largestArea(corners)
	if err != nil {
		return err
	}
	part2, err := largestInteriorArea(corners, edges)
	if err != nil {
		return err
	}

	if part2 > part1 {
		return fmt.Errorf("interior rectangle %d is bigger than any rectangle %d", part2, part1)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", part1, part2); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day9(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
