// Day 12: Passage Pathing
//
// Each line of the input joins two caves, "A-b". Big caves (upper case)
// can be visited any number of times, small ones (lower case) once.
//
// Part 1 counts paths from start to end.
// Part 2 also allows one small cave (not start or end) to be visited twice.
//
// Paths are counted by depth-first search with an explicit stack of frames
// (cave, next neighbour to try, small caves visited, whether the second
// visit is used).
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 10
	maxCaves      = 32
	maxDepth      = 64
	maxPaths      = 10_000_000
)

// caves is the cave graph.
type caves struct {
	ids        map[string]int
	adj        [maxCaves][maxCaves]bool
	small      [maxCaves]bool
	n          int
	start, end int
}

// id returns a cave's number, numbering it if new.
func (c *caves) id(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty cave name")
	}
	if i, ok := c.ids[name]; ok {
		return i, nil
	}
	if c.n == maxCaves {
		return 0, fmt.Errorf("more than %d caves", maxCaves)
	}
	c.ids[name], c.small[c.n] = c.n, name[0] >= 'a'
	c.n++
	return c.n - 1, nil
}

// load parses the passages.
func load(path string) (*caves, error) {
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
	c := &caves{ids: make(map[string]int, maxCaves)}
	for _, line := range strings.Fields(string(content)) {
		a, b, found := strings.Cut(line, "-")
		if !found {
			return nil, fmt.Errorf("%s: malformed passage %q", path, line)
		}
		x, err := c.id(a)
		if err != nil {
			return nil, err
		}
		y, err := c.id(b)
		if err != nil {
			return nil, err
		}
		if !c.small[x] && !c.small[y] {
			return nil, fmt.Errorf("%s: two big caves %q joined; paths would be endless", path, line)
		}
		c.adj[x][y], c.adj[y][x] = true, true
	}
	var okS, okE bool
	c.start, okS = c.ids["start"]
	c.end, okE = c.ids["end"]
	if !okS || !okE {
		return nil, fmt.Errorf("%s: no start or no end", path)
	}
	return c, nil
}

// frame is one cave on the search path.
type frame struct {
	cave, next int
	visited    uint32
	twice      bool
}

// paths counts paths from start to end; allowTwice permits one small cave
// twice.
func (c *caves) paths(allowTwice bool) (int, error) {
	var stack [maxDepth]frame
	stack[0] = frame{cave: c.start, visited: 1 << c.start}
	depth, count := 1, 0
	for steps := 0; depth > 0; steps++ {
		if steps > maxPaths*maxCaves {
			return 0, errors.New("the search ran too long")
		}
		f := &stack[depth-1]
		if f.cave == c.end || f.next == c.n {
			if f.cave == c.end {
				count++
			}
			depth--
			continue
		}
		to := f.next
		f.next++
		if !c.adj[f.cave][to] || to == c.start {
			continue
		}
		seen := c.small[to] && f.visited>>to&1 == 1
		if seen && (!allowTwice || f.twice || to == c.end) {
			continue
		}
		if depth == maxDepth {
			return 0, fmt.Errorf("paths deeper than %d", maxDepth)
		}
		stack[depth] = frame{cave: to, visited: f.visited | 1<<to, twice: f.twice || seen}
		depth++
	}
	if count < 1 {
		return 0, errors.New("no path from start to end")
	}
	return count, nil
}

func day12(path string) error {
	c, err := load(path)
	if err != nil {
		return err
	}
	once, err := c.paths(false)
	if err != nil {
		return err
	}
	twice, err := c.paths(true)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", once, twice); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
