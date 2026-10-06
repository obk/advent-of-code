// Day 8: Haunted Wasteland
//
// The input is a list of L/R turns, then nodes "AAA = (BBB, CCC)".
//
// Part 1 counts the steps from AAA to ZZZ, repeating the turns as needed.
// Part 2 starts at every node ending in A at once and counts the steps
// until all are on nodes ending in Z.
//
// In these inputs each start reaches a Z node after a number of steps that
// then repeats exactly, so part 2 is the least common multiple of each
// start's first Z time.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 16
	nodes         = 26 * 26 * 26
	maxSteps      = 10_000_000
)

// network is the turns and each node's left and right neighbours.
type network struct {
	turns       string
	left, right [nodes]int
	used        [nodes]bool
}

// id turns a three-letter name into a number.
func id(name string) (int, error) {
	if len(name) != 3 || strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") != "" {
		return 0, fmt.Errorf("invalid node %q", name)
	}
	v := 0
	for k := range 3 {
		c := int(name[k] - 'A')
		if name[k] < 'A' {
			c = int(name[k]-'0') % 26 // digits only appear in the examples
		}
		v = v*26 + c
	}
	return v, nil
}

// walk counts steps from start until a node ending in Z (or exactly ZZZ).
func (n *network) walk(start int, exact bool) (int, error) {
	if start < 0 || start >= nodes || !n.used[start] || len(n.turns) == 0 {
		return 0, fmt.Errorf("node %d is not in the network", start)
	}
	at := start
	for step := 0; step < maxSteps; step++ {
		if (exact && at == nodes-1) || (!exact && at%26 == 25) {
			return step, nil
		}
		if n.turns[step%len(n.turns)] == 'L' {
			at = n.left[at]
		} else {
			at = n.right[at]
		}
	}
	return 0, fmt.Errorf("no end within %d steps", maxSteps)
}

// lcm is the least common multiple of two positive numbers.
func lcm(a, b int) (int, error) {
	if a < 1 || b < 1 {
		return 0, fmt.Errorf("lcm of %d and %d", a, b)
	}
	x, y := a, b
	for k := 0; k < 128 && y != 0; k++ {
		x, y = y, x%y
	}
	if a/x > (1<<62)/b {
		return 0, errors.New("the lcm overflows")
	}
	return a / x * b, nil
}

func day8(path string) error {
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
	turns, rest, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found || turns == "" || strings.Trim(turns, "LR") != "" {
		return fmt.Errorf("%s: malformed turns", path)
	}
	n := &network{turns: turns}
	for _, l := range strings.Split(rest, "\n") {
		f := strings.Fields(strings.NewReplacer("=", " ", "(", " ", ",", " ", ")", " ").Replace(l))
		if len(f) != 3 {
			return fmt.Errorf("%s: malformed node %q", path, l)
		}
		var ids [3]int
		for k := range 3 {
			if ids[k], err = id(f[k]); err != nil {
				return err
			}
		}
		n.left[ids[0]], n.right[ids[0]], n.used[ids[0]] = ids[1], ids[2], true
	}
	if !n.used[0] || !n.used[nodes-1] {
		return fmt.Errorf("%s: no AAA or no ZZZ", path)
	}
	steps, err := n.walk(0, true)
	if err != nil {
		return err
	}
	all := 1
	for node := 0; node < nodes; node += 26 {
		if !n.used[node] {
			continue
		}
		s, err := n.walk(node, false)
		if err != nil {
			return err
		}
		if all, err = lcm(all, s); err != nil {
			return err
		}
	}

	if all < steps {
		return fmt.Errorf("the ghosts' %d steps are fewer than the %d from AAA alone", all, steps)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, all); err != nil {
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
