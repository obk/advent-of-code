// Day 16: Reindeer Maze
//
// The input is a maze with start S and end E. The reindeer starts at S
// facing east; a step forward costs 1 and a quarter turn 1000.
//
// Part 1 is the lowest score from S to E.
// Part 2 counts the tiles on at least one lowest-scoring path.
//
// Dijkstra's algorithm over (tile, facing), once forwards from S and once
// backwards from E (in every facing). A state lies on a best path exactly
// when its two distances add up to the best score.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	states        = maxSide * maxSide * 4
	maxHeap       = 4 * states
)

// item is a heap entry.
type item struct {
	cost, state int
}

// maze is the map and the searches' buffers.
type maze struct {
	rows     []string
	w, h     int
	sx, sy   int
	ex, ey   int
	from, to [states]int // distance from the start, and to the end
	heap     []item
	heapLen  int
}

// push adds an item to the heap.
func (m *maze) push(it item) error {
	if m.heapLen == len(m.heap) {
		return errors.New("heap is full")
	}
	if it.cost < 0 {
		return fmt.Errorf("negative cost %d", it.cost)
	}
	i := m.heapLen
	m.heap[i], m.heapLen = it, m.heapLen+1
	for level := 0; level < 64 && i > 0; level++ {
		parent := (i - 1) / 2
		if m.heap[parent].cost <= m.heap[i].cost {
			break
		}
		m.heap[parent], m.heap[i] = m.heap[i], m.heap[parent]
		i = parent
	}
	return nil
}

// pop removes the cheapest item.
func (m *maze) pop() (item, error) {
	if m.heapLen == 0 {
		return item{}, errors.New("heap is empty")
	}
	top := m.heap[0]
	m.heapLen--
	m.heap[0] = m.heap[m.heapLen]
	i := 0
	for level := 0; level < 64; level++ {
		small := i
		for _, k := range [2]int{2*i + 1, 2*i + 2} {
			if k < m.heapLen && m.heap[k].cost < m.heap[small].cost {
				small = k
			}
		}
		if small == i {
			break
		}
		m.heap[small], m.heap[i] = m.heap[i], m.heap[small]
		i = small
	}
	if m.heapLen > 0 && m.heap[0].cost < top.cost {
		return item{}, errors.New("the heap is out of order")
	}
	return top, nil
}

// search fills dist from the seeded states; backwards steps go against the
// facing. Facings: 0 east, 1 south, 2 west, 3 north.
func (m *maze) search(dist *[states]int, backwards bool) error {
	if dist == nil {
		return errors.New("no distance table")
	}
	dx, dy := [4]int{1, 0, -1, 0}, [4]int{0, 1, 0, -1}
	for range maxHeap {
		if m.heapLen == 0 {
			return nil
		}
		it, err := m.pop()
		if err != nil {
			return err
		}
		if it.cost > dist[it.state] {
			continue
		}
		cell, d := it.state/4, it.state%4
		x, y := cell%m.w, cell/m.w
		sign := 1
		if backwards {
			sign = -1
		}
		nx, ny := x+sign*dx[d], y+sign*dy[d]
		if nx < 0 || ny < 0 || nx >= m.w || ny >= m.h {
			return fmt.Errorf("stepped off the maze at %d,%d", nx, ny)
		}
		next := [3]item{{it.cost + 1, (ny*m.w+nx)*4 + d}, {it.cost + 1000, cell*4 + (d+1)%4}, {it.cost + 1000, cell*4 + (d+3)%4}}
		for k, n := range next {
			if k == 0 && m.rows[ny][nx] == '#' {
				continue
			}
			if dist[n.state] < 0 || n.cost < dist[n.state] {
				dist[n.state] = n.cost
				if err := m.push(n); err != nil {
					return err
				}
			}
		}
	}
	return errors.New("the search ran too long")
}

// load reads the maze.
func load(path string) (*maze, error) {
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
	m := &maze{rows: strings.Fields(string(content)), sx: -1, ex: -1, heap: make([]item, maxHeap)}
	m.h = len(m.rows)
	if m.h < 3 || m.h > maxSide || len(m.rows[0]) > maxSide {
		return nil, fmt.Errorf("%s: maze is too small or larger than %d", path, maxSide)
	}
	m.w = len(m.rows[0])
	for y, r := range m.rows {
		if len(r) != m.w || strings.Trim(r, ".#SE") != "" || r[0] != '#' || r[m.w-1] != '#' {
			return nil, fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if x := strings.IndexByte(r, 'S'); x >= 0 {
			m.sx, m.sy = x, y
		}
		if x := strings.IndexByte(r, 'E'); x >= 0 {
			m.ex, m.ey = x, y
		}
	}
	if m.sx < 0 || m.ex < 0 {
		return nil, fmt.Errorf("%s: no S or no E", path)
	}
	return m, nil
}

// solve returns the best score and the number of tiles on best paths.
func (m *maze) solve() (int, int, error) {
	for i := range m.from {
		m.from[i], m.to[i] = -1, -1
	}
	start := (m.sy*m.w + m.sx) * 4 // facing east
	m.from[start], m.heapLen = 0, 0
	if err := m.push(item{0, start}); err != nil {
		return 0, 0, err
	}
	if err := m.search(&m.from, false); err != nil {
		return 0, 0, err
	}
	best := -1
	for d := range 4 {
		end := (m.ey*m.w+m.ex)*4 + d
		if c := m.from[end]; c >= 0 && (best < 0 || c < best) {
			best = c
		}
		m.to[end] = 0
		if err := m.push(item{0, end}); err != nil {
			return 0, 0, err
		}
	}
	if best < 0 {
		return 0, 0, errors.New("E cannot be reached")
	}
	if err := m.search(&m.to, true); err != nil {
		return 0, 0, err
	}
	tiles := 0
	for cell := range m.w * m.h {
		for d := range 4 {
			if s := cell*4 + d; m.from[s] >= 0 && m.to[s] >= 0 && m.from[s]+m.to[s] == best {
				tiles++
				break
			}
		}
	}
	if tiles < 2 {
		return 0, 0, fmt.Errorf("only %d tiles on the best paths", tiles)
	}
	return best, tiles, nil
}

func day16(path string) error {
	m, err := load(path)
	if err != nil {
		return err
	}
	best, tiles, err := m.solve()
	if err != nil {
		return err
	}

	if best < 1 {
		return errors.New("S and E are the same tile")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", best, tiles); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
