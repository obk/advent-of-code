// Day 22: Reactor Reboot
//
// Each line of the input turns a cuboid of cubes on or off:
// "on x=10..12,y=10..12,z=10..12". All cubes start off.
//
// Part 1 counts cubes left on within -50..50 on every axis.
// Part 2 counts every cube left on.
//
// The count is kept as a list of signed cuboids whose volumes add up to
// the answer. For each new step, every cuboid already listed that overlaps
// it gets its overlap added with the opposite sign (cancelling that part),
// and an "on" step then adds itself. The list is allocated once with a
// fixed capacity.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 16
	maxSteps      = 1_000
	maxCuboids    = 1 << 20
	maxCoord      = 1_000_000
)

// cuboid is an inclusive box with a sign (+1 or -1); on is used for steps.
type cuboid struct {
	lo, hi [3]int
	sign   int
	on     bool
}

// overlap returns the intersection of a and b, if any.
func overlap(a, b cuboid) (cuboid, bool) {
	var c cuboid
	for k := range 3 {
		c.lo[k], c.hi[k] = max(a.lo[k], b.lo[k]), min(a.hi[k], b.hi[k])
		if c.lo[k] > c.hi[k] {
			return c, false
		}
	}
	return c, true
}

// readSteps parses the reboot steps.
func readSteps(path string) ([]cuboid, error) {
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
	if len(lines) > maxSteps {
		return nil, fmt.Errorf("%s: more than %d steps", path, maxSteps)
	}
	steps := make([]cuboid, len(lines))
	for i, l := range lines {
		state, box, _ := strings.Cut(l, " ")
		c := &steps[i]
		n, err := fmt.Sscanf(box, "x=%d..%d,y=%d..%d,z=%d..%d", &c.lo[0], &c.hi[0], &c.lo[1], &c.hi[1], &c.lo[2], &c.hi[2])
		if err != nil || n != 6 || (state != "on" && state != "off") {
			return nil, fmt.Errorf("%s: malformed step %q", path, l)
		}
		for k := range 3 {
			if c.lo[k] > c.hi[k] || c.lo[k] < -maxCoord || c.hi[k] > maxCoord {
				return nil, fmt.Errorf("%s: step %q has a bad range", path, l)
			}
		}
		c.on = state == "on"
	}
	return steps, nil
}

// lit runs the steps (clipped to clip, if given) and counts lit cubes,
// using list as the signed cuboid list.
func lit(steps []cuboid, clip *cuboid, list []cuboid) (int, error) {
	if cap(list) == 0 {
		return 0, errors.New("no room for cuboids")
	}
	list = list[:0]
	for _, s := range steps {
		if clip != nil {
			var ok bool
			on := s.on
			if s, ok = overlap(s, *clip); !ok {
				continue
			}
			s.on = on
		}
		n := len(list)
		for i := range n {
			o, ok := overlap(s, list[i])
			if !ok {
				continue
			}
			if len(list) == cap(list) {
				return 0, fmt.Errorf("more than %d cuboids", cap(list))
			}
			o.sign = -list[i].sign
			list = append(list, o)
		}
		if s.on {
			if len(list) == cap(list) {
				return 0, fmt.Errorf("more than %d cuboids", cap(list))
			}
			s.sign = 1
			list = append(list, s)
		}
	}
	total := 0
	for _, c := range list {
		total += c.sign * (c.hi[0] - c.lo[0] + 1) * (c.hi[1] - c.lo[1] + 1) * (c.hi[2] - c.lo[2] + 1)
	}
	if total < 0 {
		return 0, errors.New("a negative number of cubes is lit")
	}
	return total, nil
}

func day22(path string) error {
	steps, err := readSteps(path)
	if err != nil {
		return err
	}
	list := make([]cuboid, 0, maxCuboids)
	region := cuboid{lo: [3]int{-50, -50, -50}, hi: [3]int{50, 50, 50}}
	small, err := lit(steps, &region, list)
	if err != nil {
		return err
	}
	all, err := lit(steps, nil, list)
	if err != nil {
		return err
	}

	// The whole reactor includes the small region.
	if all < small {
		return fmt.Errorf("%d cubes lit in all, fewer than the %d near the centre", all, small)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", small, all); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
