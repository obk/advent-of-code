// Day 20: Particle Swarm
//
// Each line of the input is a particle: "p=<x,y,z>, v=<...>, a=<...>". Every
// tick, each particle's velocity grows by its acceleration and then its
// position by its velocity.
//
// Part 1 is the particle that stays closest to the origin (by taxicab
// distance) in the long run.
// Part 2 removes particles that collide (share a position after a tick) and
// counts those left once collisions stop.
//
// In the long run the particle with the smallest acceleration wins, then the
// one whose velocity points least away, and so on - which is awkward to
// compare exactly. Instead the swarm is simulated for many ticks; once every
// particle's velocity has turned to follow its acceleration, the order by
// distance no longer changes, which is checked by requiring the same
// particle to lead for the whole last stretch.
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
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxValue      = 100_000
	ticks         = 2_000
	settled       = 500 // the leader must hold for the last this many ticks
)

// vec is a 3D integer vector.
type vec struct {
	x, y, z int
}

// particle is one particle's state.
type particle struct {
	p, v, a vec
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

// parseParticle reads one line.
func parseParticle(line string) (particle, error) {
	clean := strings.NewReplacer("p=<", "", "v=<", "", "a=<", "", ">", "", " ", "").Replace(line)
	f := strings.Split(clean, ",")
	if len(f) != 9 {
		return particle{}, fmt.Errorf("malformed particle %q", line)
	}
	var n [9]int
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return particle{}, fmt.Errorf("invalid number in %q: %w", line, err)
		}
		if v < -maxValue || v > maxValue {
			return particle{}, fmt.Errorf("number in %q is outside ±%d", line, maxValue)
		}
		n[i] = v
	}
	return particle{vec{n[0], n[1], n[2]}, vec{n[3], n[4], n[5]}, vec{n[6], n[7], n[8]}}, nil
}

func parseInput(path string) ([]particle, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	swarm := make([]particle, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		p, err := parseParticle(line)
		if err != nil {
			return nil, err
		}
		swarm = append(swarm, p)
	}
	if len(swarm) == 0 {
		return nil, fmt.Errorf("%s: no particles", path)
	}
	if len(swarm) > maxLines {
		return nil, fmt.Errorf("%s: %d particles, more than %d", path, len(swarm), maxLines)
	}
	return swarm, nil
}

// distance is the taxicab distance of v from the origin.
func distance(v vec) int {
	return max(v.x, -v.x) + max(v.y, -v.y) + max(v.z, -v.z)
}

// step advances one particle by one tick. Over the ticks simulated the
// values stay far below the limits checked here.
func (p *particle) step() error {
	p.v.x, p.v.y, p.v.z = p.v.x+p.a.x, p.v.y+p.a.y, p.v.z+p.a.z
	if distance(p.v) > 1<<40 {
		return fmt.Errorf("velocity %+v has run away", p.v)
	}
	p.p.x, p.p.y, p.p.z = p.p.x+p.v.x, p.p.y+p.v.y, p.p.z+p.v.z
	if distance(p.p) > 1<<50 {
		return fmt.Errorf("position %+v has run away", p.p)
	}
	return nil
}

// closest simulates the swarm (a copy, in work) and returns the particle
// that leads for the whole final stretch.
func closest(swarm, work []particle) (int, error) {
	if len(work) != len(swarm) || len(swarm) == 0 {
		return 0, errors.New("work space does not match the swarm")
	}
	if len(swarm) > maxLines {
		return 0, fmt.Errorf("%d particles, more than %d", len(swarm), maxLines)
	}
	copy(work, swarm)
	leader := -1
	for t := range ticks {
		best := 0
		for i := range work {
			if err := work[i].step(); err != nil {
				return 0, err
			}
			if distance(work[i].p) < distance(work[best].p) {
				best = i
			}
		}
		if t >= ticks-settled && leader >= 0 && best != leader {
			return 0, fmt.Errorf("the lead changed from %d to %d at tick %d", leader, best, t)
		}
		leader = best
	}
	if leader < 0 {
		return 0, errors.New("no particle led")
	}
	return leader, nil
}

// survivors simulates collisions on a copy of the swarm (in work) and counts
// the particles left. alive is scratch space. Collisions must have stopped
// for the whole final stretch.
func survivors(swarm, work []particle, alive []bool) (int, error) {
	if len(work) != len(swarm) || len(alive) != len(swarm) || len(swarm) == 0 {
		return 0, errors.New("work space does not match the swarm")
	}
	if len(swarm) > maxLines {
		return 0, fmt.Errorf("%d particles, more than %d", len(swarm), maxLines)
	}
	copy(work, swarm)
	for i := range alive {
		alive[i] = true
	}
	lastCollision := 0
	for t := range ticks {
		for i := range work {
			if err := work[i].step(); err != nil {
				return 0, err
			}
		}
		for i := range work {
			hit := false
			for j := i + 1; j < len(work); j++ {
				if alive[i] && alive[j] && work[i].p == work[j].p {
					alive[j], hit = false, true
				}
			}
			if hit {
				alive[i], lastCollision = false, t
			}
		}
	}
	if lastCollision >= ticks-settled {
		return 0, fmt.Errorf("particles were still colliding at tick %d", lastCollision)
	}
	left := 0
	for _, a := range alive {
		if a {
			left++
		}
	}
	return left, nil
}

func day20(path string) error {
	swarm, err := parseInput(path)
	if err != nil {
		return err
	}
	leader, err := closest(swarm, make([]particle, len(swarm)))
	if err != nil {
		return err
	}

	left, err := survivors(swarm, make([]particle, len(swarm)), make([]bool, len(swarm)))
	if err != nil {
		return err
	}
	if leader < 0 || leader >= len(swarm) {
		return fmt.Errorf("leader %d is not a particle", leader)
	}
	if left > len(swarm) {
		return fmt.Errorf("%d survivors out of %d", left, len(swarm))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", leader, left); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
