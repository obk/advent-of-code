// Day 10: Hoof It
//
// The input is a topographic map of heights 0-9. A hiking trail climbs
// from 0 to 9 one level per step (up, down, left or right).
//
// Part 1 sums, over trailheads (height 0), how many 9s each can reach.
// Part 2 sums how many distinct trails start at each trailhead.
//
// Part 2 counts trails to the 9s level by level, from 9 down. Part 1
// floods from each trailhead, stamping visits with the trailhead number.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 14
	maxSide       = 64
	cells         = maxSide * maxSide
)

// island is the map and the searches' buffers.
type island struct {
	rows  []string
	w, h  int
	stamp [cells]int
	ways  [cells]int
	queue [cells]int
}

// reach counts the 9s reachable from (x, y), stamping visits with tag.
func (m *island) reach(x, y, tag int) (int, error) {
	if tag < 1 || m.rows[y][x] != '0' {
		return 0, errors.New("not a trailhead, or a bad tag")
	}
	m.queue[0], m.stamp[y*m.w+x] = y*m.w+x, tag
	tail, nines := 1, 0
	for head := 0; head < tail; head++ {
		cx, cy := m.queue[head]%m.w, m.queue[head]/m.w
		if m.rows[cy][cx] == '9' {
			nines++
			continue
		}
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || ny < 0 || nx >= m.w || ny >= m.h || m.stamp[ny*m.w+nx] == tag || m.rows[ny][nx] != m.rows[cy][cx]+1 {
				continue
			}
			m.stamp[ny*m.w+nx] = tag
			m.queue[tail], tail = ny*m.w+nx, tail+1
		}
	}
	return nines, nil
}

// trails counts, for every cell, the climbs from it to a 9.
func (m *island) trails() error {
	if m.w*m.h > cells {
		return errors.New("map is larger than the table")
	}
	for level := byte('9'); level >= '0'; level-- {
		for y := range m.h {
			for x := range m.w {
				if m.rows[y][x] != level {
					continue
				}
				if level == '9' {
					m.ways[y*m.w+x] = 1
					continue
				}
				n := 0
				for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
					nx, ny := x+d[0], y+d[1]
					if nx >= 0 && ny >= 0 && nx < m.w && ny < m.h && m.rows[ny][nx] == level+1 {
						n += m.ways[ny*m.w+nx]
					}
				}
				if n < 0 {
					return errors.New("trail count overflowed")
				}
				m.ways[y*m.w+x] = n
			}
		}
	}
	return nil
}

func day10(path string) error {
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
	m := &island{rows: strings.Fields(string(content))}
	m.h = len(m.rows)
	if m.h == 0 || m.h > maxSide || len(m.rows[0]) > maxSide {
		return fmt.Errorf("%s: map is empty or larger than %d", path, maxSide)
	}
	m.w = len(m.rows[0])
	for y, r := range m.rows {
		if len(r) != m.w || strings.Trim(r, "0123456789") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	if err := m.trails(); err != nil {
		return err
	}
	score, rating, tag := 0, 0, 0
	for y := range m.h {
		for x := range m.w {
			if m.rows[y][x] != '0' {
				continue
			}
			tag++
			n, err := m.reach(x, y, tag)
			if err != nil {
				return err
			}
			score, rating = score+n, rating+m.ways[y*m.w+x]
		}
	}

	// Every reachable 9 has at least one trail to it.
	if rating < score {
		return fmt.Errorf("%d trails but %d reachable peaks", rating, score)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", score, rating); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
