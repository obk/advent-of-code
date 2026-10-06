// Day 10: Cathode-Ray Tube
//
// Each line of the input is "noop" (one cycle) or "addx V" (two cycles,
// then X += V). X starts at 1.
//
// Part 1 sums cycle * X during cycles 20, 60, ..., 220.
// Part 2 draws a 40x6 screen: during cycle c, the pixel at column
// (c-1) mod 40 is lit if X is within 1 of it. The screen shows eight
// capital letters, read with the known glyphs of the Advent of Code font.
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
	maxInputBytes = 1 << 12
	width         = 40
	height        = 6
	letters       = 8
)

// crt runs the program, returning the signal strength sum and filling the
// screen.
func crt(lines []string, screen *[height][width]bool) (int, error) {
	x, cycle, signal := 1, 0, 0
	for _, l := range lines {
		cycles, add := 1, 0
		if l != "noop" {
			v, err := strconv.Atoi(strings.TrimPrefix(l, "addx "))
			if err != nil || !strings.HasPrefix(l, "addx ") {
				return 0, fmt.Errorf("unknown instruction %q", l)
			}
			cycles, add = 2, v
		}
		for range cycles {
			if cycle >= width*height {
				return 0, errors.New("the program runs past the screen")
			}
			col := cycle % width
			screen[cycle/width][col] = x >= col-1 && x <= col+1
			cycle++
			if cycle%40 == 20 {
				signal += cycle * x
			}
		}
		x += add
	}
	if cycle < width*height {
		return 0, fmt.Errorf("the program stopped after %d cycles", cycle)
	}
	return signal, nil
}

// read returns the letters on the screen.
func read(screen *[height][width]bool, out []byte) ([]byte, error) {
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
		for r := range height {
			for c := range 4 {
				cell[r*4+c] = '.'
				if screen[r][l*5+c] {
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
	var screen [height][width]bool
	signal, err := crt(strings.Split(strings.TrimSpace(string(content)), "\n"), &screen)
	if err != nil {
		return err
	}
	text, err := read(&screen, make([]byte, 0, letters))
	if err != nil {
		return err
	}

	if signal == 0 {
		return errors.New("signal strength is zero")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", signal, text); err != nil {
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
