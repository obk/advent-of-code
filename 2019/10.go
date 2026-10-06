// Day 10: Monitoring Station
//
// The input is a map of asteroids ('#'). A station on an asteroid sees
// another asteroid unless a third lies exactly on the line between them.
//
// Part 1 is the most asteroids any one asteroid can see.
// Part 2 puts a laser on that asteroid. Pointing up and turning clockwise,
// it destroys the first asteroid in each direction per turn. The answer is
// 100*x + y of the 200th asteroid destroyed.
//
// Seen from a station, two asteroids hide each other exactly when their
// offsets reduce (dividing by the gcd) to the same direction. So a station
// sees as many asteroids as there are distinct reduced directions, counted
// with a fixed table indexed by direction.
//
// An asteroid that is the k-th nearest in its direction goes on turn k, and
// within a turn the clockwise angle decides. Its place in the whole order is
// the number of asteroids before it by (turn, angle).
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 16
	maxSide       = 64
	span          = 2*maxSide + 1 // directions run from -maxSide to maxSide
)

// field is the asteroid map and a direction table for counting.
type field struct {
	rocks [][2]int
	seen  [span * span]bool
	turn  []int
	angle []float64
}

// load parses the map.
func (f *field) load(path string) error {
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
	rows := strings.Fields(string(content))
	if len(rows) > maxSide {
		return fmt.Errorf("%s: %d rows, more than %d", path, len(rows), maxSide)
	}
	f.rocks = make([][2]int, 0, maxSide*maxSide)
	f.turn = make([]int, maxSide*maxSide)
	f.angle = make([]float64, maxSide*maxSide)
	for y, row := range rows {
		if len(row) > maxSide || strings.Trim(row, ".#") != "" {
			return fmt.Errorf("%s: row %d is not up to %d of '.' and '#'", path, y, maxSide)
		}
		for x := range len(row) {
			if row[x] == '#' {
				f.rocks = append(f.rocks, [2]int{x, y})
			}
		}
	}
	if len(f.rocks) < 2 {
		return fmt.Errorf("%s: %d asteroids", path, len(f.rocks))
	}
	return nil
}

// gcd is Euclid's algorithm for non-negative numbers below 2*maxSide.
func gcd(a, b int) int {
	for i := 0; i < 2*maxSide && b != 0; i++ {
		a, b = b, a%b
	}
	return a
}

// visible counts the asteroids station s can see.
func (f *field) visible(s int) (int, error) {
	if s < 0 || s >= len(f.rocks) {
		return 0, fmt.Errorf("asteroid %d does not exist", s)
	}
	clear(f.seen[:])
	count := 0
	for i, r := range f.rocks {
		if i == s {
			continue
		}
		dx, dy := r[0]-f.rocks[s][0], r[1]-f.rocks[s][1]
		g := gcd(max(dx, -dx), max(dy, -dy))
		if g == 0 {
			return 0, errors.New("two asteroids share a position")
		}
		k := (dy/g+maxSide)*span + dx/g + maxSide
		if k < 0 || k >= len(f.seen) {
			return 0, fmt.Errorf("direction %d,%d is outside the table", dx, dy)
		}
		if !f.seen[k] {
			f.seen[k], count = true, count+1
		}
	}
	if count >= len(f.rocks) {
		return 0, fmt.Errorf("asteroid %d sees %d of %d others", s, count, len(f.rocks)-1)
	}
	return count, nil
}

// best returns the asteroid that sees the most, and how many it sees.
func (f *field) best() (station, most int, err error) {
	if len(f.rocks) < 2 {
		return 0, 0, errors.New("fewer than two asteroids")
	}
	for s := range f.rocks {
		n, err := f.visible(s)
		if err != nil {
			return 0, 0, err
		}
		if n > most {
			station, most = s, n
		}
	}
	if most < 1 {
		return 0, 0, errors.New("no asteroid sees another")
	}
	return station, most, nil
}

// direction returns asteroid r's reduced offset from asteroid s and the
// gcd it was reduced by.
func (f *field) direction(s, r int) (dx, dy, g int, err error) {
	if s < 0 || r < 0 || s >= len(f.rocks) || r >= len(f.rocks) || s == r {
		return 0, 0, 0, fmt.Errorf("asteroids %d and %d are not a pair", s, r)
	}
	dx, dy = f.rocks[r][0]-f.rocks[s][0], f.rocks[r][1]-f.rocks[s][1]
	g = gcd(max(dx, -dx), max(dy, -dy))
	if g == 0 {
		return 0, 0, 0, errors.New("two asteroids share a position")
	}
	return dx / g, dy / g, g, nil
}

// vaporized returns the n-th asteroid (from 1) the laser at s destroys.
func (f *field) vaporized(s, n int) (int, error) {
	if s < 0 || s >= len(f.rocks) {
		return 0, fmt.Errorf("station %d does not exist", s)
	}
	if n < 1 || n >= len(f.rocks) {
		return 0, fmt.Errorf("asteroid %d of %d cannot be destroyed", n, len(f.rocks)-1)
	}
	for r := range f.rocks {
		if r == s {
			continue
		}
		dx, dy, g, err := f.direction(s, r)
		if err != nil {
			return 0, err
		}
		// Clockwise from up, with y growing downwards; in [0, 2*pi).
		f.angle[r] = math.Mod(math.Atan2(float64(dx), float64(-dy))+2*math.Pi, 2*math.Pi)
		f.turn[r] = 0
		for o := range f.rocks {
			if o == s || o == r {
				continue
			}
			ox, oy, og, err := f.direction(s, o)
			if err != nil {
				return 0, err
			}
			if ox == dx && oy == dy && og < g {
				f.turn[r]++
			}
		}
	}
	for r := range f.rocks {
		if r == s {
			continue
		}
		before := 0
		for o := range f.rocks {
			if o != s && o != r && (f.turn[o] < f.turn[r] || (f.turn[o] == f.turn[r] && f.angle[o] < f.angle[r])) {
				before++
			}
		}
		if before == n-1 {
			return r, nil
		}
	}
	return 0, fmt.Errorf("no asteroid is destroyed %d-th", n)
}

func day10(path string) error {
	f := new(field)
	if err := f.load(path); err != nil {
		return err
	}
	station, most, err := f.best()
	if err != nil {
		return err
	}
	r, err := f.vaporized(station, 200)
	if err != nil {
		return err
	}
	if r == station {
		return errors.New("the laser destroyed its own asteroid")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", most, 100*f.rocks[r][0]+f.rocks[r][1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
