// Day 8: Two-Factor Authentication
//
// The input is a list of operations on a 50x6 screen of pixels, all off at
// first: "rect AxB" lights the AxB rectangle at the top left, "rotate row y=A
// by B" shifts a row right with wrap-around, and "rotate column x=A by B"
// shifts a column down.
//
// Part 1 is how many pixels are lit at the end.
// Part 2 is the code the screen then shows, in capital letters 5 pixels wide
// (4 of glyph and 1 of spacing) and 6 tall.
//
// The screen is a fixed array; a rotation copies the row or column into a
// small scratch array and writes it back shifted. The letters are read by
// matching each 4x6 cell against the known glyphs of the Advent of Code font;
// a cell that matches none is an error rather than a guess.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	width         = 50
	height        = 6
	glyphWidth    = 5 // including one column of spacing
	letters       = width / glyphWidth
)

type screen [height][width]bool

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// twoNumbers reads "AxB" or "y=A by B" style pairs separated by sep.
func twoNumbers(text, sep string) (a, b int, err error) {
	left, right, found := strings.Cut(text, sep)
	if !found {
		return 0, 0, fmt.Errorf("malformed operands %q", text)
	}
	if a, err = strconv.Atoi(left); err != nil {
		return 0, 0, fmt.Errorf("invalid operand in %q: %w", text, err)
	}
	if b, err = strconv.Atoi(right); err != nil {
		return 0, 0, fmt.Errorf("invalid operand in %q: %w", text, err)
	}
	if a < 0 || b < 0 {
		return 0, 0, fmt.Errorf("negative operand in %q", text)
	}
	return a, b, nil
}

// apply carries out one operation.
func (s *screen) apply(line string) error {
	switch {
	case strings.HasPrefix(line, "rect "):
		w, h, err := twoNumbers(strings.TrimPrefix(line, "rect "), "x")
		if err != nil {
			return err
		}
		if w > width || h > height {
			return fmt.Errorf("rectangle in %q is larger than the screen", line)
		}
		for r := range h {
			for c := range w {
				s[r][c] = true
			}
		}
	case strings.HasPrefix(line, "rotate row y="):
		row, by, err := twoNumbers(strings.TrimPrefix(line, "rotate row y="), " by ")
		if err != nil {
			return err
		}
		if row >= height {
			return fmt.Errorf("row in %q is off the screen", line)
		}
		old := s[row]
		for c := range width {
			s[row][(c+by)%width] = old[c]
		}
	case strings.HasPrefix(line, "rotate column x="):
		col, by, err := twoNumbers(strings.TrimPrefix(line, "rotate column x="), " by ")
		if err != nil {
			return err
		}
		if col >= width {
			return fmt.Errorf("column in %q is off the screen", line)
		}
		var old [height]bool
		for r := range height {
			old[r] = s[r][col]
		}
		for r := range height {
			s[(r+by)%height][col] = old[r]
		}
	default:
		return fmt.Errorf("unknown operation %q", line)
	}
	return nil
}

// countLit counts the pixels that are on.
func (s *screen) countLit() (int, error) {
	lit := 0
	for r := range height {
		for c := range width {
			if s[r][c] {
				lit++
			}
		}
	}
	if lit > width*height {
		return 0, fmt.Errorf("%d pixels lit on a screen of %d", lit, width*height)
	}
	return lit, nil
}

// read decodes the letters on the screen into out.
func (s *screen) read(out []byte) ([]byte, error) {
	if cap(out) < letters {
		return nil, fmt.Errorf("room for %d letters, need %d", cap(out), letters)
	}
	// The font: each letter followed by its six rows of four pixels.
	font := [17]string{
		"A.##.#..##..######..##..#", "B###.#..####.#..##..####.", "C.##.#..##...#...#..#.##.",
		"E#####...###.#...#...####", "F#####...###.#...#...#...", "G.##.#..##...#.###..#.###",
		"H#..##..######..##..##..#", "I.###..#...#...#...#..###", "J..##...#...#...##..#.##.",
		"K#..##.#.##..#.#.#.#.#..#", "L#...#...#...#...#...####", "O.##.#..##..##..##..#.##.",
		"P###.#..##..####.#...#...", "R###.#..##..####.#.#.#..#", "S.####...#....##....####.",
		"U#..##..##..##..##..#.##.", "Z####...#..#..#..#...####",
	}

	out = out[:0]
	var cell [24]byte
	for l := range letters {
		for r := range height {
			for c := range 4 {
				cell[r*4+c] = '.'
				if s[r][l*glyphWidth+c] {
					cell[r*4+c] = '#'
				}
			}
		}
		letter := byte(0)
		for _, g := range font {
			if g[1:] == string(cell[:]) {
				letter = g[0]
			}
		}
		if letter == 0 {
			return nil, fmt.Errorf("letter %d does not match the font: %s", l, cell[:])
		}
		out = append(out, letter)
	}
	return out, nil
}

func day8(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	var s screen
	operations := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if err := s.apply(line); err != nil {
			return err
		}
		operations++
	}

	if operations == 0 {
		return fmt.Errorf("%s: no operations found", path)
	}
	lit, err := s.countLit()
	if err != nil {
		return err
	}
	code, err := s.read(make([]byte, 0, letters))
	if err != nil {
		return err
	}
	if len(code) != letters {
		return fmt.Errorf("read %d letters, expected %d", len(code), letters)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", lit, code); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
