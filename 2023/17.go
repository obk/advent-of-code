// Day 17: Clumsy Crucible
//
// The input is a grid of heat-loss digits. A crucible goes from the top
// left to the bottom right, losing the heat of each block it enters. It
// cannot reverse, and must turn after a run of straight moves.
//
// Part 1 allows runs of 1 to 3 blocks.
// Part 2 (ultra crucibles) needs runs of 4 to 10 blocks, including the
// last run into the corner.
// Each answer is the least heat loss.
//
// Each move goes straight for an allowed number of blocks and then turns,
// so a state is a block and whether the last run was horizontal or
// vertical. Dijkstra's algorithm with a fixed-capacity heap.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	maxStates     = maxSide * maxSide * 2
	maxHeap       = 16 * maxStates
)

// item is a heap entry.
type item struct {
	cost, state int
}

// city is the grid and the search's buffers.
type city struct {
	rows    []string
	w, h    int
	dist    [maxStates]int
	heap    []item
	heapLen int
}

// push adds an item to the heap.
func (c *city) push(it item) error {
	if c.heapLen == len(c.heap) {
		return errors.New("heap is full")
	}
	if it.cost < 0 {
		return fmt.Errorf("negative cost %d", it.cost)
	}
	i := c.heapLen
	c.heap[i], c.heapLen = it, c.heapLen+1
	for level := 0; level < 64 && i > 0; level++ {
		parent := (i - 1) / 2
		if c.heap[parent].cost <= c.heap[i].cost {
			break
		}
		c.heap[parent], c.heap[i] = c.heap[i], c.heap[parent]
		i = parent
	}
	return nil
}

// pop removes the cheapest item.
func (c *city) pop() (item, error) {
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
			if k < c.heapLen && c.heap[k].cost < c.heap[small].cost {
				small = k
			}
		}
		if small == i {
			break
		}
		c.heap[small], c.heap[i] = c.heap[i], c.heap[small]
		i = small
	}
	if c.heapLen > 0 && c.heap[0].cost < top.cost {
		return item{}, errors.New("the heap is out of order")
	}
	return top, nil
}

// least returns the least heat loss with runs of lo..hi blocks.
func (c *city) least(lo, hi int) (int, error) {
	if lo < 1 || hi < lo || hi > maxSide {
		return 0, fmt.Errorf("runs of %d..%d are not allowed", lo, hi)
	}
	for i := range c.dist {
		c.dist[i] = -1
	}
	c.heapLen = 0
	for _, s := range [2]int{0, 1} { // at the start, either way is open
		c.dist[s] = 0
		if err := c.push(item{0, s}); err != nil {
			return 0, err
		}
	}
	for range maxHeap {
		it, err := c.pop()
		if err != nil {
			return 0, errors.New("the corner cannot be reached")
		}
		if it.cost > c.dist[it.state] {
			continue
		}
		cell, vertical := it.state/2, it.state%2 == 1
		x, y := cell%c.w, cell/c.w
		if x == c.w-1 && y == c.h-1 {
			return it.cost, nil
		}
		// Turn: a horizontal run follows a vertical one, and vice versa.
		turned := 1 // the new run is vertical
		if vertical {
			turned = 0
		}
		for _, sign := range [2]int{1, -1} {
			cost, nx, ny := it.cost, x, y
			for k := 1; k <= hi; k++ {
				if vertical {
					nx += sign
				} else {
					ny += sign
				}
				if nx < 0 || ny < 0 || nx >= c.w || ny >= c.h {
					break
				}
				cost += int(c.rows[ny][nx] - '0')
				next := (ny*c.w+nx)*2 + turned
				if k >= lo && (c.dist[next] < 0 || cost < c.dist[next]) {
					c.dist[next] = cost
					if err := c.push(item{cost, next}); err != nil {
						return 0, err
					}
				}
			}
		}
	}
	return 0, errors.New("the search ran too long")
}

func day17(path string) error {
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
	for y, r := range rows {
		if len(r) != len(rows[0]) || strings.Trim(r, "123456789") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	c := &city{rows: rows, w: len(rows[0]), h: len(rows), heap: make([]item, maxHeap)}
	normal, err := c.least(1, 3)
	if err != nil {
		return err
	}
	ultra, err := c.least(4, 10)
	if err != nil {
		return err
	}

	if normal < 1 || ultra < 1 {
		return errors.New("a crucible lost no heat")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", normal, ultra); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
