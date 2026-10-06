// Day 23: Amphipod
//
// The input is a burrow: a hallway of 11 spaces over four side rooms, two
// deep, holding amphipods A-D. Each must reach its own room (A leftmost);
// steps cost 1, 10, 100 and 1000 energy for A-D. An amphipod moves at
// most twice: from its room to a hallway space not just outside a room,
// then into its own room once that holds only its own kind.
//
// Part 1 is the least energy to sort them.
// Part 2 inserts the rows "#D#C#B#A#" and "#D#B#A#C#" into the middle of
// the rooms, making them four deep, and asks again.
//
// Dijkstra's algorithm over burrow states. A state packs the 11 hallway
// spaces and up to 16 room spaces, each 0-4, into one base-5 number. The
// heap and the distance map are allocated once with fixed capacities.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 8
	hall          = 11
	rooms         = 4
	maxDepth      = 4
	maxStates     = 1 << 20
	maxHeap       = 1 << 22
)

// burrow is one arrangement; 0 is empty, 1-4 are A-D.
type burrow struct {
	hall  [hall]int
	room  [rooms][maxDepth]int // index 0 is the top
	depth int
}

// encode packs a burrow into a number.
func (b *burrow) encode() (uint64, error) {
	if b.depth < 1 || b.depth > maxDepth {
		return 0, fmt.Errorf("rooms %d deep cannot be packed", b.depth)
	}
	var v uint64
	for _, c := range b.hall {
		v = v*5 + uint64(c)
	}
	for r := range rooms {
		for i := range b.depth {
			if b.room[r][i] < 0 || b.room[r][i] > 4 {
				return 0, fmt.Errorf("room %d holds %d", r, b.room[r][i])
			}
			v = v*5 + uint64(b.room[r][i])
		}
	}
	return v, nil
}

// decode unpacks a number into b (whose depth must be set).
func (b *burrow) decode(v uint64) error {
	if b.depth < 1 || b.depth > maxDepth {
		return fmt.Errorf("rooms %d deep cannot be unpacked", b.depth)
	}
	for r := rooms - 1; r >= 0; r-- {
		for i := b.depth - 1; i >= 0; i-- {
			b.room[r][i], v = int(v%5), v/5
		}
	}
	for h := hall - 1; h >= 0; h-- {
		b.hall[h], v = int(v%5), v/5
	}
	if v != 0 {
		return fmt.Errorf("state has %d left over after unpacking", v)
	}
	return nil
}

// item is a heap entry.
type item struct {
	cost  int
	state uint64
}

// search holds Dijkstra's heap and best costs.
type search struct {
	heap    []item
	heapLen int
	best    map[uint64]int
}

// push adds an item to the heap.
func (s *search) push(it item) error {
	if s.heapLen == len(s.heap) {
		return errors.New("heap is full")
	}
	if it.cost < 0 {
		return fmt.Errorf("negative cost %d", it.cost)
	}
	i := s.heapLen
	s.heap[i], s.heapLen = it, s.heapLen+1
	for level := 0; level < 64 && i > 0; level++ {
		parent := (i - 1) / 2
		if s.heap[parent].cost <= s.heap[i].cost {
			break
		}
		s.heap[parent], s.heap[i] = s.heap[i], s.heap[parent]
		i = parent
	}
	return nil
}

// pop removes the cheapest item.
func (s *search) pop() (item, error) {
	if s.heapLen == 0 {
		return item{}, errors.New("heap is empty")
	}
	top := s.heap[0]
	s.heapLen--
	s.heap[0] = s.heap[s.heapLen]
	i := 0
	for level := 0; level < 64; level++ {
		small := i
		for _, k := range [2]int{2*i + 1, 2*i + 2} {
			if k < s.heapLen && s.heap[k].cost < s.heap[small].cost {
				small = k
			}
		}
		if small == i {
			break
		}
		s.heap[small], s.heap[i] = s.heap[i], s.heap[small]
		i = small
	}
	if s.heapLen > 0 && s.heap[0].cost < top.cost {
		return item{}, errors.New("the heap is out of order")
	}
	return top, nil
}

// relax records a state at a cost if it is the best yet.
func (s *search) relax(state uint64, cost int) error {
	if cost < 0 {
		return fmt.Errorf("negative cost %d", cost)
	}
	if c, ok := s.best[state]; ok && c <= cost {
		return nil
	}
	if _, ok := s.best[state]; !ok && len(s.best) == maxStates {
		return fmt.Errorf("more than %d states", maxStates)
	}
	s.best[state] = cost
	return s.push(item{cost, state})
}

// clear reports whether hallway spaces strictly between from and to (and
// to itself) are empty.
func (b *burrow) clear(from, to int) (bool, error) {
	if from < 0 || to < 0 || from >= hall || to >= hall || from == to {
		return false, fmt.Errorf("cannot walk from %d to %d", from, to)
	}
	step := 1
	if to < from {
		step = -1
	}
	for h := from + step; h != to+step; h += step {
		if b.hall[h] != 0 {
			return false, nil
		}
	}
	return true, nil
}

// settled reports whether room r holds only its own kind from index i down.
func (b *burrow) settled(r, i int) (bool, error) {
	if r < 0 || r >= rooms || i < 0 {
		return false, fmt.Errorf("room %d from %d does not exist", r, i)
	}
	for k := i; k < b.depth; k++ {
		if b.room[r][k] != 0 && b.room[r][k] != r+1 {
			return false, nil
		}
	}
	return true, nil
}

// expand relaxes every move from one state.
func (s *search) expand(b *burrow, it item) error {
	energy := [5]int{0, 1, 10, 100, 1000}
	if b.depth < 1 || b.depth > maxDepth {
		return fmt.Errorf("rooms %d deep", b.depth)
	}
	for r := range rooms { // out of a room into the hallway
		door, top := 2+2*r, b.depth
		for k := b.depth - 1; k >= 0; k-- {
			if b.room[r][k] != 0 {
				top = k
			}
		}
		done, err := b.settled(r, min(top, b.depth-1))
		if err != nil {
			return err
		}
		if top == b.depth || done {
			continue
		}
		kind := b.room[r][top]
		for _, h := range [7]int{0, 1, 3, 5, 7, 9, 10} {
			if err := s.step(b, &b.room[r][top], &b.hall[h], it.cost+(top+1+max(h-door, door-h))*energy[kind], door, h); err != nil {
				return err
			}
		}
	}
	for h, kind := range b.hall { // from the hallway into its own room
		if kind < 0 || kind > 4 {
			return fmt.Errorf("hallway space %d holds %d", h, kind)
		}
		if kind == 0 {
			continue
		}
		r := kind - 1
		ok, err := b.settled(r, 0)
		if err != nil {
			return err
		}
		bottom := -1
		for k := range b.depth {
			if b.room[r][k] == 0 {
				bottom = k
			}
		}
		if !ok || bottom < 0 {
			continue
		}
		door := 2 + 2*r
		if err := s.step(b, &b.hall[h], &b.room[r][bottom], it.cost+(bottom+1+max(h-door, door-h))*energy[kind], h, door); err != nil {
			return err
		}
	}
	return nil
}

// step moves the amphipod in cell from to the empty cell to, if the
// hallway between hall spaces a and b is clear, relaxes the new state and
// moves it back.
func (s *search) step(b *burrow, from, to *int, cost, a, c int) error {
	ok, err := b.clear(a, c)
	if err != nil || !ok {
		return err
	}
	if *from == 0 || *to != 0 {
		return errors.New("a move must take an amphipod into an empty space")
	}
	*from, *to = 0, *from
	state, err := b.encode()
	*to, *from = 0, *to
	if err != nil {
		return err
	}
	return s.relax(state, cost)
}

// organise returns the least energy to sort the burrow.
func (s *search) organise(start burrow) (int, error) {
	if start.depth < 1 || start.depth > maxDepth {
		return 0, fmt.Errorf("rooms %d deep are not supported", start.depth)
	}
	if len(s.heap) == 0 || s.best == nil {
		return 0, errors.New("no search buffers")
	}
	goal := burrow{depth: start.depth}
	for r := range rooms {
		for i := range start.depth {
			goal.room[r][i] = r + 1
		}
	}
	target, err := goal.encode()
	if err != nil {
		return 0, err
	}
	clear(s.best)
	s.heapLen = 0
	first, err := start.encode()
	if err != nil {
		return 0, err
	}
	if err := s.relax(first, 0); err != nil {
		return 0, err
	}
	b := burrow{depth: start.depth}
	for range maxHeap {
		it, err := s.pop()
		if err != nil {
			return 0, errors.New("the amphipods cannot be sorted")
		}
		if s.best[it.state] < it.cost {
			continue
		}
		if it.state == target {
			return it.cost, nil
		}
		if err := b.decode(it.state); err != nil {
			return 0, err
		}
		if err := s.expand(&b, it); err != nil {
			return 0, err
		}
	}
	return 0, errors.New("the search ran too long")
}

// readBurrow parses the diagram, adding extra rows into the rooms' middle.
func readBurrow(text string, extra []string) (burrow, error) {
	var b burrow
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) != 5 {
		return b, fmt.Errorf("burrow has %d lines, expected 5", len(lines))
	}
	if lines[1] != "#...........#" {
		return b, fmt.Errorf("hallway %q is not empty", lines[1])
	}
	rows := make([]string, 0, maxDepth)
	rows = append(rows, lines[2])
	rows = append(rows, extra...)
	rows = append(rows, lines[3])
	b.depth = len(rows)
	for i, row := range rows {
		if len(row) < 10 {
			return b, fmt.Errorf("room row %q is too short", row)
		}
		for r := range rooms {
			c := row[3+2*r]
			if c < 'A' || c > 'D' {
				return b, fmt.Errorf("room row %q has %q where an amphipod should be", row, c)
			}
			b.room[r][i] = int(c-'A') + 1
		}
	}
	var count [rooms + 1]int
	for r := range rooms {
		for i := range b.depth {
			count[b.room[r][i]]++
		}
	}
	for kind := 1; kind <= rooms; kind++ {
		if count[kind] != b.depth {
			return b, fmt.Errorf("%d amphipods of kind %c, expected %d", count[kind], 'A'+kind-1, b.depth)
		}
	}
	return b, nil
}

func day23(path string) error {
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
	s := &search{heap: make([]item, maxHeap), best: make(map[uint64]int, maxStates)}
	var answers [2]int
	for k, extra := range [2][]string{nil, {"  #D#C#B#A#", "  #D#B#A#C#"}} {
		b, err := readBurrow(string(content), extra)
		if err != nil {
			return err
		}
		if answers[k], err = s.organise(b); err != nil {
			return err
		}
	}

	if answers[0] < 1 || answers[1] < 1 {
		return errors.New("the burrow was already sorted")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
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
