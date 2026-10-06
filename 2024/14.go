// Day 14: Restroom Redoubt
//
// Each line of the input is a robot, "p=0,4 v=3,-3", on a 101x103 grid
// whose edges wrap.
//
// Part 1 multiplies the robot counts of the four quadrants (ignoring the
// middle row and column) after 100 seconds.
// Part 2 is the first second at which the robots draw a Christmas tree.
//
// The picture appears when every robot is on its own tile, which is taken
// as the test. Positions repeat after 101*103 seconds, which bounds the
// search.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 16
	width         = 101
	height        = 103
	maxRobots     = 1_000
)

// robot is a position and velocity.
type robot struct {
	px, py, vx, vy int
}

// at returns where r is after t seconds.
func at(r robot, t int) (int, int) {
	return ((r.px+t*r.vx)%width + width) % width, ((r.py+t*r.vy)%height + height) % height
}

// safety multiplies the quadrant counts after t seconds.
func safety(robots []robot, t int) (int, error) {
	if t < 0 || len(robots) == 0 {
		return 0, errors.New("negative time or no robots")
	}
	var q [4]int
	for _, r := range robots {
		x, y := at(r, t)
		if x == width/2 || y == height/2 {
			continue
		}
		k := 0
		if x > width/2 {
			k++
		}
		if y > height/2 {
			k += 2
		}
		q[k]++
	}
	if q[0]+q[1]+q[2]+q[3] > len(robots) {
		return 0, errors.New("counted more robots than there are")
	}
	return q[0] * q[1] * q[2] * q[3], nil
}

// tree finds the first second with every robot on its own tile.
func tree(robots []robot) (int, error) {
	if len(robots) == 0 || len(robots) > maxRobots {
		return 0, fmt.Errorf("%d robots is out of range", len(robots))
	}
	if len(robots) > width*height {
		return 0, errors.New("more robots than tiles")
	}
	var stamp [height][width]int
	for t := 1; t <= width*height; t++ {
		clash := false
		for _, r := range robots {
			x, y := at(r, t)
			if stamp[y][x] == t {
				clash = true
				break
			}
			stamp[y][x] = t
		}
		if !clash {
			return t, nil
		}
	}
	return 0, errors.New("the robots never spread out")
}

func day14(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxRobots {
		return fmt.Errorf("%s: more than %d robots", path, maxRobots)
	}
	robots := make([]robot, len(lines))
	for i, l := range lines {
		r := &robots[i]
		n, err := fmt.Sscanf(l, "p=%d,%d v=%d,%d", &r.px, &r.py, &r.vx, &r.vy)
		if err != nil || n != 4 || r.px < 0 || r.py < 0 || r.px >= width || r.py >= height {
			return fmt.Errorf("%s: invalid robot %q", path, l)
		}
		if max(r.vx, -r.vx) > width || max(r.vy, -r.vy) > height {
			return fmt.Errorf("%s: robot %q is too fast", path, l)
		}
	}
	factor, err := safety(robots, 100)
	if err != nil {
		return err
	}
	t, err := tree(robots)
	if err != nil {
		return err
	}

	if factor < 1 {
		return errors.New("a quadrant is empty")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", factor, t); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
