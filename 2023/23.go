// Day 23: A Long Walk
//
// The input is a map of paths '.', forest '#' and slopes (^ > v <). A hike
// goes from the gap in the top row to the gap in the bottom row, never
// stepping on the same tile twice.
//
// Part 1 is the longest hike when slopes can only be walked downhill (in
// their arrow's direction).
// Part 2 is the longest hike treating slopes as ordinary paths.
//
// The paths are corridors between junctions (tiles with three or more
// ways out). Walking each corridor gives a small graph of junctions with
// corridor lengths; a depth-first search with an explicit stack and a
// bit set of visited junctions then tries every hike.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	maxNodes      = 64
)

// trails is the map, its junctions and the corridor lengths between them
// (0 for no corridor), one-way with slopes and both ways without.
type trails struct {
	rows        []string
	w, h        int
	node        [maxSide][maxSide]int // junction number + 1
	n           int
	slope, flat [maxNodes][maxNodes]int
	start, end  int
}

// open reports whether (x, y) can be walked on.
func (t *trails) open(x, y int) bool {
	return x >= 0 && y >= 0 && x < t.w && y < t.h && t.rows[y][x] != '#'
}

// findNodes numbers the junctions, the start and the end.
func (t *trails) findNodes() error {
	for y := range t.h {
		for x := range t.w {
			if !t.open(x, y) {
				continue
			}
			ways := 0
			for _, d := range [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
				if t.open(x+d[0], y+d[1]) {
					ways++
				}
			}
			if ways < 3 && y != 0 && y != t.h-1 {
				continue
			}
			if t.n == maxNodes {
				return fmt.Errorf("more than %d junctions", maxNodes)
			}
			t.node[y][x], t.n = t.n+1, t.n+1
			if y == 0 {
				t.start = t.n - 1
			}
			if y == t.h-1 {
				t.end = t.n - 1
			}
		}
	}
	if t.n < 2 {
		return errors.New("no start and end")
	}
	if t.start == t.end {
		return errors.New("start and end are the same tile")
	}
	return nil
}

// corridor walks from junction tile (x, y) in direction d to the next
// junction, returning it, the length, and whether slopes allow the walk.
func (t *trails) corridor(x, y, d int) (to, length int, downhill bool, err error) {
	if d < 0 || d > 3 {
		return 0, 0, false, fmt.Errorf("direction %d does not exist", d)
	}
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	downhill = true
	for length = 1; length <= t.w*t.h; length++ {
		x, y = x+dirs[d][0], y+dirs[d][1]
		if c := t.rows[y][x]; c != '.' && strings.IndexByte("^>v<", c) != d {
			downhill = false
		}
		if t.node[y][x] > 0 {
			return t.node[y][x] - 1, length, downhill, nil
		}
		next := -1
		for k := range 4 {
			if k != (d+2)%4 && t.open(x+dirs[k][0], y+dirs[k][1]) {
				next = k
			}
		}
		if next < 0 {
			return -1, 0, false, nil // a dead end
		}
		d = next
	}
	return 0, 0, false, errors.New("a corridor never ended")
}

// connect walks every corridor out of every junction.
func (t *trails) connect() error {
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	for y := range t.h {
		for x := range t.w {
			from := t.node[y][x] - 1
			if from < 0 {
				continue
			}
			for d := range 4 {
				if !t.open(x+dirs[d][0], y+dirs[d][1]) {
					continue
				}
				to, length, downhill, err := t.corridor(x, y, d)
				if err != nil {
					return err
				}
				if to < 0 || to == from {
					continue
				}
				t.flat[from][to] = max(t.flat[from][to], length)
				if downhill {
					t.slope[from][to] = max(t.slope[from][to], length)
				}
			}
		}
	}
	for a := range t.n {
		for b := range t.n {
			if t.flat[a][b] != t.flat[b][a] {
				return fmt.Errorf("corridor %d-%d differs by direction", a, b)
			}
		}
	}
	return nil
}

// frame is one junction on the hike.
type frame struct {
	at, next, length int
	visited          uint64
}

// longest returns the longest hike over the given corridors.
func (t *trails) longest(edges *[maxNodes][maxNodes]int) (int, error) {
	// Each junction's neighbours as a short list, so the search does not
	// scan every junction at every step.
	if t.n > maxNodes || edges == nil {
		return 0, errors.New("too many junctions, or no corridors")
	}
	var adj [maxNodes][4]int
	var deg [maxNodes]int
	for a := range t.n {
		for b := range t.n {
			if edges[a][b] == 0 {
				continue
			}
			if deg[a] == 4 {
				return 0, fmt.Errorf("junction %d has more than four corridors", a)
			}
			adj[a][deg[a]], deg[a] = b, deg[a]+1
		}
	}
	var stack [maxNodes + 1]frame
	stack[0] = frame{at: t.start, visited: 1 << t.start}
	depth, best := 1, -1
	for steps := 0; depth > 0; steps++ {
		if steps > 1<<30 {
			return 0, errors.New("the search ran too long")
		}
		f := &stack[depth-1]
		if f.at == t.end || f.next == deg[f.at] {
			if f.at == t.end {
				best = max(best, f.length)
			}
			depth--
			continue
		}
		to := adj[f.at][f.next]
		f.next++
		if f.visited>>to&1 == 1 {
			continue
		}
		stack[depth] = frame{at: to, length: f.length + edges[f.at][to], visited: f.visited | 1<<to}
		depth++
	}
	if best < 0 {
		return 0, errors.New("no hike reaches the end")
	}
	return best, nil
}

func day23(path string) error {
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
	t := &trails{rows: strings.Fields(string(content))}
	t.h = len(t.rows)
	if t.h < 3 || t.h > maxSide || len(t.rows[0]) > maxSide {
		return fmt.Errorf("%s: map is too small or larger than %d", path, maxSide)
	}
	t.w = len(t.rows[0])
	for y, r := range t.rows {
		if len(r) != t.w || strings.Trim(r, ".#^>v<") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	if err := t.findNodes(); err != nil {
		return err
	}
	if err := t.connect(); err != nil {
		return err
	}
	downhill, err := t.longest(&t.slope)
	if err != nil {
		return err
	}
	any, err := t.longest(&t.flat)
	if err != nil {
		return err
	}

	// Every downhill hike is also a hike with slopes as paths.
	if any < downhill {
		return fmt.Errorf("the hike without slopes (%d) is shorter than with (%d)", any, downhill)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", downhill, any); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day23(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
