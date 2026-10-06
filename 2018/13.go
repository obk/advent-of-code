// Day 13: Mine Cart Madness
//
// The input is a map of tracks ('|', '-', '/', '\', '+') with carts on them
// ('^', 'v', '<', '>'). Each tick the carts move one step in reading order of
// their positions. Curves turn a cart; at a '+' a cart turns left, goes
// straight, then turns right, cycling through those choices.
//
// Part 1 is the location "x,y" of the first crash.
// Part 2 removes both carts at every crash, and asks where the last cart is
// at the end of the first tick it is alone.
//
// The tracks are kept as the input text with the carts replaced by the track
// beneath them. Each tick the carts are put in reading order with an
// insertion sort (there are only a few), and a crash is any cart moving onto
// another.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 16
	maxCarts      = 64
	maxTicks      = 1_000_000
)

// cart is one cart: position, direction (0 up, 1 right, 2 down, 3 left), how
// many intersections it has passed, and whether it has crashed.
type cart struct {
	x, y, dir, turns int
	crashed          bool
}

// railway is the tracks and the carts on them.
type railway struct {
	tracks [][]byte
	carts  [maxCarts]cart
	n      int
}

// load parses the map, replacing each cart by the straight track under it.
func (r *railway) load(path string) error {
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
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("%s: empty map", path)
	}
	r.tracks = make([][]byte, len(lines))
	for y, line := range lines {
		r.tracks[y] = []byte(line)
		for x := range len(line) {
			dir := strings.IndexByte("^>v<", line[x])
			if dir < 0 {
				continue
			}
			if r.n == maxCarts {
				return fmt.Errorf("%s: more than %d carts", path, maxCarts)
			}
			r.carts[r.n] = cart{x: x, y: y, dir: dir}
			r.n++
			r.tracks[y][x] = "|-|-"[dir]
		}
	}
	if r.n < 2 {
		return fmt.Errorf("%s: %d carts cannot crash", path, r.n)
	}
	return nil
}

// sortCarts puts the carts in reading order of their positions.
func (r *railway) sortCarts() error {
	if r.n > maxCarts {
		return fmt.Errorf("%d carts, more than %d", r.n, maxCarts)
	}
	for i := 1; i < r.n; i++ {
		for j := i; j > 0; j-- {
			a, b := r.carts[j-1], r.carts[j]
			if a.y < b.y || (a.y == b.y && a.x <= b.x) {
				break
			}
			r.carts[j-1], r.carts[j] = b, a
		}
	}
	for i := 1; i < r.n; i++ {
		a, b := r.carts[i-1], r.carts[i]
		if a.y > b.y || (a.y == b.y && a.x > b.x) {
			return fmt.Errorf("carts out of order at %d", i)
		}
	}
	return nil
}

// move advances cart i one step and turns it for the track it lands on.
func (r *railway) move(i int) error {
	if i < 0 || i >= r.n {
		return fmt.Errorf("cart %d does not exist", i)
	}
	if r.carts[i].crashed {
		return fmt.Errorf("cart %d has crashed and cannot move", i)
	}
	c := &r.carts[i]
	c.x += [4]int{0, 1, 0, -1}[c.dir]
	c.y += [4]int{-1, 0, 1, 0}[c.dir]
	if c.y < 0 || c.y >= len(r.tracks) || c.x < 0 || c.x >= len(r.tracks[c.y]) {
		return fmt.Errorf("cart ran off the map at %d,%d", c.x, c.y)
	}
	switch r.tracks[c.y][c.x] {
	case '/':
		c.dir = [4]int{1, 0, 3, 2}[c.dir]
	case '\\':
		c.dir = [4]int{3, 2, 1, 0}[c.dir]
	case '+':
		c.dir = (c.dir + [3]int{3, 0, 1}[c.turns%3]) % 4
		c.turns++
	case '|', '-':
	default:
		return fmt.Errorf("cart left the track at %d,%d", c.x, c.y)
	}
	return nil
}

// firstCrash runs ticks until two carts collide and returns where.
func (r *railway) firstCrash() (x, y int, err error) {
	if r.n < 2 {
		return 0, 0, fmt.Errorf("%d carts cannot crash", r.n)
	}
	for range maxTicks {
		if err := r.sortCarts(); err != nil {
			return 0, 0, err
		}
		for i := range r.n {
			if err := r.move(i); err != nil {
				return 0, 0, err
			}
			for j := range r.n {
				if j != i && r.carts[j].x == r.carts[i].x && r.carts[j].y == r.carts[i].y {
					return r.carts[i].x, r.carts[i].y, nil
				}
			}
		}
	}
	return 0, 0, fmt.Errorf("no crash within %d ticks", maxTicks)
}

// lastCart runs ticks, removing crashed carts, until one cart is left, and
// returns its position at the end of that tick.
func (r *railway) lastCart() (x, y int, err error) {
	if r.n%2 == 0 {
		return 0, 0, fmt.Errorf("%d carts cannot leave exactly one", r.n)
	}
	if r.n > maxCarts {
		return 0, 0, fmt.Errorf("%d carts, more than %d", r.n, maxCarts)
	}
	for range maxTicks {
		if err := r.sortCarts(); err != nil {
			return 0, 0, err
		}
		for i := range r.n {
			if r.carts[i].crashed {
				continue
			}
			if err := r.move(i); err != nil {
				return 0, 0, err
			}
			for j := range r.n {
				c := &r.carts[j]
				if j != i && !c.crashed && c.x == r.carts[i].x && c.y == r.carts[i].y {
					c.crashed, r.carts[i].crashed = true, true
				}
			}
		}
		left, last := 0, -1
		for i := range r.n {
			if !r.carts[i].crashed {
				left, last = left+1, i
			}
		}
		if left == 1 {
			return r.carts[last].x, r.carts[last].y, nil
		}
	}
	return 0, 0, fmt.Errorf("more than one cart left after %d ticks", maxTicks)
}

func day13(path string) error {
	r := new(railway)
	if err := r.load(path); err != nil {
		return err
	}
	x, y, err := r.firstCrash()
	if err != nil {
		return err
	}
	r2 := new(railway)
	if err := r2.load(path); err != nil {
		return err
	}
	lx, ly, err := r2.lastCart()
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d,%d\nPart 2: %d,%d\n", x, y, lx, ly); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
