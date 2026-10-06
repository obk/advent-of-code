// Day 18: Lavaduct Lagoon
//
// Each line of the input is a dig instruction: a direction (U D L R), a
// distance, and a colour "(#70c710)". The trench is a closed loop, then
// its inside is dug out too.
//
// Part 1 is the lagoon's volume (in cubic metres, one per square).
// Part 2 reads the colour instead: its first five hex digits are the
// distance and the last is the direction (0 R, 1 D, 2 L, 3 U).
//
// The shoelace formula gives the area enclosed by the trench's centre
// line; the trench squares themselves add half the perimeter plus one.
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
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 16
	maxSteps      = 1_000
	maxDistance   = 1 << 20
)

// lagoon returns the volume dug by the plan, reading the colours if hex.
func lagoon(lines []string, hex bool) (int, error) {
	if len(lines) == 0 || len(lines) > maxSteps {
		return 0, fmt.Errorf("%d steps is out of range", len(lines))
	}
	x, y, twice, perimeter := 0, 0, 0, 0
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) != 3 || len(f[2]) != 9 {
			return 0, fmt.Errorf("malformed step %q", l)
		}
		dir := strings.Index("RDLU", f[0])
		n, err := strconv.Atoi(f[1])
		if hex {
			v, errH := strconv.ParseInt(f[2][2:8], 16, 64)
			dir, n, err = int(v&15), int(v>>4), errH
		}
		if err != nil || dir < 0 || dir > 3 || n < 1 || n > maxDistance {
			return 0, fmt.Errorf("invalid step %q", l)
		}
		nx, ny := x+n*[4]int{1, 0, -1, 0}[dir], y+n*[4]int{0, 1, 0, -1}[dir]
		twice += x*ny - nx*y
		perimeter += n
		x, y = nx, ny
	}
	if x != 0 || y != 0 {
		return 0, fmt.Errorf("the trench ends at %d,%d, not where it began", x, y)
	}
	return max(twice, -twice)/2 + perimeter/2 + 1, nil
}

func day18(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	plain, err := lagoon(lines, false)
	if err != nil {
		return err
	}
	big, err := lagoon(lines, true)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", plain, big); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
