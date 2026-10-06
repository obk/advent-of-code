// Day 15: Warehouse Woes
//
// The input is a warehouse map (walls '#', boxes 'O', the robot '@'),
// then the robot's moves (^ v < >). A move pushes any boxes in the way,
// unless that would push something into a wall.
//
// Part 1 sums 100*row + column over boxes after every move.
// Part 2 first doubles the map's width: '#' becomes "##", 'O' becomes
// "[]", '.' becomes "..", '@' becomes "@.". The sum uses each box's left
// half.
//
// A push gathers the boxes it would move with a queue (a wide box pushed
// up or down can push two more), checks none hits a wall, then moves them
// all, farthest first.
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
	maxInputBytes = 1 << 15
	maxRows       = 64
	maxCols       = 128
	maxPush       = maxRows * maxCols
)

// warehouse is the map, the robot, and the push buffers.
type warehouse struct {
	g      [maxRows][maxCols]byte
	w, h   int
	rx, ry int
	queue  [maxPush][2]int
	queued [maxRows][maxCols]bool
}

// load builds the map, doubled in width if wide.
func (wh *warehouse) load(rows []string, wide bool) error {
	wh.h, wh.w, wh.rx = len(rows), 0, -1
	if wh.h == 0 || wh.h > maxRows {
		return fmt.Errorf("%d rows is out of range", wh.h)
	}
	for y, r := range rows {
		x := 0
		for k := range len(r) {
			cells, n := [2]byte{r[k]}, 1
			if wide {
				n = 2
				switch r[k] {
				case 'O':
					cells = [2]byte{'[', ']'}
				case '@':
					cells = [2]byte{'@', '.'}
				default:
					cells = [2]byte{r[k], r[k]}
				}
			}
			for c := range n {
				if x == maxCols {
					return fmt.Errorf("row %d is too wide", y)
				}
				wh.g[y][x] = cells[c]
				if cells[c] == '@' {
					wh.rx, wh.ry, wh.g[y][x] = x, y, '.'
				}
				x++
			}
		}
		wh.w = x
	}
	if wh.rx < 0 {
		return errors.New("no robot")
	}
	return nil
}

// push moves the robot one step in direction (dx, dy), pushing boxes.
func (wh *warehouse) push(dx, dy int) error {
	if max(dx, -dx)+max(dy, -dy) != 1 {
		return fmt.Errorf("move %d,%d is not one step", dx, dy)
	}
	clear(wh.queued[:])
	wh.queue[0] = [2]int{wh.rx, wh.ry}
	tail := 1
	for head := 0; head < tail; head++ {
		x, y := wh.queue[head][0]+dx, wh.queue[head][1]+dy
		switch c := wh.g[y][x]; c {
		case '#':
			return nil // blocked: nothing moves
		case '.':
			continue
		}
		// The cell ahead holds (part of) a box: queue it, and its other half.
		cells := [2][2]int{{x, y}, {x, y}}
		switch wh.g[y][x] {
		case '[':
			cells[1][0] = x + 1
		case ']':
			cells[1][0] = x - 1
		}
		for _, c := range cells {
			if wh.queued[c[1]][c[0]] {
				continue
			}
			if tail == len(wh.queue) {
				return errors.New("push queue is full")
			}
			wh.queued[c[1]][c[0]], wh.queue[tail], tail = true, c, tail+1
		}
	}
	// Move the queued cells, farthest first, so nothing is overwritten.
	for k := tail - 1; k >= 1; k-- {
		x, y := wh.queue[k][0], wh.queue[k][1]
		wh.g[y+dy][x+dx], wh.g[y][x] = wh.g[y][x], '.'
	}
	wh.rx, wh.ry = wh.rx+dx, wh.ry+dy
	return nil
}

// gps sums 100*row + column over boxes (their left halves).
func (wh *warehouse) gps() int {
	sum := 0
	for y := range wh.h {
		for x := range wh.w {
			if c := wh.g[y][x]; c == 'O' || c == '[' {
				sum += 100*y + x
			}
		}
	}
	return sum
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
	plan, moves, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found {
		return fmt.Errorf("%s: no blank line between map and moves", path)
	}
	moves = strings.Join(strings.Fields(moves), "")
	var answers [2]int
	wh := new(warehouse)
	for k, wide := range [2]bool{false, true} {
		if err := wh.load(strings.Fields(plan), wide); err != nil {
			return err
		}
		for i := range len(moves) {
			d := strings.IndexByte("^>v<", moves[i])
			if d < 0 {
				return fmt.Errorf("%s: unknown move %q", path, moves[i])
			}
			if err := wh.push([4]int{0, 1, 0, -1}[d], [4]int{-1, 0, 1, 0}[d]); err != nil {
				return err
			}
		}
		answers[k] = wh.gps()
		if answers[k] < 1 {
			return errors.New("no boxes in the warehouse")
		}
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
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
