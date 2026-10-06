// Day 18: Many-Worlds Interpretation
//
// The input is a vault map: walls '#', open '.', the entrance '@', keys
// 'a'..'z' and doors 'A'..'Z'; a door opens once its key is held.
//
// Part 1 is the fewest steps to collect every key.
// Part 2 walls off the 3x3 square around the entrance, leaving entrances at
// its four corners, one robot at each; the robots move one at a time and
// share keys. The answer is again the fewest total steps.
//
// A breadth-first search from the entrance and from each key gives the
// distance to every other key and the doors on the way. Dijkstra's
// algorithm then searches states (where each robot stands, which keys are
// held) packed into one number, with a fixed-capacity heap.
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
	maxInputBytes = 1 << 14
	maxSide       = 96
	keys          = 26
	maxRobots     = 4
	points        = keys + maxRobots // keys, then robot starts
	maxStates     = 1 << 21
	maxHeap       = 1 << 22
)

// item is a heap entry: a state and its distance.
type item struct {
	dist  int
	state uint64
}

// vault holds the map, the key-to-key graph and the search's buffers.
type vault struct {
	grid      [][]byte
	pos       [points][2]int
	present   [points]bool
	robots    int
	allKeys   uint64
	dist      [points][points]int
	doors     [points][points]uint64
	seen      [maxSide * maxSide]int
	need      [maxSide * maxSide]uint64
	queue     [maxSide * maxSide]int
	heap      []item
	best      map[uint64]int
	heapLen   int
	stateBits int
}

// load parses the map and finds the keys and the entrance.
func load(path string) (*vault, error) {
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
	rows := strings.Fields(string(content))
	if len(rows) < 3 || len(rows) > maxSide {
		return nil, fmt.Errorf("%s: %d rows, expected 3..%d", path, len(rows), maxSide)
	}
	v := &vault{grid: make([][]byte, len(rows))}
	for y, row := range rows {
		v.grid[y] = []byte(row)
	}
	for y, row := range v.grid {
		if len(row) != len(v.grid[0]) || len(row) > maxSide {
			return nil, fmt.Errorf("%s: row %d is not %d wide", path, y, len(v.grid[0]))
		}
		for x := range len(row) {
			if c := row[x]; c >= 'a' && c <= 'z' {
				v.pos[c-'a'], v.present[c-'a'] = [2]int{x, y}, true
				v.allKeys |= 1 << (c - 'a')
			}
		}
	}
	v.heap = make([]item, maxHeap)
	v.best = make(map[uint64]int, maxStates)
	return v, nil
}

// placeRobots records every '@' as a robot start.
func (v *vault) placeRobots() error {
	v.robots = 0
	for y, row := range v.grid {
		for x := range len(row) {
			if row[x] != '@' {
				continue
			}
			if v.robots == maxRobots {
				return fmt.Errorf("more than %d entrances", maxRobots)
			}
			v.pos[keys+v.robots], v.present[keys+v.robots] = [2]int{x, y}, true
			v.robots++
		}
	}
	if v.robots == 0 {
		return errors.New("no entrance")
	}
	return nil
}

// explore searches outward from point p, filling its distances and doors to
// every key (-1 where unreachable).
func (v *vault) explore(p int) error {
	if p < 0 || p >= points || !v.present[p] {
		return fmt.Errorf("point %d is not on the map", p)
	}
	w := len(v.grid[0])
	for i := range v.seen {
		v.seen[i] = -1
	}
	for k := range keys {
		v.dist[p][k] = -1
	}
	if v.grid[v.pos[p][1]][v.pos[p][0]] == '#' {
		return fmt.Errorf("point %d is inside a wall", p)
	}
	start := v.pos[p][1]*w + v.pos[p][0]
	v.seen[start], v.need[start], v.queue[0] = 0, 0, start
	for head, tail := 0, 1; head < tail; head++ {
		at := v.queue[head]
		for _, d := range [4]int{-w, w, -1, 1} {
			next := at + d
			if next < 0 || next >= w*len(v.grid) || v.seen[next] >= 0 {
				continue
			}
			c := v.grid[next/w][next%w]
			if c == '#' {
				continue
			}
			if tail == len(v.queue) {
				return errors.New("search queue is full")
			}
			v.seen[next], v.need[next] = v.seen[at]+1, v.need[at]
			if c >= 'A' && c <= 'Z' {
				v.need[next] |= 1 << (c - 'A')
			}
			if c >= 'a' && c <= 'z' {
				v.dist[p][c-'a'], v.doors[p][c-'a'] = v.seen[next], v.need[next]
			}
			v.queue[tail], tail = next, tail+1
		}
	}
	return nil
}

// split turns the single entrance into four, walled off from each other.
func (v *vault) split() error {
	if err := v.placeRobots(); err != nil {
		return err
	}
	if v.robots != 1 {
		return fmt.Errorf("%d entrances; the map is already split", v.robots)
	}
	x, y := v.pos[keys][0], v.pos[keys][1]
	if x < 2 || y < 2 || y+2 >= len(v.grid) || x+2 >= len(v.grid[0]) {
		return errors.New("the entrance is too close to the edge to split")
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if c := v.grid[y+dy][x+dx]; c != '.' && c != '@' {
				return fmt.Errorf("%c next to the entrance; it cannot be split", c)
			}
			v.grid[y+dy][x+dx] = '#'
			if dx != 0 && dy != 0 {
				v.grid[y+dy][x+dx] = '@'
			}
		}
	}
	return nil
}

func day18(path string) error {
	v, err := load(path)
	if err != nil {
		return err
	}
	steps, err := v.collect()
	if err != nil {
		return err
	}

	if steps < 1 {
		return fmt.Errorf("collecting the keys took %d steps", steps)
	}
	if err := v.split(); err != nil {
		return err
	}
	four, err := v.collect()
	if err != nil {
		return err
	}
	if four < 1 {
		return fmt.Errorf("four robots took %d steps", four)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, four); err != nil {
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

// A state packs the keys held (bits 0..25) and each robot's point (5 bits
// each, from bit 26).

// robotAt returns robot r's point in a state.
func robotAt(state uint64, r int) (int, error) {
	if r < 0 || r >= maxRobots {
		return 0, fmt.Errorf("robot %d does not exist", r)
	}
	p := int(state>>(keys+5*r)) & 31
	if p >= points {
		return 0, fmt.Errorf("robot %d is at point %d", r, p)
	}
	return p, nil
}

// push adds an item to the heap.
func (v *vault) push(it item) error {
	if v.heapLen == len(v.heap) {
		return fmt.Errorf("heap is full at %d items", len(v.heap))
	}
	if it.dist < 0 {
		return fmt.Errorf("item at negative distance %d", it.dist)
	}
	i := v.heapLen
	v.heap[i], v.heapLen = it, v.heapLen+1
	for level := 0; level < 64 && i > 0; level++ {
		parent := (i - 1) / 2
		if v.heap[parent].dist <= v.heap[i].dist {
			break
		}
		v.heap[parent], v.heap[i] = v.heap[i], v.heap[parent]
		i = parent
	}
	return nil
}

// pop removes the nearest item from the heap.
func (v *vault) pop() (item, error) {
	if v.heapLen == 0 {
		return item{}, errors.New("heap is empty")
	}
	top := v.heap[0]
	v.heapLen--
	v.heap[0] = v.heap[v.heapLen]
	i := 0
	for level := 0; level < 64; level++ {
		small := i
		for _, c := range [2]int{2*i + 1, 2*i + 2} {
			if c < v.heapLen && v.heap[c].dist < v.heap[small].dist {
				small = c
			}
		}
		if small == i {
			break
		}
		v.heap[small], v.heap[i] = v.heap[i], v.heap[small]
		i = small
	}
	if v.heapLen > 0 && v.heap[0].dist < top.dist {
		return item{}, errors.New("the heap is out of order")
	}
	return top, nil
}

// relax offers a state at a distance, queueing it if it is the best yet.
func (v *vault) relax(state uint64, dist int) error {
	if dist < 0 {
		return fmt.Errorf("negative distance %d", dist)
	}
	if d, ok := v.best[state]; ok && d <= dist {
		return nil
	}
	if _, ok := v.best[state]; !ok && len(v.best) == maxStates {
		return fmt.Errorf("more than %d states", maxStates)
	}
	v.best[state] = dist
	return v.push(item{dist, state})
}

// collect returns the fewest steps for the robots to gather every key.
func (v *vault) collect() (int, error) {
	if err := v.placeRobots(); err != nil {
		return 0, err
	}
	if v.allKeys == 0 {
		return 0, errors.New("the vault has no keys")
	}
	for p := range points {
		if v.present[p] {
			if err := v.explore(p); err != nil {
				return 0, err
			}
		}
	}
	clear(v.best)
	v.heapLen = 0
	var start uint64
	for r := range v.robots {
		start |= uint64(keys+r) << (keys + 5*r)
	}
	if err := v.relax(start, 0); err != nil {
		return 0, err
	}
	for range maxHeap {
		it, err := v.pop()
		if err != nil {
			return 0, errors.New("the keys cannot all be reached")
		}
		if v.best[it.state] < it.dist {
			continue // a shorter way here was already found
		}
		held := it.state & (1<<keys - 1)
		if held == v.allKeys {
			return it.dist, nil
		}
		if err := v.expand(it, held); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("the search popped more than %d items", maxHeap)
}

// expand relaxes every move of one robot to a key it can reach next.
func (v *vault) expand(it item, held uint64) error {
	if held&^v.allKeys != 0 {
		return fmt.Errorf("state holds keys %b that do not exist", held)
	}
	for r := range v.robots {
		at, err := robotAt(it.state, r)
		if err != nil {
			return err
		}
		for k := range keys {
			d := v.dist[at][k]
			if d < 0 || held&(1<<k) != 0 || v.doors[at][k]&^held != 0 {
				continue
			}
			next := it.state&^(31<<(keys+5*r)) | uint64(k)<<(keys+5*r) | 1<<k
			if err := v.relax(next, it.dist+d); err != nil {
				return err
			}
		}
	}
	return nil
}
