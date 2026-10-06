// Day 22: Grid Computing
//
// The input is the output of df for a grid of storage nodes, one line per
// node: "/dev/grid/node-x0-y0     92T   72T    20T   78%" (size, used,
// available, use%).
//
// Part 1 counts the viable pairs: ordered pairs of different nodes A, B where
// A is not empty and A's data would fit in B's free space.
// Part 2 is the fewest moves (each moving all of one node's data to an
// adjacent node) that bring the data from the top-right node to node 0,0.
//
// With about a thousand nodes, every pair is simply tried for part 1.
//
// Part 2 rests on the grid's shape, which the puzzle points out and the code
// checks: there is exactly one empty node, every node's data either fits in
// it or is too big to fit anywhere (an immovable "wall"). So the only thing
// that ever moves is the hole, swapping places with a neighbour, and the
// goal data moves when the hole swaps with it. That makes a state just
// (hole, goal), searched breadth first.
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
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 20
	maxLines      = 5_000
	maxTerabytes  = 10_000
)

// storage is one node's position and disk usage.
type storage struct {
	x, y              int
	size, used, avail int
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

// parseNode reads one df line.
func parseNode(line string) (storage, error) {
	f := strings.Fields(line)
	if len(f) != 5 {
		return storage{}, fmt.Errorf("malformed node %q", line)
	}
	name, found := strings.CutPrefix(f[0], "/dev/grid/node-x")
	if !found {
		return storage{}, fmt.Errorf("malformed node %q", line)
	}
	xText, yText, found := strings.Cut(name, "-y")
	if !found {
		return storage{}, fmt.Errorf("malformed node name in %q", line)
	}

	var values [5]int
	texts := [5]string{xText, yText, f[1], f[2], f[3]}
	for i, t := range texts {
		v, err := strconv.Atoi(strings.TrimSuffix(t, "T"))
		if err != nil {
			return storage{}, fmt.Errorf("invalid number in %q: %w", line, err)
		}
		if v < 0 || v > maxTerabytes {
			return storage{}, fmt.Errorf("number in %q is outside 0..%d", line, maxTerabytes)
		}
		values[i] = v
	}
	n := storage{x: values[0], y: values[1], size: values[2], used: values[3], avail: values[4]}
	if n.used+n.avail != n.size {
		return storage{}, fmt.Errorf("node %q: used and available do not add up to its size", line)
	}
	return n, nil
}

func parseInput(path string) ([]storage, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	nodes := make([]storage, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "/dev/grid/") {
			continue // the command and header lines
		}
		n, err := parseNode(line)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if len(nodes) < 2 {
		return nil, fmt.Errorf("%s: %d nodes, need at least two", path, len(nodes))
	}
	return nodes, nil
}

// viablePairs counts the ordered pairs whose data would fit.
func viablePairs(nodes []storage) (int, error) {
	if len(nodes) > maxLines {
		return 0, fmt.Errorf("%d nodes, more than %d", len(nodes), maxLines)
	}
	pairs := 0
	for i, a := range nodes {
		for j, b := range nodes {
			if i != j && a.used > 0 && a.used <= b.avail {
				pairs++
			}
		}
	}
	if pairs > len(nodes)*(len(nodes)-1) {
		return 0, fmt.Errorf("%d pairs among %d nodes", pairs, len(nodes))
	}
	return pairs, nil
}

// grid is the node layout for part 2's search.
type grid struct {
	width, height int
	wall          []bool // indexed y*width + x
	seen          []bool // indexed (hole cell)*cells + goal cell
	queue         []int
	depth         []int
}

// build lays the nodes out, finds the hole and checks the shape part 2 needs.
func build(nodes []storage) (g grid, hole int, err error) {
	for _, n := range nodes {
		g.width, g.height = max(g.width, n.x+1), max(g.height, n.y+1)
	}
	cells := g.width * g.height
	if cells != len(nodes) || cells > maxLines {
		return grid{}, 0, fmt.Errorf("%d nodes do not fill a %dx%d grid", len(nodes), g.width, g.height)
	}
	hole = -1
	for _, n := range nodes {
		if n.used == 0 {
			if hole >= 0 {
				return grid{}, 0, errors.New("more than one empty node")
			}
			hole = n.y*g.width + n.x
		}
	}
	if hole < 0 {
		return grid{}, 0, errors.New("no empty node")
	}
	space := nodes[0].size
	for _, n := range nodes {
		if n.used == 0 {
			space = n.size
		}
	}
	// Data can only ever move into the hole if no other node has room for
	// even the smallest amount of data anywhere.
	smallest, roomiest := maxTerabytes, 0
	for _, n := range nodes {
		if n.used > 0 {
			smallest = min(smallest, n.used)
			roomiest = max(roomiest, n.avail)
		}
	}
	if roomiest >= smallest {
		return grid{}, 0, fmt.Errorf("a full node has %dT free, so data could move outside the hole", roomiest)
	}
	g.wall = make([]bool, cells)
	for _, n := range nodes {
		g.wall[n.y*g.width+n.x] = n.used > space
	}
	g.seen = make([]bool, cells*cells)
	g.queue = make([]int, cells*cells)
	g.depth = make([]int, cells*cells)
	return g, hole, nil
}

// fewestMoves searches (hole, goal) states for the goal reaching cell 0.
func (g *grid) fewestMoves(hole, goal int) (int, error) {
	cells := g.width * g.height
	if hole < 0 || hole >= cells || goal < 0 || goal >= cells {
		return 0, fmt.Errorf("hole %d or goal %d is off the grid", hole, goal)
	}
	if g.wall[goal] {
		return 0, errors.New("the goal data can never move")
	}

	g.seen[hole*cells+goal] = true
	g.queue[0], g.depth[0] = hole*cells+goal, 0
	tail := 1
	for head := 0; head < tail; head++ {
		h, gl := g.queue[head]/cells, g.queue[head]%cells
		if gl == 0 {
			return g.depth[head], nil
		}
		x, y := h%g.width, h/g.width
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= g.width || ny >= g.height || g.wall[ny*g.width+nx] {
				continue
			}
			next := ny*g.width + nx
			nextGoal := gl
			if next == gl {
				nextGoal = h // the hole swaps with the goal data
			}
			if g.seen[next*cells+nextGoal] {
				continue
			}
			g.seen[next*cells+nextGoal] = true
			g.queue[tail], g.depth[tail] = next*cells+nextGoal, g.depth[head]+1
			tail++
		}
	}
	return 0, errors.New("the goal data cannot reach node 0,0")
}

func day22(path string) error {
	nodes, err := parseInput(path)
	if err != nil {
		return err
	}
	pairs, err := viablePairs(nodes)
	if err != nil {
		return err
	}
	g, hole, err := build(nodes)
	if err != nil {
		return err
	}
	moves, err := g.fewestMoves(hole, g.width-1)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", pairs, moves); err != nil {
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
