// Day 20: Race Condition
//
// The input is a racetrack: one path from S to E through walls. Once per
// race, a program may cheat: pass through walls for up to n picoseconds,
// ending back on the track.
//
// Part 1 counts the cheats of up to 2 picoseconds that save at least 100.
// Part 2 counts those of up to 20 picoseconds.
//
// The track's distance from S is found along its single path. A cheat
// from track tile a to track tile b within Manhattan distance n saves
// dist[b] - dist[a] - distance(a, b).
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	minSaving     = 100
)

// track is the map and each tile's distance from S (-1 off the track).
type track struct {
	rows []string
	w, h int
	dist [maxSide][maxSide]int
}

// trace walks the single path from S, numbering tiles; it returns the
// track's length.
func (t *track) trace(sx, sy int) (int, error) {
	for y := range t.h {
		for x := range t.w {
			t.dist[y][x] = -1
		}
	}
	x, y, d := sx, sy, 0
	t.dist[y][x] = 0
	for step := 0; step < t.w*t.h; step++ {
		if t.rows[y][x] == 'E' {
			return d, nil
		}
		next := false
		for _, m := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := x+m[0], y+m[1]
			if t.rows[ny][nx] != '#' && t.dist[ny][nx] < 0 && !next {
				x, y, d, next = nx, ny, d+1, true
				t.dist[y][x] = d
			}
		}
		if !next {
			return 0, errors.New("the track has a dead end")
		}
	}
	return 0, errors.New("the track never reaches E")
}

// cheats counts cheats of up to n that save at least minSaving.
func (t *track) cheats(n int) (int, error) {
	if n < 1 || n > maxSide {
		return 0, fmt.Errorf("cheat length %d is out of range", n)
	}
	count := 0
	for y := range t.h {
		for x := range t.w {
			if t.dist[y][x] < 0 {
				continue
			}
			for dy := -n; dy <= n; dy++ {
				for dx := -n + max(dy, -dy); dx <= n-max(dy, -dy); dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || ny < 0 || nx >= t.w || ny >= t.h || t.dist[ny][nx] < 0 {
						continue
					}
					if t.dist[ny][nx]-t.dist[y][x]-max(dx, -dx)-max(dy, -dy) >= minSaving {
						count++
					}
				}
			}
		}
	}
	return count, nil
}

func day20(path string) error {
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
	t := &track{rows: strings.Fields(string(content))}
	t.h = len(t.rows)
	if t.h < 3 || t.h > maxSide || len(t.rows[0]) > maxSide {
		return fmt.Errorf("%s: track is too small or larger than %d", path, maxSide)
	}
	t.w = len(t.rows[0])
	sx, sy := -1, -1
	for y, r := range t.rows {
		if len(r) != t.w || strings.Trim(r, ".#SE") != "" || r[0] != '#' || r[t.w-1] != '#' {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if x := strings.IndexByte(r, 'S'); x >= 0 {
			sx, sy = x, y
		}
	}
	if sx < 0 {
		return fmt.Errorf("%s: no S", path)
	}
	if _, err := t.trace(sx, sy); err != nil {
		return err
	}
	short, err := t.cheats(2)
	if err != nil {
		return err
	}
	long, err := t.cheats(20)
	if err != nil {
		return err
	}

	// Every cheat of up to 2 is also a cheat of up to 20.
	if long < short {
		return fmt.Errorf("%d long cheats but %d short ones", long, short)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", short, long); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
