// Day 13: Transparent Origami
//
// The input is dots "x,y" on transparent paper, then folds "fold along
// x=5" or "fold along y=7"; a fold reflects the dots past the line onto the
// other side.
//
// Part 1 counts the visible dots after the first fold.
// Part 2 makes every fold and reads the eight capital letters the dots
// spell, each 4 wide (with 1 between) and 6 tall, matched against the
// known glyphs of the Advent of Code font.
//
// Folding maps each dot independently, so the dots are folded in place and
// drawn into a fixed grid to count and read them.
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
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 14
	maxDots       = 2_000
	maxCoord      = 2_000
	letters       = 8
	maxFolds      = 32
)

// fold is a line to fold along: x = at, or y = at.
type fold struct {
	alongX bool
	at     int
}

// load parses the dots and the folds.
func load(path string) ([][2]int, []fold, error) {
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
	dotText, foldText, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	dotLines, foldLines := strings.Fields(dotText), strings.Split(foldText, "\n")
	if !found || len(dotLines) > maxDots || len(foldLines) > maxFolds {
		return nil, nil, fmt.Errorf("%s: malformed sections", path)
	}
	dots := make([][2]int, len(dotLines))
	for i, l := range dotLines {
		a, b, ok := strings.Cut(l, ",")
		x, errX := strconv.Atoi(a)
		y, errY := strconv.Atoi(b)
		if !ok || errX != nil || errY != nil || x < 0 || y < 0 || x > maxCoord || y > maxCoord {
			return nil, nil, fmt.Errorf("%s: invalid dot %q", path, l)
		}
		dots[i] = [2]int{x, y}
	}
	folds := make([]fold, len(foldLines))
	for i, l := range foldLines {
		axis, v, ok := strings.Cut(strings.TrimPrefix(l, "fold along "), "=")
		at, err := strconv.Atoi(v)
		if !ok || err != nil || (axis != "x" && axis != "y") || at < 1 {
			return nil, nil, fmt.Errorf("%s: invalid fold %q", path, l)
		}
		folds[i] = fold{axis == "x", at}
	}
	return dots, folds, nil
}

// apply folds every dot in place.
func apply(dots [][2]int, f fold) error {
	if f.at < 1 {
		return fmt.Errorf("fold line %d is not positive", f.at)
	}
	k := 1
	if f.alongX {
		k = 0
	}
	for i := range dots {
		if dots[i][k] == f.at {
			return fmt.Errorf("a dot lies on the fold line %d", f.at)
		}
		if dots[i][k] > f.at {
			dots[i][k] = 2*f.at - dots[i][k]
		}
		if dots[i][k] < 0 {
			return errors.New("a dot folded past the edge")
		}
	}
	return nil
}

// visible draws the dots into grid (cleared first) and counts them.
func visible(dots [][2]int, grid *[maxCoord + 1][maxCoord + 1]bool) int {
	*grid = [maxCoord + 1][maxCoord + 1]bool{}
	n := 0
	for _, d := range dots {
		if !grid[d[1]][d[0]] {
			grid[d[1]][d[0]], n = true, n+1
		}
	}
	return n
}

// read returns the letters drawn in the top-left of the grid.
func read(grid *[maxCoord + 1][maxCoord + 1]bool, out []byte) ([]byte, error) {
	if cap(out) < letters {
		return nil, errors.New("no room for the letters")
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
		for r := range 6 {
			for c := range 4 {
				cell[r*4+c] = '.'
				if grid[r][l*5+c] {
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

func day13(path string) error {
	dots, folds, err := load(path)
	if err != nil {
		return err
	}
	if len(folds) == 0 {
		return errors.New("no folds")
	}
	grid := new([maxCoord + 1][maxCoord + 1]bool)
	if err := apply(dots, folds[0]); err != nil {
		return err
	}
	first := visible(dots, grid)
	for _, f := range folds[1:] {
		if err := apply(dots, f); err != nil {
			return err
		}
	}
	if remaining := visible(dots, grid); remaining > first {
		return fmt.Errorf("folding raised the dots from %d to %d", first, remaining)
	}
	code, err := read(grid, make([]byte, 0, letters))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", first, code); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
