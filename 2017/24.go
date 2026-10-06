// Day 24: Electromagnetic Moat
//
// Each line of the input is a component with two ports, "3/7". A bridge is a
// chain of components, each joined to the next by matching ports, starting
// from a port of 0. Its strength is the sum of all its port numbers.
//
// Part 1 is the strength of the strongest bridge.
// Part 2 is the strength of the longest bridge, the strongest on a tie.
//
// Every bridge is explored depth first with an explicit stack. A frame holds
// the open port at the end, the set of components used (a bitmask), the
// strength and length so far, and which component to try next.
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
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 16
	maxParts      = 64 // one bit each in the used mask
	maxPort       = 1_000
	maxSteps      = 100_000_000
)

// part is one component.
type part struct {
	a, b int
}

// frame is one bridge being extended.
type frame struct {
	port     int
	used     uint64
	strength int
	length   int
	next     int // the next component to try
}

// readParts parses the components.
func readParts(path string) ([]part, error) {
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
	lines := strings.Fields(string(content))
	if len(lines) == 0 || len(lines) > maxParts {
		return nil, fmt.Errorf("%s: %d components, expected 1..%d", path, len(lines), maxParts)
	}
	parts := make([]part, len(lines))
	for i, line := range lines {
		a, b, found := strings.Cut(line, "/")
		x, errA := strconv.Atoi(a)
		y, errB := strconv.Atoi(b)
		if !found || errA != nil || errB != nil || x < 0 || y < 0 || x > maxPort || y > maxPort {
			return nil, fmt.Errorf("%s: invalid component %q", path, line)
		}
		parts[i] = part{x, y}
	}
	zero := false
	for _, p := range parts {
		zero = zero || p.a == 0 || p.b == 0
	}
	if !zero {
		return nil, fmt.Errorf("%s: no component has a 0 port to start from", path)
	}
	return parts, nil
}

// strongest explores every bridge and returns the greatest strength, and
// the strength of the longest (strongest on a tie).
func strongest(parts []part) (best, longest int, err error) {
	if len(parts) == 0 || len(parts) > maxParts {
		return 0, 0, fmt.Errorf("%d components is outside 1..%d", len(parts), maxParts)
	}
	var stack [maxParts + 1]frame
	top, length := 0, 0
	for range maxSteps {
		if top < 0 {
			return best, longest, nil
		}
		f := &stack[top]
		best = max(best, f.strength)
		if f.length > length || (f.length == length && f.strength > longest) {
			length, longest = f.length, f.strength
		}
		if f.next >= len(parts) {
			top--
			continue
		}
		i := f.next
		f.next++
		p := parts[i]
		if f.used>>i&1 == 1 || (p.a != f.port && p.b != f.port) {
			continue
		}
		other := p.a + p.b - f.port // the far port of the component
		if top+1 > maxParts {
			return 0, 0, errors.New("bridge longer than the number of components")
		}
		stack[top+1] = frame{port: other, used: f.used | 1<<i, strength: f.strength + p.a + p.b, length: f.length + 1}
		top++
	}
	return 0, 0, fmt.Errorf("search did not finish within %d steps", maxSteps)
}

func day24(path string) error {
	parts, err := readParts(path)
	if err != nil {
		return err
	}
	best, longest, err := strongest(parts)
	if err != nil {
		return err
	}

	if longest > best {
		return fmt.Errorf("the longest bridge (%d) beats the strongest (%d)", longest, best)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", best, longest); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
