// Day 19: Beacon Scanner
//
// The input is reports from scanners, each listing the beacons it sees
// relative to itself, in its own unknown orientation (one of 24). Two
// scanners that share 12 beacons can be aligned.
//
// Part 1 counts the distinct beacons.
// Part 2 is the largest Manhattan distance between two scanners.
//
// Scanners are aligned breadth-first from scanner 0. For an aligned
// scanner A and another B, each rotation of B is tried: every pair of
// beacons (one from each) suggests an offset for B, and an offset
// suggested 12 times aligns it. Offset counts go in a map allocated once.
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
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 16
	maxScanners   = 48
	maxBeacons    = 32
	maxCoord      = 4_000
	needed        = 12
)

// vec is a point.
type vec [3]int

// rotation maps a vector: result[k] = sign[k] * v[perm[k]].
type rotation struct {
	perm, sign [3]int
}

// rotations returns the 24 proper rotations: axis permutations with
// signs whose determinant is +1.
func rotations() ([24]rotation, error) {
	var out [24]rotation
	n := 0
	perms := [6][3]int{{0, 1, 2}, {1, 2, 0}, {2, 0, 1}, {0, 2, 1}, {2, 1, 0}, {1, 0, 2}}
	for p, perm := range perms {
		parity := 1
		if p >= 3 {
			parity = -1
		}
		for s := range 8 {
			sign := [3]int{1 - 2*(s&1), 1 - 2*(s>>1&1), 1 - 2*(s>>2&1)}
			if parity*sign[0]*sign[1]*sign[2] == 1 {
				out[n], n = rotation{perm, sign}, n+1
			}
		}
	}
	if n != 24 {
		return out, fmt.Errorf("found %d rotations, not 24", n)
	}
	for i := range out {
		for j := range i {
			if out[i] == out[j] {
				return out, fmt.Errorf("rotations %d and %d are the same", j, i)
			}
		}
	}
	return out, nil
}

// apply rotates v.
func (r rotation) apply(v vec) vec {
	return vec{r.sign[0] * v[r.perm[0]], r.sign[1] * v[r.perm[1]], r.sign[2] * v[r.perm[2]]}
}

// key packs an offset into a map key.
func key(v vec) (int, error) {
	for _, c := range v {
		if c < -4*maxCoord || c >= 4*maxCoord {
			return 0, fmt.Errorf("offset %v is too large to pack", v)
		}
	}
	return ((v[0]+4*maxCoord)*8*maxCoord+v[1]+4*maxCoord)*8*maxCoord + v[2] + 4*maxCoord, nil
}

// survey is the reports and what is known so far.
type survey struct {
	beacons [maxScanners][maxBeacons]vec
	count   [maxScanners]int
	n       int
	aligned [maxScanners]bool
	pos     [maxScanners]vec
	abs     [maxScanners][maxBeacons]vec // beacons in scanner 0's frame
	offsets map[int]int
	rots    [24]rotation
}

// load parses the reports.
func load(path string) (*survey, error) {
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
	reports := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(reports) == 0 || len(reports) > maxScanners {
		return nil, fmt.Errorf("%s: more than %d scanners", path, maxScanners)
	}
	s := &survey{n: len(reports), offsets: make(map[int]int, maxBeacons*maxBeacons)}
	for i, r := range reports {
		lines := strings.Split(r, "\n")
		if len(lines)-1 > maxBeacons || len(lines) < needed+1 {
			return nil, fmt.Errorf("%s: scanner %d sees %d beacons", path, i, len(lines)-1)
		}
		for b, l := range lines[1:] {
			f := strings.Split(l, ",")
			if len(f) != 3 {
				return nil, fmt.Errorf("%s: malformed beacon %q", path, l)
			}
			for k := range 3 {
				v, err := strconv.Atoi(f[k])
				if err != nil || v < -maxCoord || v > maxCoord {
					return nil, fmt.Errorf("%s: invalid beacon %q", path, l)
				}
				s.beacons[i][b][k] = v
			}
			for _, earlier := range s.beacons[i][:b] {
				if earlier == s.beacons[i][b] {
					return nil, fmt.Errorf("%s: scanner %d lists %q twice", path, i, l)
				}
			}
		}
		s.count[i] = len(lines) - 1
	}
	s.rots, err = rotations()
	return s, err
}

// align tries to place scanner b using aligned scanner a.
func (s *survey) align(a, b int) (bool, error) {
	if !s.aligned[a] || s.aligned[b] {
		return false, fmt.Errorf("scanner %d must be aligned and %d not", a, b)
	}
	for _, r := range s.rots {
		clear(s.offsets)
		for _, pa := range s.abs[a][:s.count[a]] {
			for _, pb := range s.beacons[b][:s.count[b]] {
				rb := r.apply(pb)
				off := vec{pa[0] - rb[0], pa[1] - rb[1], pa[2] - rb[2]}
				k, err := key(off)
				if err != nil {
					return false, err
				}
				s.offsets[k]++
				if s.offsets[k] < needed {
					continue
				}
				s.aligned[b], s.pos[b] = true, off
				shared := 0
				for i, p := range s.beacons[b][:s.count[b]] {
					q := r.apply(p)
					s.abs[b][i] = vec{q[0] + off[0], q[1] + off[1], q[2] + off[2]}
					for _, pa := range s.abs[a][:s.count[a]] {
						if pa == s.abs[b][i] {
							shared++
						}
					}
				}
				if shared < needed {
					return false, fmt.Errorf("scanners %d and %d share only %d beacons", a, b, shared)
				}
				return true, nil
			}
		}
	}
	return false, nil
}

// alignAll aligns every scanner, breadth-first from scanner 0.
func (s *survey) alignAll() error {
	if s.n < 1 {
		return errors.New("no scanners")
	}
	var queue [maxScanners]int
	s.aligned[0] = true
	copy(s.abs[0][:], s.beacons[0][:s.count[0]])
	queue[0] = 0
	tail := 1
	for head := 0; head < tail; head++ {
		for b := range s.n {
			if s.aligned[b] {
				continue
			}
			ok, err := s.align(queue[head], b)
			if err != nil {
				return err
			}
			if ok {
				queue[tail], tail = b, tail+1
			}
		}
	}
	if tail != s.n {
		return fmt.Errorf("aligned %d of %d scanners", tail, s.n)
	}
	return nil
}

func day19(path string) error {
	s, err := load(path)
	if err != nil {
		return err
	}
	if err := s.alignAll(); err != nil {
		return err
	}
	clear(s.offsets)
	for i := range s.n {
		for _, p := range s.abs[i][:s.count[i]] {
			k, err := key(p)
			if err != nil {
				return err
			}
			s.offsets[k] = 1
		}
	}
	far := 0
	for i := range s.n {
		for j := range s.n {
			d := 0
			for k := range 3 {
				d += max(s.pos[i][k]-s.pos[j][k], s.pos[j][k]-s.pos[i][k])
			}
			far = max(far, d)
		}
	}

	if len(s.offsets) < s.count[0] {
		return fmt.Errorf("%d beacons, fewer than scanner 0 alone sees", len(s.offsets))
	}
	if s.n > 1 && far < 1 {
		return errors.New("every scanner is in the same place")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", len(s.offsets), far); err != nil {
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
