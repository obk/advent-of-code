// Day 12: Hill Climbing Algorithm
//
// The input is a heightmap of letters a-z, with the start S (height a)
// and the goal E (height z). A step may climb at most one level, and
// descend any amount.
//
// Part 1 is the fewest steps from S to E.
// Part 2 is the fewest steps from any square of height a to E.
//
// One breadth-first search runs backwards from E (where a step may descend
// at most one level), giving the distance from every square at once.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 16
	maxCells      = 1 << 14
)

// height returns a square's height, with S as a and E as z.
func height(c byte) int {
	switch c {
	case 'S':
		return 0
	case 'E':
		return 25
	}
	return int(c - 'a')
}

// distances fills dist with the steps from each square to E (-1 if none).
func distances(rows []string, end int, dist, queue []int) error {
	w, h := len(rows[0]), len(rows)
	if end < 0 || end >= w*h || len(dist) < w*h || len(queue) < w*h {
		return errors.New("goal or buffers out of range")
	}
	for i := range dist[:w*h] {
		dist[i] = -1
	}
	dist[end], queue[0] = 0, end
	tail := 1
	for head := 0; head < tail; head++ {
		x, y := queue[head]%w, queue[head]/w
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= w || ny >= h || dist[ny*w+nx] >= 0 {
				continue
			}
			// Backwards: the step from (nx, ny) up to (x, y) climbs at most one.
			if height(rows[y][x])-height(rows[ny][nx]) > 1 {
				continue
			}
			dist[ny*w+nx] = dist[y*w+x] + 1
			queue[tail], tail = ny*w+nx, tail+1
		}
	}
	return nil
}

func day12(path string) error {
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
	if len(rows) == 0 || len(rows)*len(rows[0]) > maxCells {
		return fmt.Errorf("%s: map is empty or larger than %d squares", path, maxCells)
	}
	start, end := -1, -1
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, "abcdefghijklmnopqrstuvwxyzSE") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if x := strings.IndexByte(row, 'S'); x >= 0 {
			start = y*len(row) + x
		}
		if x := strings.IndexByte(row, 'E'); x >= 0 {
			end = y*len(row) + x
		}
	}
	if start < 0 || end < 0 {
		return fmt.Errorf("%s: no S or no E", path)
	}
	dist, queue := make([]int, maxCells), make([]int, maxCells)
	if err := distances(rows, end, dist, queue); err != nil {
		return err
	}
	best := -1
	for i := range len(rows) * len(rows[0]) {
		if height(rows[i/len(rows[0])][i%len(rows[0])]) == 0 && dist[i] >= 0 && (best < 0 || dist[i] < best) {
			best = dist[i]
		}
	}
	if dist[start] < 0 {
		return errors.New("E cannot be reached from S")
	}
	// S is itself at height a, so the best start is no farther.
	if best < 0 || best > dist[start] {
		return fmt.Errorf("best start at %d steps, but S is %d", best, dist[start])
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", dist[start], best); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
