// Day 25: Snowverload
//
// Each line of the input is a component and the components it is wired
// to, "jqt: rhn xhk nvd". Cutting exactly three wires splits the
// components into two groups.
//
// Part 1 multiplies the two groups' sizes. (Day 25 has no second puzzle.)
//
// With every wire a unit of capacity, the two sides are exactly the pairs
// of components between which at most three units can flow. From one
// component, a maximum flow (augmenting paths found breadth-first) is tried
// to each other until one stops at three; the components still reachable
// from the first then form its group.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 16
	maxNodes      = 2_048
	maxWires      = 4_096
	cut           = 3
)

// graph is the wiring as arcs: arc 2k goes a->b and arc 2k+1 b->a.
type graph struct {
	ids        map[string]int
	n          int
	from, to   [2 * maxWires]int
	m          int   // arcs
	start, adj []int // arcs out of each node: adj[start[v]:start[v+1]]
	flow       [2 * maxWires]int
	prev       [maxNodes]int // the arc used to reach each node
	queue      [maxNodes]int
}

// id returns a component's number, numbering it if new.
func (g *graph) id(name string) (int, error) {
	if i, ok := g.ids[name]; ok {
		return i, nil
	}
	if g.n == maxNodes || name == "" {
		return 0, errors.New("too many components, or an empty name")
	}
	g.ids[name], g.n = g.n, g.n+1
	return g.n - 1, nil
}

// load parses the wiring and builds each node's list of arcs.
func (g *graph) load(text string) error {
	g.ids = make(map[string]int, maxNodes)
	for _, l := range strings.Split(strings.TrimSpace(text), "\n") {
		name, rest, ok := strings.Cut(l, ": ")
		if !ok {
			return fmt.Errorf("malformed line %q", l)
		}
		a, err := g.id(name)
		if err != nil {
			return err
		}
		for _, o := range strings.Fields(rest) {
			b, err := g.id(o)
			if err != nil || g.m+2 > len(g.from) {
				return fmt.Errorf("too many wires at %q", o)
			}
			if a == b {
				return fmt.Errorf("component %q is wired to itself", o)
			}
			g.from[g.m], g.to[g.m], g.from[g.m+1], g.to[g.m+1] = a, b, b, a
			g.m += 2
		}
	}
	if g.n < 2 || g.m == 0 {
		return errors.New("fewer than two components, or no wires")
	}
	g.start, g.adj = make([]int, g.n+1), make([]int, g.m)
	for k := range g.m {
		g.start[g.from[k]+1]++
	}
	for v := range g.n {
		g.start[v+1] += g.start[v]
	}
	var fill [maxNodes]int
	for k := range g.m {
		v := g.from[k]
		g.adj[g.start[v]+fill[v]], fill[v] = k, fill[v]+1
	}
	return nil
}

// reach searches the residual graph from s; it returns how many nodes it
// reached and whether t was one, leaving the path in prev.
func (g *graph) reach(s, t int) (int, bool, error) {
	if s < 0 || t < 0 || s >= g.n || t >= g.n {
		return 0, false, fmt.Errorf("nodes %d and %d are out of range", s, t)
	}
	for v := range g.n {
		g.prev[v] = -2
	}
	g.prev[s], g.queue[0] = -1, s
	tail := 1
	for head := 0; head < tail; head++ {
		v := g.queue[head]
		for _, k := range g.adj[g.start[v]:g.start[v+1]] {
			w := g.to[k]
			if g.prev[w] != -2 || g.flow[k] >= 1 {
				continue
			}
			g.prev[w], g.queue[tail], tail = k, w, tail+1
		}
	}
	if tail > g.n {
		return 0, false, errors.New("reached more nodes than exist")
	}
	return tail, g.prev[t] != -2, nil
}

// maxFlow pushes up to cut+1 units from s to t and returns how many went.
func (g *graph) maxFlow(s, t int) (int, error) {
	if s == t || s < 0 || t < 0 || s >= g.n || t >= g.n {
		return 0, fmt.Errorf("cannot flow from %d to %d", s, t)
	}
	clear(g.flow[:g.m])
	for units := 0; units <= cut; units++ {
		_, ok, err := g.reach(s, t)
		if err != nil {
			return 0, err
		}
		if !ok {
			return units, nil
		}
		for v, k := t, 0; v != s && k < g.n; k++ {
			arc := g.prev[v]
			if arc < 0 {
				return 0, errors.New("broken augmenting path")
			}
			g.flow[arc]++
			g.flow[arc^1]--
			v = g.from[arc]
		}
	}
	return cut + 1, nil
}

func day25(path string) error {
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
	g := new(graph)
	if err := g.load(string(content)); err != nil {
		return err
	}
	for t := 1; t < g.n; t++ {
		f, err := g.maxFlow(0, t)
		if err != nil {
			return err
		}
		if f != cut {
			continue
		}
		side, _, err := g.reach(0, t)
		if err != nil {
			return err
		}
		if side == 0 || side == g.n {
			return errors.New("the cut leaves one group empty")
		}

		if _, err := fmt.Printf("Part 1: %d\n", side*(g.n-side)); err != nil {
			return err
		}
		return nil
	}
	return errors.New("no three-wire cut")
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
