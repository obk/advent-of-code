// Day 24: Air Duct Spelunking
//
// The input is a map of air ducts: '#' walls, '.' open, and digits marking
// points of interest, 0 being the robot's start.
//
// Part 1 is the fewest steps to visit every point of interest, starting at 0.
// Part 2 is the fewest steps to do that and then return to 0.
//
// A breadth-first search from each point gives the distances between every
// pair of points; the order of visits is then a small travelling salesman
// problem, solved exactly with the Held-Karp table best[set][last].
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxPoints     = 10
	unreachable   = 1 << 40
)

// ducts holds the map, where each point of interest is, and BFS scratch space.
type ducts struct {
	grid   []string
	width  int
	points [maxPoints]int // cell index of each point, -1 if absent
	count  int
	dist   []int
	queue  []int
	pair   [maxPoints][maxPoints]int
	best   [1 << maxPoints][maxPoints]int
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

// load checks the map and finds the points of interest.
func (d *ducts) load(lines []string) error {
	if len(lines) == 0 || len(lines[0]) == 0 {
		return errors.New("empty map")
	}
	d.grid, d.width = lines, len(lines[0])
	for i := range d.points {
		d.points[i] = -1
	}
	for r, line := range lines {
		if len(line) != d.width {
			return fmt.Errorf("row %d is %d wide, expected %d", r, len(line), d.width)
		}
		for c := range len(line) {
			ch := line[c]
			if ch >= '0' && ch <= '9' {
				if int(ch-'0') >= maxPoints || d.points[ch-'0'] >= 0 {
					return fmt.Errorf("point %c appears twice or is too high", ch)
				}
				d.points[ch-'0'] = r*d.width + c
				d.count = max(d.count, int(ch-'0')+1)
			} else if ch != '#' && ch != '.' {
				return fmt.Errorf("row %d holds %q", r, ch)
			}
		}
	}
	for i := range d.count {
		if d.points[i] < 0 {
			return fmt.Errorf("point %d is missing", i)
		}
	}
	d.dist = make([]int, len(lines)*d.width)
	d.queue = make([]int, len(lines)*d.width)
	return nil
}

// distancesFrom runs a BFS from point p and records its distance to every
// other point.
func (d *ducts) distancesFrom(p int) error {
	if p < 0 || p >= d.count {
		return fmt.Errorf("point %d does not exist", p)
	}
	for i := range d.dist {
		d.dist[i] = -1
	}
	start := d.points[p]
	d.dist[start], d.queue[0] = 0, start
	tail := 1
	for head := 0; head < tail; head++ {
		cell := d.queue[head]
		for _, next := range [4]int{cell - 1, cell + 1, cell - d.width, cell + d.width} {
			if next < 0 || next >= len(d.dist) || d.dist[next] >= 0 {
				continue
			}
			if d.grid[next/d.width][next%d.width] == '#' {
				continue
			}
			d.dist[next] = d.dist[cell] + 1
			d.queue[tail] = next
			tail++
		}
	}
	for q := range d.count {
		d.pair[p][q] = d.dist[d.points[q]]
		if d.pair[p][q] < 0 {
			return fmt.Errorf("point %d cannot reach point %d", p, q)
		}
	}
	return nil
}

// shortestTour fills the Held-Karp table for routes starting at point 0 and
// returns the shortest that visits every point, and the shortest that also
// comes back to 0.
func (d *ducts) shortestTour() (open, closed int, err error) {
	if d.count < 1 || d.count > maxPoints {
		return 0, 0, fmt.Errorf("%d points is outside 1..%d", d.count, maxPoints)
	}
	all := 1<<d.count - 1
	for set := range all + 1 {
		for last := range d.count {
			d.best[set][last] = unreachable
		}
	}
	d.best[1][0] = 0
	for set := 1; set <= all; set += 2 { // every route includes point 0
		for last := range d.count {
			if d.best[set][last] == unreachable {
				continue
			}
			for next := range d.count {
				if set>>next&1 == 1 {
					continue
				}
				grown := set | 1<<next
				d.best[grown][next] = min(d.best[grown][next], d.best[set][last]+d.pair[last][next])
			}
		}
	}

	open, closed = unreachable, unreachable
	for last := range d.count {
		open = min(open, d.best[all][last])
		closed = min(closed, d.best[all][last]+d.pair[last][0])
	}
	if open == unreachable || closed < open {
		return 0, 0, fmt.Errorf("no sensible tour (open %d, closed %d)", open, closed)
	}
	return open, closed, nil
}

func day24(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	d := new(ducts)
	if err := d.load(lines); err != nil {
		return err
	}
	for p := range d.count {
		if err := d.distancesFrom(p); err != nil {
			return err
		}
	}
	tour, round, err := d.shortestTour()
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", tour, round); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
