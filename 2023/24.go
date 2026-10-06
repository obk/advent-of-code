// Day 24: Never Tell Me The Odds
//
// Each line of the input is a hailstone: position and velocity, "19, 13,
// 30 @ -2, 1, -2".
//
// Part 1 counts pairs whose paths, ignoring z, cross in the future of both
// within 200000000000000..400000000000000 on x and y.
// Part 2 finds a position and velocity (all integers) from which a thrown
// rock hits every hailstone, and sums the position's coordinates.
//
// Part 2 tries rock velocities (vx, vy) in a range. Seen from the rock,
// every hailstone moves at its velocity minus the rock's and must pass
// through the rock's start. Two hailstones' relative paths fix that point
// (exactly, in integers); two more confirm it; then z follows from the
// times of the first two hits, and every hailstone is checked.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 16
	maxStones     = 400
	lo            = 200_000_000_000_000
	hi            = 400_000_000_000_000
	maxVelocity   = 1_000
)

// stone is a hailstone's position and velocity.
type stone struct {
	p, v [3]int
}

// crossings counts pairs whose x/y paths cross ahead, inside the area.
func crossings(s []stone, from, to float64) (int, error) {
	if len(s) < 2 || from > to {
		return 0, errors.New("too few hailstones or an empty area")
	}
	n := 0
	for i := range s {
		for j := i + 1; j < len(s); j++ {
			a, b := s[i], s[j]
			d := a.v[0]*b.v[1] - a.v[1]*b.v[0]
			if d == 0 {
				continue // parallel
			}
			dx, dy := b.p[0]-a.p[0], b.p[1]-a.p[1]
			if max(dx, -dx, dy, -dy) > 1<<50 {
				return 0, errors.New("hailstones are too far apart for exact products")
			}
			t := float64(dx*b.v[1]-dy*b.v[0]) / float64(d)
			u := float64(dx*a.v[1]-dy*a.v[0]) / float64(d)
			x, y := float64(a.p[0])+t*float64(a.v[0]), float64(a.p[1])+t*float64(a.v[1])
			if t >= 0 && u >= 0 && x >= from && x <= to && y >= from && y <= to {
				n++
			}
		}
	}
	return n, nil
}

// hits reports whether a rock thrown from p at v hits s.
func hits(s stone, p, v [3]int) bool {
	t := -1
	for k := range 3 {
		rel, gap := s.v[k]-v[k], p[k]-s.p[k]
		if rel == 0 && gap != 0 {
			return false // never closes the gap on this axis
		}
		if rel == 0 {
			continue
		}
		if gap%rel != 0 || gap/rel < 0 || (t >= 0 && gap/rel != t) {
			return false
		}
		t = gap / rel
	}
	return true
}

// tryVelocity tests a rock velocity (vx, vy) and returns the rock's start
// if it works.
func tryVelocity(s []stone, vx, vy int) ([3]int, bool, error) {
	var p [3]int
	if len(s) < 2 {
		return p, false, errors.New("too few hailstones")
	}
	a0, a1 := s[0].v[0]-vx, s[0].v[1]-vy
	b0, b1 := s[1].v[0]-vx, s[1].v[1]-vy
	den := a0*b1 - a1*b0
	if den == 0 {
		return p, false, nil
	}
	d0, d1 := s[1].p[0]-s[0].p[0], s[1].p[1]-s[0].p[1]
	n0, n1 := d0*b1-d1*b0, d0*a1-d1*a0
	if n0%den != 0 || n1%den != 0 {
		return p, false, nil
	}
	t0, t1 := n0/den, n1/den
	if t0 < 0 || t1 < 0 || t0 == t1 {
		return p, false, nil
	}
	p[0], p[1] = s[0].p[0]+t0*a0, s[0].p[1]+t0*a1
	// Where both hits happen on z fixes the rock's z velocity and start.
	z0, z1 := s[0].p[2]+t0*s[0].v[2], s[1].p[2]+t1*s[1].v[2]
	if (z0-z1)%(t0-t1) != 0 {
		return p, false, nil
	}
	vz := (z0 - z1) / (t0 - t1)
	p[2] = z0 - t0*vz
	v := [3]int{vx, vy, vz}
	for _, st := range s {
		if !hits(st, p, v) {
			return p, false, nil
		}
	}
	return p, true, nil
}

// throw searches rock velocities and returns the sum of the start.
func throw(s []stone) (int, error) {
	if len(s) < 4 {
		return 0, errors.New("too few hailstones to fix a throw")
	}
	if len(s) > maxStones {
		return 0, fmt.Errorf("more than %d hailstones", maxStones)
	}
	for vx := -maxVelocity; vx <= maxVelocity; vx++ {
		for vy := -maxVelocity; vy <= maxVelocity; vy++ {
			p, ok, err := tryVelocity(s, vx, vy)
			if err != nil {
				return 0, err
			}
			if ok {
				return p[0] + p[1] + p[2], nil
			}
		}
	}
	return 0, fmt.Errorf("no rock velocity within %d works", maxVelocity)
}

func day24(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxStones {
		return fmt.Errorf("%s: more than %d hailstones", path, maxStones)
	}
	stones := make([]stone, len(lines))
	for i, l := range lines {
		st := &stones[i]
		n, err := fmt.Sscanf(strings.ReplaceAll(l, " ", ""), "%d,%d,%d@%d,%d,%d", &st.p[0], &st.p[1], &st.p[2], &st.v[0], &st.v[1], &st.v[2])
		if err != nil || n != 6 {
			return fmt.Errorf("%s: malformed hailstone %q", path, l)
		}
		if st.v == [3]int{} {
			return fmt.Errorf("%s: hailstone %q does not move", path, l)
		}
	}
	from, to := float64(lo), float64(hi)
	if len(stones) < 10 { // the example uses a smaller area
		from, to = 7, 27
	}
	count, err := crossings(stones, from, to)
	if err != nil {
		return err
	}
	if count > len(stones)*(len(stones)-1)/2 {
		return fmt.Errorf("%d crossings among %d hailstones", count, len(stones))
	}
	sum, err := throw(stones)
	if err != nil {
		return err
	}

	if sum < 1 {
		return fmt.Errorf("the rock starts at coordinates summing to %d", sum)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, sum); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
