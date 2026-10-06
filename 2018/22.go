// Day 22: Mode Maze
//
// The input gives a cave's depth and a target position. Each region's
// geologic index is 0 at the mouth and the target, x*16807 along y=0,
// y*48271 along x=0, and otherwise the product of the erosion levels to its
// left and above; its erosion level is (index + depth) mod 20183, and its
// type is the erosion level mod 3 (rocky, wet, narrow).
//
// Part 1 is the total risk (the type, 0..2) of the rectangle from the mouth
// to the target.
// Part 2 is the fewest minutes to reach the target with the torch, moving
// one region a minute and switching tools in seven. Rocky regions allow the
// torch or climbing gear, wet ones gear or neither, narrow ones torch or
// neither.
//
// Numbering the tools neither=0, torch=1, gear=2 makes a tool usable in a
// region exactly when it differs from the region's type. Part 2 is Dijkstra
// over (region, tool) states with a binary heap in a fixed array; entries
// that have gone stale are skipped when popped.
//
// Erosion levels depend only on regions up and to the left, so the table is
// filled row by row in a fixed array large enough to leave room around the
// target.
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
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 10
	margin        = 100 // room beyond the target for any route around it
	maxCoord      = 1_000
	maxDepth      = 100_000
	modulus       = 20183
	switchTime    = 7
	torch         = 1
	maxHeap       = 4 * 3 * (maxCoord + margin) * (maxCoord + margin) / 4
)

// entry is a heap item: a time and a packed state (region*3 + tool).
type entry struct {
	time, state int
}

// cave holds the erosion levels, and Dijkstra's tables.
type cave struct {
	erosion [maxCoord + margin][maxCoord + margin]int
	tx, ty  int
	w, h    int
	best    []int
	heap    []entry
}

// readScan parses "depth: D" and "target: X,Y".
func readScan(path string) (depth, tx, ty int, err error) {
	if path == "" {
		return 0, 0, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, 0, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	f := strings.Fields(strings.NewReplacer("depth:", "", "target:", "", ",", " ").Replace(string(content)))
	if len(f) != 3 {
		return 0, 0, 0, fmt.Errorf("%s: expected depth and target", path)
	}
	var n [3]int
	for i, s := range f {
		if n[i], err = strconv.Atoi(s); err != nil {
			return 0, 0, 0, fmt.Errorf("%s: invalid number %q: %w", path, s, err)
		}
	}
	if n[0] < 0 || n[0] > maxDepth || n[1] < 0 || n[2] < 0 || n[1] > maxCoord || n[2] > maxCoord {
		return 0, 0, 0, fmt.Errorf("%s: depth or target out of range", path)
	}
	return n[0], n[1], n[2], nil
}

// fill computes every erosion level in the cave.
func (c *cave) fill(depth, tx, ty int) error {
	if tx < 0 || ty < 0 || tx >= maxCoord || ty >= maxCoord {
		return fmt.Errorf("target %d,%d is out of range", tx, ty)
	}
	if depth < 0 || depth > maxDepth {
		return fmt.Errorf("depth %d is out of range", depth)
	}
	c.tx, c.ty = tx, ty
	c.w, c.h = tx+margin, ty+margin
	for y := range c.h {
		for x := range c.w {
			var index int
			switch {
			case (x == 0 && y == 0) || (x == tx && y == ty):
				index = 0
			case y == 0:
				index = x * 16807
			case x == 0:
				index = y * 48271
			default:
				index = c.erosion[y][x-1] * c.erosion[y-1][x]
			}
			c.erosion[y][x] = (index + depth) % modulus
		}
	}
	return nil
}

// risk sums the region types from the mouth to the target.
func (c *cave) risk() (int, error) {
	if c.w == 0 || c.h == 0 {
		return 0, errors.New("cave is not filled")
	}
	if c.tx >= c.w || c.ty >= c.h {
		return 0, fmt.Errorf("target %d,%d lies outside the filled cave", c.tx, c.ty)
	}
	total := 0
	for y := 0; y <= c.ty; y++ {
		for x := 0; x <= c.tx; x++ {
			total += c.erosion[y][x] % 3
		}
	}
	return total, nil
}

// push adds an entry to the heap.
func (c *cave) push(e entry) error {
	if len(c.heap) == cap(c.heap) {
		return fmt.Errorf("heap of %d entries is full", cap(c.heap))
	}
	if e.time < 0 {
		return fmt.Errorf("negative time %d", e.time)
	}
	c.heap = append(c.heap, e)
	// A heap of n entries is under 64 levels deep.
	for i, level := len(c.heap)-1, 0; i > 0 && level < 64; level++ {
		parent := (i - 1) / 2
		if c.heap[parent].time <= c.heap[i].time {
			break
		}
		c.heap[parent], c.heap[i] = c.heap[i], c.heap[parent]
		i = parent
	}
	return nil
}

// pop removes and returns the earliest entry.
func (c *cave) pop() (entry, error) {
	if len(c.heap) == 0 {
		return entry{}, errors.New("heap is empty")
	}
	top := c.heap[0]
	last := len(c.heap) - 1
	c.heap[0], c.heap = c.heap[last], c.heap[:last]
	if top.state < 0 {
		return entry{}, fmt.Errorf("popped invalid state %d", top.state)
	}
	for i, level := 0, 0; 2*i+1 < len(c.heap) && level < 64; level++ {
		child := 2*i + 1
		if child+1 < len(c.heap) && c.heap[child+1].time < c.heap[child].time {
			child++
		}
		if c.heap[i].time <= c.heap[child].time {
			break
		}
		c.heap[i], c.heap[child] = c.heap[child], c.heap[i]
		i = child
	}
	return top, nil
}

// relax records a quicker way into a state and queues it.
func (c *cave) relax(state, time int) error {
	if time < 0 {
		return fmt.Errorf("negative time %d", time)
	}
	if state < 0 || state >= len(c.best) {
		return fmt.Errorf("state %d is out of range", state)
	}
	if c.best[state] >= 0 && c.best[state] <= time {
		return nil
	}
	c.best[state] = time
	return c.push(entry{time, state})
}

// rescue runs Dijkstra from the mouth with the torch to the target with the
// torch.
func (c *cave) rescue() (int, error) {
	if c.w == 0 || len(c.best) != c.w*c.h*3 {
		return 0, errors.New("cave or tables are not set up")
	}
	for i := range c.best {
		c.best[i] = -1
	}
	c.heap = c.heap[:0]
	goal := (c.ty*c.w+c.tx)*3 + torch
	if err := c.relax(torch, 0); err != nil {
		return 0, err
	}
	for range maxHeap {
		e, err := c.pop()
		if err != nil {
			return 0, err
		}
		if e.time > c.best[e.state] {
			continue // stale
		}
		if e.state == goal {
			return e.time, nil
		}
		region, tool := e.state/3, e.state%3
		x, y := region%c.w, region/c.w
		kind := c.erosion[y][x] % 3
		if kind == tool {
			return 0, fmt.Errorf("state %d holds a tool its region forbids", e.state)
		}
		if err := c.relax(region*3+3-kind-tool, e.time+switchTime); err != nil {
			return 0, err // the one other tool this region allows
		}
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= c.w || ny >= c.h || c.erosion[ny][nx]%3 == tool {
				continue
			}
			if err := c.relax((ny*c.w+nx)*3+tool, e.time+1); err != nil {
				return 0, err
			}
		}
	}
	return 0, errors.New("the target was not reached")
}

func day22(path string) error {
	depth, tx, ty, err := readScan(path)
	if err != nil {
		return err
	}
	c := new(cave)
	if err := c.fill(depth, tx, ty); err != nil {
		return err
	}
	total, err := c.risk()
	if err != nil {
		return err
	}

	if c.w < 1 || c.h < 1 {
		return fmt.Errorf("cave of %dx%d is empty", c.w, c.h)
	}
	c.best = make([]int, c.w*c.h*3)
	c.heap = make([]entry, 0, 6*c.w*c.h)
	minutes, err := c.rescue()
	if err != nil {
		return err
	}
	if minutes < tx+ty {
		return fmt.Errorf("%d minutes is quicker than walking straight to %d,%d", minutes, tx, ty)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, minutes); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
