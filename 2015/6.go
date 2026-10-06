// Day 6: Probably a Fire Hazard
//
// The input is a list of instructions for a 1000x1000 grid of lights, each
// "turn on", "turn off" or "toggle" followed by the inclusive rectangle it
// applies to, as "x1,y1 through x2,y2". All lights start off.
//
// Part 1 treats the lights as on/off and counts how many end up lit.
// Part 2 reads the same words as brightness changes - on +1, off -1 (not below
// zero), toggle +2 - and sums the final brightness.
//
// Both grids are applied side by side, cell by cell; the grids are allocated
// once and the instructions only touch the cells inside their rectangles.
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
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	gridSide      = 1000
)

// The three things an instruction can do.
const (
	turnOn = iota
	turnOff
	toggle
)

// instruction is one parsed line: what to do and to which rectangle.
type instruction struct {
	action         int
	x1, y1, x2, y2 int
}

// grids holds both parts' lights, indexed y*gridSide + x.
type grids struct {
	lit        [gridSide * gridSide]bool
	brightness [gridSide * gridSide]int
}

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

// parseCorner reads a corner written as "x,y".
func parseCorner(text string) (x, y int, err error) {
	xText, yText, found := strings.Cut(text, ",")
	if !found {
		return 0, 0, fmt.Errorf("malformed corner %q", text)
	}
	if x, err = strconv.Atoi(xText); err != nil {
		return 0, 0, fmt.Errorf("invalid corner %q: %w", text, err)
	}
	if y, err = strconv.Atoi(yText); err != nil {
		return 0, 0, fmt.Errorf("invalid corner %q: %w", text, err)
	}
	if x < 0 || x >= gridSide || y < 0 || y >= gridSide {
		return 0, 0, fmt.Errorf("corner %q is off the grid", text)
	}
	return x, y, nil
}

// parseInstruction reads a line such as "turn on 0,0 through 999,999".
func parseInstruction(line string) (instruction, error) {
	var in instruction
	var rest string
	switch {
	case strings.HasPrefix(line, "turn on "):
		in.action, rest = turnOn, strings.TrimPrefix(line, "turn on ")
	case strings.HasPrefix(line, "turn off "):
		in.action, rest = turnOff, strings.TrimPrefix(line, "turn off ")
	case strings.HasPrefix(line, "toggle "):
		in.action, rest = toggle, strings.TrimPrefix(line, "toggle ")
	default:
		return instruction{}, fmt.Errorf("unknown instruction %q", line)
	}

	from, to, found := strings.Cut(rest, " through ")
	if !found {
		return instruction{}, fmt.Errorf("malformed rectangle in %q", line)
	}
	var err error
	if in.x1, in.y1, err = parseCorner(from); err != nil {
		return instruction{}, err
	}
	if in.x2, in.y2, err = parseCorner(to); err != nil {
		return instruction{}, err
	}
	if in.x1 > in.x2 || in.y1 > in.y2 {
		return instruction{}, fmt.Errorf("rectangle in %q runs backwards", line)
	}
	return in, nil
}

func parseInput(path string) ([]instruction, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	instructions := make([]instruction, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		in, err := parseInstruction(line)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, in)
	}
	if len(instructions) == 0 {
		return nil, fmt.Errorf("%s: no instructions found", path)
	}
	return instructions, nil
}

// apply carries out one instruction on both grids.
func (g *grids) apply(in instruction) error {
	if in.x1 < 0 || in.y1 < 0 || in.x2 >= gridSide || in.y2 >= gridSide {
		return fmt.Errorf("rectangle %v is off the grid", in)
	}
	if in.action != turnOn && in.action != turnOff && in.action != toggle {
		return fmt.Errorf("unknown action %d", in.action)
	}

	for y := in.y1; y <= in.y2; y++ {
		for x := in.x1; x <= in.x2; x++ {
			i := y*gridSide + x
			switch in.action {
			case turnOn:
				g.lit[i] = true
				g.brightness[i]++
			case turnOff:
				g.lit[i] = false
				g.brightness[i] = max(g.brightness[i]-1, 0)
			case toggle:
				g.lit[i] = !g.lit[i]
				g.brightness[i] += 2
			}
		}
	}
	return nil
}

// totals counts the lit lights and sums the brightness.
func (g *grids) totals() (lit, brightness int, err error) {
	for i := range g.lit {
		if g.brightness[i] < 0 {
			return 0, 0, fmt.Errorf("light %d has negative brightness", i)
		}
		// A light is only lit by "turn on" or "toggle", which both raise
		// its brightness.
		if g.lit[i] && g.brightness[i] < 1 {
			return 0, 0, fmt.Errorf("light %d is lit with brightness %d", i, g.brightness[i])
		}
		if g.lit[i] {
			lit++
		}
		brightness += g.brightness[i]
	}
	if lit > len(g.lit) {
		return 0, 0, fmt.Errorf("%d lights lit out of %d", lit, len(g.lit))
	}
	return lit, brightness, nil
}

func day6(path string) error {
	instructions, err := parseInput(path)
	if err != nil {
		return err
	}

	if len(instructions) == 0 {
		return errors.New("no instructions to follow")
	}
	g := new(grids)
	for _, in := range instructions {
		if err := g.apply(in); err != nil {
			return err
		}
	}
	lit, brightness, err := g.totals()
	if err != nil {
		return err
	}
	if brightness < lit {
		return fmt.Errorf("total brightness %d is below the %d lights lit", brightness, lit)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", lit, brightness); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
