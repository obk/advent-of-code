// Day 16: Proboscidea Volcanium
//
// Each line of the input is a valve, its flow rate and its tunnels. From
// AA, moving through a tunnel or opening a valve takes a minute; an open
// valve releases its rate every minute after.
//
// Part 1 is the most pressure released in 30 minutes.
// Part 2 is the most you and an elephant release in 26 minutes, working
// at once.
//
// Only valves with a positive rate are worth visiting, and travel between
// them takes their shortest distance (Floyd-Warshall). A depth-first
// search with an explicit stack visits every order of opening them and
// records the best pressure for each set opened. For part 2 the two of
// you open disjoint sets: after a pass that makes each entry the best of
// its subsets, the answer is the best sum over a set and its complement.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 12
	maxValves     = 64
	maxUseful     = 16
	far           = 1 << 20
)

// cave is the valves and the distances between them.
type cave struct {
	rate   [maxValves]int
	dist   [maxValves][maxValves]int
	n      int
	useful [maxUseful]int
	k      int
	start  int
}

// load parses the scan and works out every distance.
func load(path string) (*cave, error) {
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
	if len(lines) > maxValves {
		return nil, fmt.Errorf("%s: more than %d valves", path, maxValves)
	}
	c := &cave{n: len(lines), start: -1}
	ids := make(map[string]int, len(lines))
	for i, l := range lines {
		f := strings.Fields(strings.NewReplacer("=", " ", ";", " ", ",", " ").Replace(l))
		if len(f) < 11 || f[0] != "Valve" {
			return nil, fmt.Errorf("%s: malformed valve %q", path, l)
		}
		ids[f[1]] = i
	}
	for i, l := range lines {
		f := strings.Fields(strings.NewReplacer("=", " ", ";", " ", ",", " ").Replace(l))
		if _, err := fmt.Sscanf(f[5], "%d", &c.rate[i]); err != nil || c.rate[i] < 0 {
			return nil, fmt.Errorf("%s: invalid rate in %q", path, l)
		}
		for j := range c.n {
			c.dist[i][j] = far
		}
		c.dist[i][i] = 0
		for _, t := range f[10:] {
			j, ok := ids[t]
			if !ok {
				return nil, fmt.Errorf("%s: tunnel to unknown valve %q", path, t)
			}
			c.dist[i][j] = 1
		}
		if f[1] == "AA" {
			c.start = i
		}
		if c.rate[i] > 0 && c.k == maxUseful {
			return nil, fmt.Errorf("%s: more than %d useful valves", path, maxUseful)
		}
		if c.rate[i] > 0 {
			c.useful[c.k], c.k = i, c.k+1
		}
	}
	if c.start < 0 {
		return nil, fmt.Errorf("%s: no valve AA", path)
	}
	return c, c.shortest()
}

// shortest turns tunnel lengths into shortest distances (Floyd-Warshall),
// and checks that every useful valve can be reached.
func (c *cave) shortest() error {
	for m := range c.n {
		for i := range c.n {
			for j := range c.n {
				c.dist[i][j] = min(c.dist[i][j], c.dist[i][m]+c.dist[m][j])
			}
		}
	}
	for _, v := range c.useful[:c.k] {
		if c.dist[c.start][v] >= far {
			return fmt.Errorf("valve %d cannot be reached from AA", v)
		}
	}
	return nil
}

// frame is one valve on the search path.
type frame struct {
	at, left, opened, released, next int
}

// explore records in best the most pressure for each set of opened valves
// within the time.
func (c *cave) explore(minutes int, best []int) error {
	if len(best) != 1<<c.k || minutes < 1 {
		return errors.New("wrong table size or no time")
	}
	clear(best)
	var stack [maxUseful + 1]frame
	stack[0] = frame{at: c.start, left: minutes}
	depth := 1
	for steps := 0; depth > 0; steps++ {
		if steps > 1<<30 {
			return errors.New("the search ran too long")
		}
		f := &stack[depth-1]
		best[f.opened] = max(best[f.opened], f.released)
		if f.next == c.k {
			depth--
			continue
		}
		j := f.next
		f.next++
		v := c.useful[j]
		left := f.left - c.dist[f.at][v] - 1
		if f.opened>>j&1 == 1 || left <= 0 {
			continue
		}
		stack[depth] = frame{at: v, left: left, opened: f.opened | 1<<j, released: f.released + left*c.rate[v]}
		depth++
	}
	return nil
}

func day16(path string) error {
	c, err := load(path)
	if err != nil {
		return err
	}
	best := make([]int, 1<<c.k)
	if err := c.explore(30, best); err != nil {
		return err
	}
	alone := 0
	for _, b := range best {
		alone = max(alone, b)
	}
	if err := c.explore(26, best); err != nil {
		return err
	}
	// Each set's entry becomes the best of any of its subsets.
	for bit := range c.k {
		for m := range best {
			if m>>bit&1 == 1 {
				best[m] = max(best[m], best[m^1<<bit])
			}
		}
	}
	together, all := 0, len(best)-1
	for m := range best {
		together = max(together, best[m]+best[all^m])
	}
	if together < alone/2 {
		return fmt.Errorf("two workers released %d, under half of one's %d", together, alone)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", alone, together); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
