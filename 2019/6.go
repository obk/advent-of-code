// Day 6: Universal Orbit Map
//
// Each line of the input is an orbit, "AAA)BBB": BBB orbits AAA. Everything
// orbits, directly or not, the centre of mass COM.
//
// Part 1 is the total number of direct and indirect orbits.
// Part 2 is the number of orbital transfers from the object YOU orbits to
// the object SAN orbits: up to their deepest common ancestor and down again.
//
// Objects are numbered as they are named, each storing what it orbits. An
// object's orbit count is one more than its parent's, so counting along a
// breadth-first order from COM gives every count without recursion.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 20
	maxObjects    = 10_000
)

// orbits holds the map: parent[i] is what object i orbits (-1 for COM).
type orbits struct {
	ids    map[string]int
	parent []int
	depth  []int
}

// id returns an object's number, numbering it if new.
func (o *orbits) id(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty object name")
	}
	if i, ok := o.ids[name]; ok {
		return i, nil
	}
	if len(o.parent) == cap(o.parent) {
		return 0, fmt.Errorf("more than %d objects", cap(o.parent))
	}
	o.ids[name] = len(o.parent)
	o.parent = append(o.parent, -1)
	return o.ids[name], nil
}

// load parses the map.
func load(path string) (*orbits, error) {
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
	o := &orbits{ids: make(map[string]int, maxObjects), parent: make([]int, 0, maxObjects)}
	for _, line := range strings.Fields(string(content)) {
		center, body, found := strings.Cut(line, ")")
		if !found {
			return nil, fmt.Errorf("%s: malformed orbit %q", path, line)
		}
		c, err := o.id(center)
		if err != nil {
			return nil, err
		}
		b, err := o.id(body)
		if err != nil {
			return nil, err
		}
		if o.parent[b] >= 0 {
			return nil, fmt.Errorf("%s: %s orbits two objects", path, body)
		}
		o.parent[b] = c
	}
	return o, nil
}

// depths fills every object's number of direct and indirect orbits. Each
// pass settles objects whose parent is settled; no chain is longer than the
// number of objects.
func (o *orbits) depths() error {
	n := len(o.parent)
	o.depth = make([]int, n)
	for i := range o.depth {
		o.depth[i] = -1
		if o.parent[i] < 0 {
			o.depth[i] = 0
		}
	}
	for range n {
		changed := false
		for i, p := range o.parent {
			if o.depth[i] < 0 && p >= 0 && o.depth[p] >= 0 {
				o.depth[i], changed = o.depth[p]+1, true
			}
		}
		if !changed {
			break
		}
	}
	roots := 0
	for i, d := range o.depth {
		if d < 0 {
			return fmt.Errorf("object %d never reaches the centre", i)
		}
		if d == 0 {
			roots++
		}
	}
	if roots != 1 {
		return fmt.Errorf("%d objects orbit nothing, expected just COM", roots)
	}
	return nil
}

// transfers counts the moves between the objects a and b orbit.
func (o *orbits) transfers(a, b string) (int, error) {
	if a == b {
		return 0, fmt.Errorf("cannot transfer from %s to itself", a)
	}
	x, okA := o.ids[a]
	y, okB := o.ids[b]
	if !okA || !okB || o.parent[x] < 0 || o.parent[y] < 0 {
		return 0, fmt.Errorf("%s or %s is missing or orbits nothing", a, b)
	}
	x, y = o.parent[x], o.parent[y]
	moves := 0
	for range 2 * len(o.parent) {
		if x == y {
			return moves, nil
		}
		// Step up from whichever is deeper (both, when level).
		if o.depth[x] >= o.depth[y] {
			x, moves = o.parent[x], moves+1
		} else {
			y, moves = o.parent[y], moves+1
		}
		if x < 0 || y < 0 {
			return 0, errors.New("the paths never meet")
		}
	}
	return 0, errors.New("the paths never meet")
}

func day6(path string) error {
	o, err := load(path)
	if err != nil {
		return err
	}
	if err := o.depths(); err != nil {
		return err
	}
	total := 0
	for _, d := range o.depth {
		total += d
	}

	if total < len(o.parent)-1 {
		return fmt.Errorf("%d orbits among %d objects", total, len(o.parent))
	}
	moves, err := o.transfers("YOU", "SAN")
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, moves); err != nil {
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
