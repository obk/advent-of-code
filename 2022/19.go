// Day 19: Not Enough Minerals
//
// Each line of the input is a blueprint: the costs of ore, clay, obsidian
// and geode robots. Starting with one ore robot, each minute every robot
// collects one of its resource, and the factory can build one robot (ready
// the next minute).
//
// Part 1 sums blueprint number times the most geodes it can open in 24
// minutes.
// Part 2 multiplies the most geodes of the first three blueprints in 32
// minutes.
//
// A depth-first search with an explicit stack chooses the next robot to
// build and jumps ahead to the minute it can be afforded. A geode robot's
// whole output is counted when it is built. Branches are cut when there
// are already enough robots of a kind to afford any robot every minute, or
// when even a new geode robot every minute could not beat the best so far.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 14
	maxBlueprints = 64
	maxMinutes    = 32
	maxSearch     = 1 << 26
)

// blueprint is robot costs: cost[robot][resource] for ore, clay, obsidian.
type blueprint struct {
	id   int
	cost [4][3]int
	most [3]int // the most of each resource any robot costs
}

// frame is one point on the search path.
type frame struct {
	left   int
	robots [3]int
	have   [3]int
	geodes int
	next   int
}

// readBlueprints parses the blueprints.
func readBlueprints(path string) ([]blueprint, error) {
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
	if len(lines) < 3 || len(lines) > maxBlueprints {
		return nil, fmt.Errorf("%s: %d blueprints, expected 3..%d", path, len(lines), maxBlueprints)
	}
	bps := make([]blueprint, len(lines))
	for i, l := range lines {
		b := &bps[i]
		n, err := fmt.Sscanf(l, "Blueprint %d: Each ore robot costs %d ore. Each clay robot costs %d ore. Each obsidian robot costs %d ore and %d clay. Each geode robot costs %d ore and %d obsidian.",
			&b.id, &b.cost[0][0], &b.cost[1][0], &b.cost[2][0], &b.cost[2][1], &b.cost[3][0], &b.cost[3][2])
		if err != nil || n != 7 {
			return nil, fmt.Errorf("%s: malformed blueprint %q", path, l)
		}
		for r := range 4 {
			for k := range 3 {
				b.most[k] = max(b.most[k], b.cost[r][k])
			}
			if b.cost[r][0] < 1 {
				return nil, fmt.Errorf("%s: robot %d costs no ore in %q", path, r, l)
			}
		}
	}
	return bps, nil
}

// build returns the frame after waiting for and building robot kind, or
// false if it cannot be built in time.
func (b *blueprint) build(f *frame, kind int) (frame, bool, error) {
	if kind < 0 || kind > 3 || f == nil {
		return frame{}, false, fmt.Errorf("cannot build robot kind %d", kind)
	}
	wait := 0
	for k := range 3 {
		need := b.cost[kind][k] - f.have[k]
		if need <= 0 {
			continue
		}
		if f.robots[k] == 0 {
			return frame{}, false, nil
		}
		wait = max(wait, (need+f.robots[k]-1)/f.robots[k])
	}
	left := f.left - wait - 1
	if left <= 0 {
		return frame{}, false, nil
	}
	n := frame{left: left, robots: f.robots, geodes: f.geodes}
	for k := range 3 {
		n.have[k] = f.have[k] + f.robots[k]*(wait+1) - b.cost[kind][k]
	}
	if kind == 3 {
		n.geodes += left
	} else {
		n.robots[kind]++
	}
	return n, true, nil
}

// geodes returns the most geodes the blueprint opens in the time.
func (b *blueprint) geodes(minutes int) (int, error) {
	if minutes < 1 || minutes > maxMinutes {
		return 0, fmt.Errorf("%d minutes is out of range", minutes)
	}
	var stack [maxMinutes + 1]frame
	stack[0] = frame{left: minutes, robots: [3]int{1, 0, 0}}
	depth, best := 1, 0
	for steps := 0; depth > 0; steps++ {
		if steps > maxSearch {
			return 0, errors.New("the search ran too long")
		}
		f := &stack[depth-1]
		best = max(best, f.geodes)
		// Even a geode robot every remaining minute could not beat best.
		if f.next == 4 || f.geodes+f.left*(f.left-1)/2 <= best {
			depth--
			continue
		}
		kind := 3 - f.next // geode robots first
		f.next++
		if kind < 3 && f.robots[kind] >= b.most[kind] {
			continue
		}
		n, ok, err := b.build(f, kind)
		if err != nil {
			return 0, err
		}
		if !ok {
			continue
		}
		if depth == len(stack) {
			return 0, errors.New("search stack overflow")
		}
		stack[depth], depth = n, depth+1
	}
	return best, nil
}

func day19(path string) error {
	bps, err := readBlueprints(path)
	if err != nil {
		return err
	}
	quality, product := 0, 1
	for i := range bps {
		g, err := bps[i].geodes(24)
		if err != nil {
			return err
		}
		quality += bps[i].id * g
		if i < 3 {
			g, err := bps[i].geodes(32)
			if err != nil {
				return err
			}
			product *= g
		}
	}

	if quality < 1 {
		return errors.New("no blueprint opens a geode in 24 minutes")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", quality, product); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
