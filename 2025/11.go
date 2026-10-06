// Day 11: Reactor
//
// Each line of the input names a component and the components it feeds into:
//
//	svr: dac fft out
//
// The wiring is a directed acyclic graph, and both parts count routes through
// it rather than finding one.
//
// Part 1 counts the routes from "you" to "out".
// Part 2 counts the routes from "svr" to "out" that pass through both "dac"
// and "fft".
//
// The components are put in topological order once, so that every component
// comes before everything it feeds. Walking that order backwards, the number
// of routes from a component to the destination is just the sum over the
// components it feeds, all of which are already known. Part 2 never
// enumerates the routes either. A route cannot visit "dac" and "fft" in both
// orders, so there are exactly two families of route, each splitting into
// three independent legs whose counts multiply - six path counts and a little
// arithmetic instead of a search.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxNodes      = 20_000
)

// graph is the reactor's wiring, with components numbered 0..n-1. The
// components fed by node i are edges[edgeStart[i]:edgeStart[i+1]].
type graph struct {
	ids       map[string]int
	edgeStart []int
	edges     []int
	order     []int   // topological order: every node before those it feeds
	paths     []int64 // scratch space for countPaths, one entry per node
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

// numberNodes gives every component named in the input a number, and counts
// how many components each one feeds.
func numberNodes(lines []string) (ids map[string]int, degree []int, err error) {
	ids = make(map[string]int)
	degree = make([]int, 0, maxNodes)
	for _, line := range lines {
		source, destinations, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		names := strings.Fields(destinations)
		for _, name := range append(names, source) {
			if _, seen := ids[name]; seen {
				continue
			}
			if len(degree) == maxNodes {
				return nil, nil, fmt.Errorf("more than %d components", maxNodes)
			}
			ids[name] = len(degree)
			degree = append(degree, 0)
		}
		if degree[ids[source]] != 0 {
			return nil, nil, fmt.Errorf("component %q is listed twice", source)
		}
		degree[ids[source]] = len(names)
	}

	if len(degree) == 0 {
		return nil, nil, errors.New("no wiring found")
	}
	return ids, degree, nil
}

// parseInput reads the wiring into a graph, its edges packed into one slice.
func parseInput(path string) (graph, error) {
	lines, err := readLines(path)
	if err != nil {
		return graph{}, err
	}
	ids, degree, err := numberNodes(lines)
	if err != nil {
		return graph{}, err
	}

	g := graph{ids: ids, edgeStart: make([]int, len(degree)+1)}
	for i, d := range degree {
		g.edgeStart[i+1] = g.edgeStart[i] + d
	}
	g.edges = make([]int, g.edgeStart[len(degree)])
	for _, line := range lines {
		source, destinations, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		next := g.edgeStart[ids[source]]
		for _, name := range strings.Fields(destinations) {
			g.edges[next] = ids[name]
			next++
		}
		if next != g.edgeStart[ids[source]+1] {
			return graph{}, fmt.Errorf("component %q has a miscounted wiring list", source)
		}
	}
	return g, nil
}

// sortTopologically fills g.order using Kahn's algorithm: repeatedly take a
// component nothing unprocessed feeds into. Any component left over sits on
// a cycle.
func (g *graph) sortTopologically() error {
	n := len(g.edgeStart) - 1
	if n < 1 || len(g.edges) != g.edgeStart[n] {
		return errors.New("graph is empty or malformed")
	}

	inDegree := make([]int, n)
	for _, to := range g.edges {
		inDegree[to]++
	}
	g.order = make([]int, 0, n)
	for node, d := range inDegree {
		if d == 0 {
			g.order = append(g.order, node)
		}
	}
	// g.order doubles as the work queue: everything before i is processed.
	// Each node joins it once, so it never grows past n.
	for i := 0; i < len(g.order) && i < n; i++ {
		for _, to := range g.edges[g.edgeStart[g.order[i]]:g.edgeStart[g.order[i]+1]] {
			inDegree[to]--
			if inDegree[to] == 0 {
				g.order = append(g.order, to)
			}
		}
	}

	if len(g.order) != n {
		return fmt.Errorf("wiring has a cycle through %d components", n-len(g.order))
	}
	g.paths = make([]int64, n)
	return nil
}

// countPaths returns how many distinct routes lead from start to end.
//
// Walking the topological order backwards, every component a node feeds has
// already been counted, so its route count is just their sum.
func (g *graph) countPaths(start, end string) (int64, error) {
	from, ok := g.ids[start]
	if !ok {
		return 0, fmt.Errorf("unknown component %q", start)
	}
	to, ok := g.ids[end]
	if !ok {
		return 0, fmt.Errorf("unknown component %q", end)
	}
	if len(g.order) != len(g.paths) {
		return 0, errors.New("graph has not been sorted")
	}

	for i := len(g.order) - 1; i >= 0; i-- {
		node := g.order[i]
		if node == to {
			g.paths[node] = 1
			continue
		}
		var paths int64
		for _, next := range g.edges[g.edgeStart[node]:g.edgeStart[node+1]] {
			if paths > math.MaxInt64-g.paths[next] {
				return 0, fmt.Errorf("route count from %q overflows int64", start)
			}
			paths += g.paths[next]
		}
		g.paths[node] = paths
	}
	return g.paths[from], nil
}

// countRoutesVia counts the routes from start to end that visit first and then
// second, as the product of the three legs' route counts.
func (g *graph) countRoutesVia(start, first, second, end string) (int64, error) {
	product := int64(1)
	for _, leg := range [3][2]string{{start, first}, {first, second}, {second, end}} {
		paths, err := g.countPaths(leg[0], leg[1])
		if err != nil {
			return 0, err
		}
		if paths < 0 {
			return 0, fmt.Errorf("negative route count %s -> %s", leg[0], leg[1])
		}
		if paths != 0 && product > math.MaxInt64/paths {
			return 0, fmt.Errorf("route count via %q and %q overflows int64", first, second)
		}
		product *= paths
	}
	return product, nil
}

func day11(path string) error {
	g, err := parseInput(path)
	if err != nil {
		return err
	}
	if err := g.sortTopologically(); err != nil {
		return err
	}

	part1, err := g.countPaths("you", "out")
	if err != nil {
		return err
	}

	// Part 2 wants the routes from "svr" to "out" that pass through both "dac"
	// and "fft". The graph is acyclic, so a route visits them in one order or
	// the other, never both ways round.
	dacThenFft, err := g.countRoutesVia("svr", "dac", "fft", "out")
	if err != nil {
		return err
	}
	fftThenDac, err := g.countRoutesVia("svr", "fft", "dac", "out")
	if err != nil {
		return err
	}
	if dacThenFft != 0 && fftThenDac != 0 {
		return errors.New("routes visit dac and fft in both orders, so the wiring has a cycle")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", part1, dacThenFft+fftThenDac); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day11(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
