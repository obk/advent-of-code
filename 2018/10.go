// Day 10: The Stars Align
//
// Each line of the input is a point of light with a position and a velocity:
// "position=< 9,  1> velocity=< 0,  2>". Every second each point moves by
// its velocity, and at one moment they line up to spell a message.
//
// Part 1 is the message.
// Part 2 is how many seconds it takes to appear.
//
// The message appears when the points are packed into the shortest band, so
// the time is found by stepping until the height of their bounding box stops
// shrinking. The letters (6 wide, 10 tall, 8 apart) are then read by
// matching each cell against known glyphs of this font; a cell that matches
// none is an error rather than a guess.
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
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxTime       = 1_000_000
	glyphW        = 6
	glyphH        = 10
	pitch         = 8 // glyph width plus spacing
	maxLetters    = 16
)

// light is one point of light.
type light struct {
	x, y, vx, vy int
}

// readLights parses the points.
func readLights(path string) ([]light, error) {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 || len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d points, more than %d", path, len(lines), maxLines)
	}
	lights := make([]light, len(lines))
	clean := strings.NewReplacer("position=<", " ", "velocity=<", " ", ">", " ", ",", " ")
	for i, line := range lines {
		f := strings.Fields(clean.Replace(line))
		if len(f) != 4 {
			return nil, fmt.Errorf("%s: malformed point %q", path, line)
		}
		var n [4]int
		for k, s := range f {
			if n[k], err = strconv.Atoi(s); err != nil {
				return nil, fmt.Errorf("%s: invalid number in %q: %w", path, line, err)
			}
		}
		if max(n[0], -n[0], n[1], -n[1]) > 1_000_000 || max(n[2], -n[2], n[3], -n[3]) > 1_000 {
			return nil, fmt.Errorf("%s: point %q is too far out or too fast", path, line)
		}
		lights[i] = light{n[0], n[1], n[2], n[3]}
	}
	return lights, nil
}

// bounds returns the bounding box of the points at time t.
func bounds(lights []light, t int) (minX, minY, maxX, maxY int, err error) {
	if len(lights) == 0 {
		return 0, 0, 0, 0, errors.New("no points")
	}
	minX, minY, maxX, maxY = 1<<62, 1<<62, -1<<62, -1<<62
	for _, l := range lights {
		x, y := l.x+l.vx*t, l.y+l.vy*t
		minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
	}
	return minX, minY, maxX, maxY, nil
}

// alignTime returns the time at which the bounding box is shortest.
func alignTime(lights []light) (int, error) {
	if len(lights) == 0 {
		return 0, errors.New("no points")
	}
	_, y0, _, y1, err := bounds(lights, 0)
	if err != nil {
		return 0, err
	}
	for t := 1; t <= maxTime; t++ {
		_, a, _, b, err := bounds(lights, t)
		if err != nil {
			return 0, err
		}
		if b-a > y1-y0 {
			return t - 1, nil
		}
		y0, y1 = a, b
	}
	return 0, fmt.Errorf("the points did not converge within %d seconds", maxTime)
}

// read decodes the message at time t into out.
func read(lights []light, t int, out []byte) ([]byte, error) {
	if t < 0 || t > maxTime {
		return nil, fmt.Errorf("time %d is outside 0..%d", t, maxTime)
	}
	minX, minY, maxX, maxY, err := bounds(lights, t)
	if err != nil {
		return nil, err
	}
	if maxY-minY+1 != glyphH {
		return nil, fmt.Errorf("message is %d tall, expected %d", maxY-minY+1, glyphH)
	}
	letters := (maxX - minX + pitch - glyphW + 1) / pitch
	if letters < 1 || letters > cap(out) {
		return nil, fmt.Errorf("message of width %d holds %d letters", maxX-minX+1, letters)
	}
	var grid [glyphH][maxLetters * pitch]byte
	for r := range glyphH {
		for c := range letters * pitch {
			grid[r][c] = '.'
		}
	}
	for _, l := range lights {
		grid[l.y+l.vy*t-minY][l.x+l.vx*t-minX] = '#'
	}
	font := [6]string{
		"E#######.....#.....#.....#####.#.....#.....#.....#.....######",
		"G.####.#....##.....#.....#.....#..####....##....##...##.###.#",
		"H#....##....##....##....########....##....##....##....##....#",
		"J...###....#.....#.....#.....#.....#.....#.#...#.#...#..###..",
		"K#....##...#.#..#..#.#...##....##....#.#...#..#..#...#.#....#",
		"Z######.....#.....#....#....#....#....#....#.....#.....######",
	}
	out = out[:0]
	var cell [glyphW * glyphH]byte
	for k := range letters {
		for r := range glyphH {
			copy(cell[r*glyphW:], grid[r][k*pitch:k*pitch+glyphW])
		}
		letter := byte(0)
		for _, g := range font {
			if g[1:] == string(cell[:]) {
				letter = g[0]
			}
		}
		if letter == 0 {
			return nil, fmt.Errorf("letter %d does not match the known glyphs: %s", k, cell[:])
		}
		out = append(out, letter)
	}
	return out, nil
}

func day10(path string) error {
	lights, err := readLights(path)
	if err != nil {
		return err
	}
	t, err := alignTime(lights)
	if err != nil {
		return err
	}
	message, err := read(lights, t, make([]byte, 0, maxLetters))
	if err != nil {
		return err
	}

	if len(message) == 0 || t < 0 {
		return fmt.Errorf("no message (%q) or negative time %d", message, t)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", message, t); err != nil {
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
