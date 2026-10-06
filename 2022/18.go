// Day 18: Boiling Boulders
//
// Each line of the input is a 1x1x1 cube of lava, "x,y,z".
//
// Part 1 counts cube faces not touching another cube.
// Part 2 counts only faces reachable from outside: air pockets trapped
// inside do not count.
//
// The cubes go in a fixed grid with a one-cell margin; a breadth-first
// flood of the outside air from a corner counts the faces it meets.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 16
	side          = 32 // coordinates 0..side-3 plus a margin either side
	cells         = side * side * side
)

// droplet is the lava and the flood's buffers.
type droplet struct {
	lava, air [side][side][side]bool
	queue     [cells][3]int
}

// load marks the cubes (shifted by one for the margin) and counts them.
func (d *droplet) load(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Fields(string(content))
	if len(lines) == 0 {
		return 0, fmt.Errorf("%s: no cubes", path)
	}
	for _, l := range lines {
		var x, y, z int
		n, err := fmt.Sscanf(l, "%d,%d,%d", &x, &y, &z)
		if err != nil || n != 3 || min(x, y, z) < 0 || max(x, y, z) > side-3 {
			return 0, fmt.Errorf("%s: invalid cube %q", path, l)
		}
		if d.lava[x+1][y+1][z+1] {
			return 0, fmt.Errorf("%s: cube %q is listed twice", path, l)
		}
		d.lava[x+1][y+1][z+1] = true
	}
	return len(lines), nil
}

// surface counts lava faces touching anything but lava.
func (d *droplet) surface() (int, error) {
	n := 0
	for x := 1; x < side-1; x++ {
		for y := 1; y < side-1; y++ {
			for z := 1; z < side-1; z++ {
				for _, v := range [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
					if d.lava[x][y][z] && !d.lava[x+v[0]][y+v[1]][z+v[2]] {
						n++
					}
				}
			}
		}
	}
	if n > 6*cells {
		return 0, fmt.Errorf("%d faces is more than the grid holds", n)
	}
	return n, nil
}

// exterior floods the outside air and counts the lava faces it touches.
func (d *droplet) exterior() (int, error) {
	if d.lava[0][0][0] {
		return 0, errors.New("the corner must be air")
	}
	d.air[0][0][0], d.queue[0] = true, [3]int{0, 0, 0}
	faces, tail := 0, 1
	for head := 0; head < tail; head++ {
		p := d.queue[head]
		for _, v := range [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
			x, y, z := p[0]+v[0], p[1]+v[1], p[2]+v[2]
			if min(x, y, z) < 0 || max(x, y, z) >= side || d.air[x][y][z] {
				continue
			}
			if d.lava[x][y][z] {
				faces++
				continue
			}
			if tail == len(d.queue) {
				return 0, errors.New("flood queue is full")
			}
			d.air[x][y][z], d.queue[tail], tail = true, [3]int{x, y, z}, tail+1
		}
	}
	return faces, nil
}

func day18(path string) error {
	d := new(droplet)
	count, err := d.load(path)
	if err != nil {
		return err
	}
	all, err := d.surface()
	if err != nil {
		return err
	}
	outside, err := d.exterior()
	if err != nil {
		return err
	}
	if outside < 1 {
		return errors.New("no face can be reached from outside")
	}
	if outside > all || all > 6*count {
		return fmt.Errorf("%d outside faces, %d faces, %d cubes do not add up", outside, all, count)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", all, outside); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
