// Day 23: Experimental Emergency Teleportation
//
// Each line of the input is a nanobot, "pos=<x,y,z>, r=R": it reaches every
// point within taxicab distance R.
//
// Part 1 counts the nanobots in range of the one with the largest radius
// (itself included).
// Part 2 finds the points in range of the most nanobots and gives the
// smallest distance from the origin to one of them.
//
// Part 2 searches cubes, best first. A cube's score is the number of bots
// whose range touches it - an upper bound for any point inside. Starting
// from one cube around everything, the cube with the highest score (then
// nearest the origin, then smallest) is split into eight. The first cube of
// size 1 to come out has a score no other cube can beat, so its point is the
// answer. The cubes wait in a binary heap in a fixed array.
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
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 20
	maxBots       = 2_000
	maxValue      = 1 << 40
	maxHeap       = 1 << 20
	maxPops       = 10_000_000
)

// cube is a search box: corner, side (a power of two), and its score.
type cube struct {
	x, y, z, side int
	bots, origin  int // bots touching it, and its distance from the origin
}

// bot is one nanobot.
type bot struct {
	x, y, z, r int
}

// readBots parses the nanobots.
func readBots(path string) ([]bot, error) {
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
	if len(lines) > maxBots {
		return nil, fmt.Errorf("%s: %d nanobots, more than %d", path, len(lines), maxBots)
	}
	bots := make([]bot, len(lines))
	clean := strings.NewReplacer("pos=<", "", ">", "", "r=", "", ",", " ")
	for i, line := range lines {
		f := strings.Fields(clean.Replace(line))
		if len(f) != 4 {
			return nil, fmt.Errorf("%s: malformed nanobot %q", path, line)
		}
		var n [4]int
		for k, s := range f {
			if n[k], err = strconv.Atoi(s); err != nil {
				return nil, fmt.Errorf("%s: invalid number in %q: %w", path, line, err)
			}
			if n[k] < -maxValue || n[k] > maxValue {
				return nil, fmt.Errorf("%s: number in %q is out of range", path, line)
			}
		}
		if n[3] < 0 {
			return nil, fmt.Errorf("%s: negative radius in %q", path, line)
		}
		bots[i] = bot{n[0], n[1], n[2], n[3]}
	}
	return bots, nil
}

// inRangeOfStrongest counts the bots the largest-radius bot reaches.
func inRangeOfStrongest(bots []bot) (int, error) {
	if len(bots) == 0 {
		return 0, errors.New("no nanobots")
	}
	strongest := bots[0]
	for _, b := range bots {
		if b.r > strongest.r {
			strongest = b
		}
	}
	count := 0
	for _, b := range bots {
		d := max(b.x-strongest.x, strongest.x-b.x) + max(b.y-strongest.y, strongest.y-b.y) + max(b.z-strongest.z, strongest.z-b.z)
		if d <= strongest.r {
			count++
		}
	}
	if count < 1 {
		return 0, errors.New("the strongest nanobot does not reach itself")
	}
	if count > len(bots) {
		return 0, fmt.Errorf("%d bots in range of %d", count, len(bots))
	}
	return count, nil
}

// before reports whether a should be searched before b.
func before(a, b cube) bool {
	if a.bots != b.bots {
		return a.bots > b.bots
	}
	if a.origin != b.origin {
		return a.origin < b.origin
	}
	return a.side < b.side
}

// score fills in how many bots touch c and how far c is from the origin.
func score(c cube, bots []bot) (cube, error) {
	if c.side < 1 || len(bots) == 0 {
		return c, errors.New("empty cube or no bots")
	}
	c.bots, c.origin = 0, 0
	for _, b := range bots {
		d := 0
		for _, axis := range [3][2]int{{b.x, c.x}, {b.y, c.y}, {b.z, c.z}} {
			v, lo, hi := axis[0], axis[1], axis[1]+c.side-1
			d += max(lo-v, 0) + max(v-hi, 0)
		}
		if d <= b.r {
			c.bots++
		}
	}
	for _, lo := range [3]int{c.x, c.y, c.z} {
		hi := lo + c.side - 1
		c.origin += max(lo, 0) + max(-hi, 0) // nearest point of [lo, hi] to 0
	}
	if c.bots > len(bots) {
		return c, fmt.Errorf("%d bots touch a cube out of %d", c.bots, len(bots))
	}
	return c, nil
}

// heapPush and heapPop keep the cubes in best-first order.
func heapPush(h []cube, c cube) ([]cube, error) {
	if len(h) == cap(h) {
		return nil, fmt.Errorf("more than %d cubes queued", cap(h))
	}
	if c.side < 1 {
		return nil, fmt.Errorf("cube of side %d", c.side)
	}
	h = append(h, c)
	for i := len(h) - 1; i > 0 && before(h[i], h[(i-1)/2]); i = (i - 1) / 2 {
		h[i], h[(i-1)/2] = h[(i-1)/2], h[i]
	}
	return h, nil
}

func heapPop(h []cube) ([]cube, cube, error) {
	if len(h) == 0 {
		return nil, cube{}, errors.New("no cubes left")
	}
	top, last := h[0], len(h)-1
	h[0], h = h[last], h[:last]
	if top.side < 1 {
		return nil, cube{}, fmt.Errorf("popped a cube of side %d", top.side)
	}
	for i, level := 0, 0; 2*i+1 < len(h) && level < 64; level++ {
		child := 2*i + 1
		if child+1 < len(h) && before(h[child+1], h[child]) {
			child++
		}
		if !before(h[child], h[i]) {
			break
		}
		h[i], h[child] = h[child], h[i]
		i = child
	}
	return h, top, nil
}

// bestPoint runs the cube search and returns the answer's distance.
func bestPoint(bots []bot, heap []cube) (int, error) {
	if len(bots) == 0 || cap(heap) == 0 {
		return 0, errors.New("no bots or no heap")
	}
	lo, hi := bots[0].x, bots[0].x
	for _, b := range bots {
		lo, hi = min(lo, b.x, b.y, b.z), max(hi, b.x, b.y, b.z)
	}
	side := 1
	for bit := 0; bit < 62 && side < hi-lo+1; bit++ {
		side *= 2
	}
	if side < hi-lo+1 {
		return 0, fmt.Errorf("the bots spread over more than %d", side)
	}
	start, err := score(cube{x: lo, y: lo, z: lo, side: side}, bots)
	if err != nil {
		return 0, err
	}
	if heap, err = heapPush(heap[:0], start); err != nil {
		return 0, err
	}
	for range maxPops {
		var c cube
		if heap, c, err = heapPop(heap); err != nil {
			return 0, err
		}
		if c.side == 1 {
			return c.origin, nil
		}
		half := c.side / 2
		for k := range 8 {
			sub := cube{x: c.x + half*(k&1), y: c.y + half*(k>>1&1), z: c.z + half*(k>>2&1), side: half}
			if sub, err = score(sub, bots); err != nil {
				return 0, err
			}
			if heap, err = heapPush(heap, sub); err != nil {
				return 0, err
			}
		}
	}
	return 0, fmt.Errorf("search did not finish within %d cubes", maxPops)
}

func day23(path string) error {
	bots, err := readBots(path)
	if err != nil {
		return err
	}
	count, err := inRangeOfStrongest(bots)
	if err != nil {
		return err
	}

	distance, err := bestPoint(bots, make([]cube, 0, maxHeap))
	if err != nil {
		return err
	}
	if distance < 0 {
		return fmt.Errorf("distance %d is negative", distance)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, distance); err != nil {
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
