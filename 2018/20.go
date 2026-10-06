// Day 20: A Regular Map
//
// The input is a regular expression of directions, "^ENWWW(NEEE|SSE(EE|N))$",
// whose routes pass through every door of a grid of rooms. A group gives
// options; every option starts from the room where the group starts.
//
// Part 1 is the largest number of doors on the shortest path to any room.
// Part 2 counts the rooms at least 1000 doors away.
//
// The expression is walked once with an explicit stack of the rooms where
// the open groups started: '(' pushes the current room, '|' goes back to it,
// ')' pops it. Each step records a door on both rooms it joins, in a fixed
// grid centred on the start. A breadth-first search over the doors then
// gives every room's distance.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 20
	reach         = 200 // rooms lie within this many steps of the start each way
	side          = 2*reach + 1
	maxDepth      = 10_000
)

// Door bits: north, east, south, west.
const (
	north = 1 << iota
	east
	south
	west
)

// facility is the grid of rooms with their doors, and search space.
type facility struct {
	doors [side * side]byte
	dist  [side * side]int
	queue [side * side]int
	stack [maxDepth]int
}

// readRegex returns the expression without its ^ and $.
func readRegex(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	text := strings.TrimSpace(string(content))
	if !strings.HasPrefix(text, "^") || !strings.HasSuffix(text, "$") {
		return "", fmt.Errorf("%s: expression must be wrapped in ^ and $", path)
	}
	return text[1 : len(text)-1], nil
}

// trace walks the expression and records every door.
func (f *facility) trace(regex string) error {
	pos, top := (reach)*side+reach, 0
	steps := [4]struct {
		dir   byte
		delta int
		door  byte
		back  byte
	}{{'N', -side, north, south}, {'E', 1, east, west}, {'S', side, south, north}, {'W', -1, west, east}}
	for i := range len(regex) {
		switch c := regex[i]; c {
		case '(':
			if top == maxDepth {
				return fmt.Errorf("groups nested deeper than %d", maxDepth)
			}
			f.stack[top], top = pos, top+1
		case '|':
			if top == 0 {
				return fmt.Errorf("'|' outside a group at %d", i)
			}
			pos = f.stack[top-1]
		case ')':
			if top == 0 {
				return fmt.Errorf("unmatched ')' at %d", i)
			}
			pos, top = f.stack[top-1], top-1
		default:
			k := strings.IndexByte("NESW", c)
			if k < 0 {
				return fmt.Errorf("unexpected %q at %d", c, i)
			}
			next := pos + steps[k].delta
			x, y := next%side, next/side
			if x <= 0 || y <= 0 || x >= side-1 || y >= side-1 {
				return fmt.Errorf("route leaves the %dx%d grid", side, side)
			}
			f.doors[pos] |= steps[k].door
			f.doors[next] |= steps[k].back
			pos = next
		}
	}
	if top != 0 {
		return fmt.Errorf("%d groups left open", top)
	}
	return nil
}

// distances runs a BFS from the start and returns the furthest distance and
// how many rooms are at least far away.
func (f *facility) distances(far int) (furthest, distant int, err error) {
	for i := range f.dist {
		f.dist[i] = -1
	}
	start := reach*side + reach
	f.dist[start], f.queue[0] = 0, start
	tail := 1
	deltas := [4]int{-side, 1, side, -1}
	for head := 0; head < tail; head++ {
		room := f.queue[head]
		furthest = max(furthest, f.dist[room])
		if f.dist[room] >= far {
			distant++
		}
		for k, d := range deltas {
			next := room + d
			if f.doors[room]>>k&1 == 0 || f.dist[next] >= 0 {
				continue
			}
			if tail == len(f.queue) {
				return 0, 0, errors.New("search queue overflowed")
			}
			f.dist[next], f.queue[tail] = f.dist[room]+1, next
			tail++
		}
	}
	if distant > tail {
		return 0, 0, fmt.Errorf("%d distant rooms out of %d", distant, tail)
	}
	return furthest, distant, nil
}

func day20(path string) error {
	regex, err := readRegex(path)
	if err != nil {
		return err
	}
	f := new(facility)
	if err := f.trace(regex); err != nil {
		return err
	}
	furthest, distant, err := f.distances(1000)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", furthest, distant); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
