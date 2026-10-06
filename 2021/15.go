// Day 15: Chiton
//
// The input is a grid of risk levels 1-9. Moving into a cell costs its
// risk.
//
// Part 1 is the least total risk from the top left to the bottom right.
// Part 2 tiles the grid 5x5; each tile right or down adds 1 to the risks,
// wrapping 9 back to 1. The answer is the same least total risk.
//
// Dijkstra's algorithm with a fixed-capacity binary heap.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 16
	maxSide       = 128
	tiles         = 5
	maxCells      = maxSide * maxSide * tiles * tiles
	maxHeap       = 4 * maxCells
)

// item is a heap entry.
type item struct {
	dist, cell int
}

// cave is the grid and the search's buffers.
type cave struct {
	rows    []string
	w, h    int
	dist    []int
	heap    []item
	heapLen int
}

// risk returns the risk of (x, y) in the tiled grid.
func (c *cave) risk(x, y int) (int, error) {
	if x < 0 || y < 0 || x >= tiles*c.w || y >= tiles*c.h {
		return 0, fmt.Errorf("%d,%d is outside the cave", x, y)
	}
	r := int(c.rows[y%c.h][x%c.w]-'0') + x/c.w + y/c.h
	return (r-1)%9 + 1, nil
}

// push adds an item to the heap.
func (c *cave) push(it item) error {
	if c.heapLen == len(c.heap) {
		return errors.New("heap is full")
	}
	if it.dist < 0 {
		return fmt.Errorf("negative distance %d", it.dist)
	}
	i := c.heapLen
	c.heap[i], c.heapLen = it, c.heapLen+1
	for level := 0; level < 64 && i > 0; level++ {
		parent := (i - 1) / 2
		if c.heap[parent].dist <= c.heap[i].dist {
			break
		}
		c.heap[parent], c.heap[i] = c.heap[i], c.heap[parent]
		i = parent
	}
	return nil
}

// pop removes the nearest item.
func (c *cave) pop() (item, error) {
	if c.heapLen == 0 {
		return item{}, errors.New("heap is empty")
	}
	top := c.heap[0]
	c.heapLen--
	c.heap[0] = c.heap[c.heapLen]
	i := 0
	for level := 0; level < 64; level++ {
		small := i
		for _, k := range [2]int{2*i + 1, 2*i + 2} {
			if k < c.heapLen && c.heap[k].dist < c.heap[small].dist {
				small = k
			}
		}
		if small == i {
			break
		}
		c.heap[small], c.heap[i] = c.heap[i], c.heap[small]
		i = small
	}
	if c.heapLen > 0 && c.heap[0].dist < top.dist {
		return item{}, errors.New("the heap is out of order")
	}
	return top, nil
}

// safest returns the least risk across the grid tiled `times` times.
func (c *cave) safest(times int) (int, error) {
	if times < 1 || times > tiles {
		return 0, fmt.Errorf("cannot tile %d times", times)
	}
	w, h := c.w*times, c.h*times
	if len(c.dist) < w*h {
		return 0, errors.New("distance table is too small")
	}
	for i := range w * h {
		c.dist[i] = -1
	}
	c.heapLen = 0
	c.dist[0] = 0
	if err := c.push(item{0, 0}); err != nil {
		return 0, err
	}
	for range maxHeap {
		it, err := c.pop()
		if err != nil {
			return 0, errors.New("the far corner cannot be reached")
		}
		if it.dist > c.dist[it.cell] {
			continue
		}
		x, y := it.cell%w, it.cell/w
		if x == w-1 && y == h-1 {
			return it.dist, nil
		}
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				continue
			}
			r, err := c.risk(nx, ny)
			if err != nil {
				return 0, err
			}
			nd := it.dist + r
			if c.dist[ny*w+nx] < 0 || nd < c.dist[ny*w+nx] {
				c.dist[ny*w+nx] = nd
				if err := c.push(item{nd, ny*w + nx}); err != nil {
					return 0, err
				}
			}
		}
	}
	return 0, errors.New("the search ran too long")
}

func day15(path string) error {
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
	if len(rows) == 0 || len(rows) > maxSide || len(rows[0]) > maxSide {
		return fmt.Errorf("%s: grid is empty or larger than %d", path, maxSide)
	}
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, "123456789") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	c := &cave{rows: rows, w: len(rows[0]), h: len(rows), dist: make([]int, maxCells), heap: make([]item, maxHeap)}
	small, err := c.safest(1)
	if err != nil {
		return err
	}
	big, err := c.safest(tiles)
	if err != nil {
		return err
	}

	if small < 1 || big < 1 {
		return fmt.Errorf("risks %d and %d are not plausible", small, big)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", small, big); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
