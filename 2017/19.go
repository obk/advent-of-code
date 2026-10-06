// Day 19: A Series of Tubes
//
// The input is a routing diagram of '|', '-' and '+' lines with letters on
// them. A packet enters from the top on the only line there, goes straight
// through crossings, turns only at '+', and stops where the line ends.
//
// Part 1 is the letters it passes, in order.
// Part 2 is how many steps it takes, counting the first cell.
//
// The walk is a loop over at most every cell of the diagram, since a path
// that visits more cells than the diagram has must be going round in circles.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxLetters    = 26
)

// diagram is the padded grid of the routing diagram.
type diagram struct {
	rows  []string
	width int
}

// readDiagram returns the diagram's lines, keeping leading spaces.
func readDiagram(path string) (diagram, error) {
	if path == "" {
		return diagram{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return diagram{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return diagram{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	rows := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(rows) > maxLines {
		return diagram{}, fmt.Errorf("%s: %d lines, more than %d", path, len(rows), maxLines)
	}
	if len(rows) == 0 || rows[0] == "" {
		return diagram{}, fmt.Errorf("%s: empty diagram", path)
	}
	d := diagram{rows: rows}
	for _, r := range rows {
		d.width = max(d.width, len(r))
	}
	return d, nil
}

// at returns the character at (r, c), or a space off the diagram.
func (d diagram) at(r, c int) byte {
	if r < 0 || r >= len(d.rows) || c < 0 || c >= len(d.rows[r]) {
		return ' '
	}
	return d.rows[r][c]
}

// follow walks the path, writing the letters seen into out, and returns them
// and the number of steps taken (counting the first cell).
func (d diagram) follow(out []byte) ([]byte, int, error) {
	if len(d.rows) == 0 {
		return nil, 0, errors.New("empty diagram")
	}
	if d.width == 0 {
		return nil, 0, errors.New("diagram has no width")
	}
	start := strings.IndexByte(d.rows[0], '|')
	if start < 0 {
		return nil, 0, errors.New("no line enters from the top")
	}
	r, c, dr, dc := 0, start, 1, 0
	out = out[:0]
	for steps := 1; steps <= len(d.rows)*d.width; steps++ {
		ch := d.at(r, c)
		if ch >= 'A' && ch <= 'Z' {
			if len(out) == cap(out) {
				return nil, 0, fmt.Errorf("more than %d letters", cap(out))
			}
			out = append(out, ch)
		}
		if ch == '+' {
			// Turn to whichever side continues the line.
			if d.at(r+dc, c+dr) != ' ' {
				dr, dc = dc, dr
			} else {
				dr, dc = -dc, -dr
			}
		}
		r, c = r+dr, c+dc
		if d.at(r, c) == ' ' {
			return out, steps, nil
		}
	}
	return nil, 0, errors.New("the path goes round in circles")
}

func day19(path string) error {
	d, err := readDiagram(path)
	if err != nil {
		return err
	}
	letters, steps, err := d.follow(make([]byte, 0, maxLetters))
	if err != nil {
		return err
	}

	if len(letters) == 0 {
		return errors.New("the path passes no letters")
	}
	if steps < len(letters) {
		return fmt.Errorf("%d steps cannot pass %d letters", steps, len(letters))
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", letters, steps); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
