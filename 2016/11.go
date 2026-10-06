// Day 11: Radioisotope Thermoelectric Generators
//
// The input describes four floors holding generators and microchips, one of
// each per element. An elevator carrying you and one or two items moves one
// floor per step. A chip is fried if it shares a floor with another element's
// generator while its own generator is elsewhere.
//
// Part 1 is the fewest steps to bring everything to the fourth floor.
// Part 2 adds two more elements, generator and chip both on the first floor.
//
// This is a breadth-first search over states. Elements are interchangeable:
// what matters is only the multiset of (generator floor, chip floor) pairs,
// so a state is the elevator floor plus those pairs sorted, packed into one
// integer. That cuts the state count enough for a plain BFS, whose queue and
// visited set are sized up front.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 12
	floors        = 4
	maxElements   = 7
	maxStates     = 1 << 21
)

// state packs the elevator floor (bits 0-1) and, from bit 2 up, one 4-bit
// pair per element: generator floor in the high two bits, chip floor in the
// low two.
type state uint32

// search holds the BFS's queue and visited set.
type search struct {
	queue   []state
	depth   []int
	visited map[state]bool
}

// parseInput returns each element's generator and chip floor.
func parseInput(path string) (gens, chips []int, err error) {
	if path == "" {
		return nil, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != floors {
		return nil, nil, fmt.Errorf("%s: %d floors, expected %d", path, len(lines), floors)
	}
	names := make([]string, 0, maxElements)
	gens, chips = make([]int, maxElements), make([]int, maxElements)
	var seenGen, seenChip [maxElements]bool
	for floor, line := range lines {
		words := strings.Fields(strings.NewReplacer(",", " ", ".", " ").Replace(line))
		for i := 1; i < len(words); i++ {
			if words[i] != "generator" && words[i] != "microchip" {
				continue
			}
			name := strings.TrimSuffix(words[i-1], "-compatible")
			id := -1
			for j, n := range names {
				if n == name {
					id = j
				}
			}
			if id < 0 {
				if len(names) == maxElements {
					return nil, nil, fmt.Errorf("%s: more than %d elements", path, maxElements)
				}
				id, names = len(names), append(names, name)
			}
			if words[i] == "generator" {
				gens[id], seenGen[id] = floor, true
			} else {
				chips[id], seenChip[id] = floor, true
			}
		}
	}
	if len(names) == 0 {
		return nil, nil, fmt.Errorf("%s: no elements found", path)
	}
	for id, name := range names {
		if !seenGen[id] || !seenChip[id] {
			return nil, nil, fmt.Errorf("%s: %s lacks a generator or a chip", path, name)
		}
	}
	// The slices keep their capacity, so part 2 can append without growing.
	return gens[:len(names)], chips[:len(names)], nil
}

// pack builds the canonical state: the pairs sorted, so equivalent
// arrangements of interchangeable elements pack to the same number.
func pack(elevator int, gens, chips []int) (state, error) {
	if elevator < 0 || elevator >= floors || len(gens) != len(chips) || len(gens) > maxElements {
		return 0, fmt.Errorf("cannot pack elevator %d with %d/%d elements", elevator, len(gens), len(chips))
	}
	var pairs [maxElements]uint32
	n := len(gens)
	for i := range n {
		if gens[i] < 0 || gens[i] >= floors || chips[i] < 0 || chips[i] >= floors {
			return 0, fmt.Errorf("element %d is on floor %d/%d", i, gens[i], chips[i])
		}
		pairs[i] = uint32(gens[i]<<2 | chips[i])
	}
	for i := 1; i < n; i++ { // insertion sort of at most maxElements pairs
		for j := i; j > 0 && pairs[j] < pairs[j-1]; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
	s := state(elevator)
	for i := range n {
		s |= state(pairs[i]) << (2 + 4*i)
	}
	return s, nil
}

// unpack splits a state into the elevator floor and the floor of every item:
// items 2i and 2i+1 are element i's generator and chip.
func unpack(s state, n int, items []int) (int, error) {
	if n < 1 || n > maxElements || len(items) < 2*n {
		return 0, fmt.Errorf("cannot unpack %d elements into %d slots", n, len(items))
	}
	for i := range n {
		pair := int(s>>(2+4*i)) & 0xF
		items[2*i], items[2*i+1] = pair>>2, pair&3
	}
	return int(s & 3), nil
}

// safe reports whether no chip is fried.
func safe(items []int, n int) bool {
	if n < 1 || len(items) < 2*n {
		return false
	}
	var hasGen [floors]bool
	for i := range n {
		hasGen[items[2*i]] = true
	}
	for i := range n {
		if items[2*i+1] != items[2*i] && hasGen[items[2*i+1]] {
			return false
		}
	}
	return true
}

// tryMove moves items a and b (b may equal a) with the elevator to floor
// next, and queues the result at tail if it is safe and new. It returns the
// new tail.
func (sr *search) tryMove(items []int, n, a, b, next, depth, tail int) (int, error) {
	if tail >= len(sr.queue) {
		return 0, fmt.Errorf("more than %d states to explore", len(sr.queue))
	}
	if a < 0 || b < 0 || a >= 2*n || b >= 2*n {
		return 0, fmt.Errorf("items %d and %d are out of range", a, b)
	}
	if next < 0 || next >= floors {
		return tail, nil
	}

	from := items[a]
	items[a], items[b] = next, next
	ok := safe(items, n)
	var gens, chips [maxElements]int
	for i := range n {
		gens[i], chips[i] = items[2*i], items[2*i+1]
	}
	items[a], items[b] = from, from
	if !ok {
		return tail, nil
	}

	s, err := pack(next, gens[:n], chips[:n])
	if err != nil {
		return 0, err
	}
	if !sr.visited[s] {
		sr.visited[s] = true
		sr.queue[tail], sr.depth[tail] = s, depth
		tail++
	}
	return tail, nil
}

// shortest runs the BFS from start until everything is on the top floor.
func (sr *search) shortest(start state, n int) (int, error) {
	if n < 1 || n > maxElements {
		return 0, fmt.Errorf("%d elements is outside 1..%d", n, maxElements)
	}
	if len(sr.queue) == 0 || len(sr.queue) != len(sr.depth) {
		return 0, errors.New("search queue is not set up")
	}
	var goal state = floors - 1
	for i := range n {
		goal |= state((floors-1)<<2|(floors-1)) << (2 + 4*i)
	}

	clear(sr.visited)
	sr.visited[start] = true
	sr.queue[0], sr.depth[0] = start, 0
	var items [2 * maxElements]int
	tail := 1
	for head := 0; head < tail; head++ {
		if sr.queue[head] == goal {
			return sr.depth[head], nil
		}
		elevator, err := unpack(sr.queue[head], n, items[:])
		if err != nil {
			return 0, err
		}
		for a := range 2 * n {
			for b := a; b < 2*n; b++ {
				if items[a] != elevator || items[b] != elevator {
					continue
				}
				for _, next := range [2]int{elevator - 1, elevator + 1} {
					if tail, err = sr.tryMove(items[:], n, a, b, next, sr.depth[head]+1, tail); err != nil {
						return 0, err
					}
				}
			}
		}
	}
	return 0, errors.New("the top floor cannot be reached")
}

func day11(path string) error {
	gens, chips, err := parseInput(path)
	if err != nil {
		return err
	}
	if len(gens) != len(chips) {
		return fmt.Errorf("%d generators for %d chips", len(gens), len(chips))
	}
	start, err := pack(0, gens, chips)
	if err != nil {
		return err
	}

	sr := search{
		queue:   make([]state, maxStates),
		depth:   make([]int, maxStates),
		visited: make(map[state]bool, maxStates),
	}
	steps, err := sr.shortest(start, len(gens))
	if err != nil {
		return err
	}

	if len(gens)+2 > cap(gens) || len(chips)+2 > cap(chips) {
		return fmt.Errorf("no room for two more elements beside %d", len(gens))
	}
	gens, chips = append(gens, 0, 0), append(chips, 0, 0)
	if start, err = pack(0, gens, chips); err != nil {
		return err
	}
	more, err := sr.shortest(start, len(gens))
	if err != nil {
		return err
	}

	if more < steps {
		return fmt.Errorf("more items took fewer steps (%d < %d)", more, steps)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, more); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day11(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
