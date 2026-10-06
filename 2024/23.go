// Day 23: LAN Party
//
// Each line of the input links two computers (two-letter names), "kh-tc".
//
// Part 1 counts sets of three computers all linked to each other with at
// least one name starting with t.
// Part 2 finds the largest set of computers all linked to each other, and
// gives their names sorted and joined with commas.
//
// Names map to numbers 0..675. For part 2, every clique lies inside some
// computer's neighbourhood; each computer's neighbours (at most 16) are
// numbered, and every subset of them is tested as a bit set.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 16
	names         = 26 * 26
	maxDegree     = 16
)

// lan is the links.
type lan struct {
	link [names][names]bool
	nbr  [names][maxDegree]int
	deg  [names]int
}

// triangles counts linked triples with a name starting with t.
func (l *lan) triangles() (int, error) {
	n := 0
	for a := range names {
		if l.deg[a] > maxDegree {
			return 0, fmt.Errorf("computer %d has too many links", a)
		}
		for _, b := range l.nbr[a][:l.deg[a]] {
			for _, c := range l.nbr[b][:l.deg[b]] {
				if a < b && b < c && l.link[a][c] && (a/26 == 't'-'a' || b/26 == 't'-'a' || c/26 == 't'-'a') {
					n++
				}
			}
		}
	}
	return n, nil
}

// clique finds the largest clique; it returns its members as a bit set
// over v's neighbours, and v.
func (l *lan) clique() (int, uint32, error) {
	bestV, bestSet, bestSize := -1, uint32(0), 0
	for v := range names {
		d := l.deg[v]
		var adj [maxDegree]uint32 // neighbours of v linked to neighbour i
		for i := range d {
			for j := range d {
				if l.link[l.nbr[v][i]][l.nbr[v][j]] {
					adj[i] |= 1 << j
				}
			}
		}
		for set := uint32(1); set < 1<<d; set++ {
			size := bits.OnesCount32(set)
			if size+1 <= bestSize {
				continue
			}
			ok := true
			for i := range d {
				if set>>i&1 == 1 && (adj[i]|1<<i)&set != set {
					ok = false
				}
			}
			if ok {
				bestV, bestSet, bestSize = v, set, size+1
			}
		}
	}
	if bestV < 0 || bestSize < 2 {
		return 0, 0, errors.New("no links")
	}
	return bestV, bestSet, nil
}

// password lists v and the chosen neighbours, sorted by name, into out.
func (l *lan) password(v int, set uint32, out []byte) ([]byte, error) {
	if v < 0 || v >= names || set>>l.deg[v] != 0 {
		return nil, fmt.Errorf("computer %d or its set %b is out of range", v, set)
	}
	var members [maxDegree + 1]int
	n := 0
	members[0], n = v, 1
	for i := range l.deg[v] {
		if set>>i&1 == 1 {
			members[n], n = l.nbr[v][i], n+1
		}
	}
	for i := 1; i < n; i++ {
		for k := i; k > 0 && members[k-1] > members[k]; k-- {
			members[k-1], members[k] = members[k], members[k-1]
		}
	}
	if n != bits.OnesCount32(set)+1 {
		return nil, fmt.Errorf("%d members from a set of %d", n, bits.OnesCount32(set))
	}
	if cap(out) < 3*n {
		return nil, errors.New("no room for the password")
	}
	out = out[:0]
	for i, m := range members[:n] {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, byte('a'+m/26), byte('a'+m%26))
	}
	return out, nil
}

func day23(path string) error {
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
	l := new(lan)
	for _, line := range strings.Fields(string(content)) {
		if len(line) != 5 || line[2] != '-' || strings.Trim(line[:2]+line[3:], "abcdefghijklmnopqrstuvwxyz") != "" {
			return fmt.Errorf("%s: malformed link %q", path, line)
		}
		a, b := int(line[0]-'a')*26+int(line[1]-'a'), int(line[3]-'a')*26+int(line[4]-'a')
		if a == b || l.link[a][b] || l.deg[a] == maxDegree || l.deg[b] == maxDegree {
			return fmt.Errorf("%s: link %q repeats, loops, or has too many neighbours", path, line)
		}
		l.link[a][b], l.link[b][a] = true, true
		l.nbr[a][l.deg[a]], l.deg[a] = b, l.deg[a]+1
		l.nbr[b][l.deg[b]], l.deg[b] = a, l.deg[b]+1
	}
	if l.deg == [names]int{} {
		return fmt.Errorf("%s: no links", path)
	}
	v, set, err := l.clique()
	if err != nil {
		return err
	}
	word, err := l.password(v, set, make([]byte, 0, 3*(maxDegree+1)))
	if err != nil {
		return err
	}

	triples, err := l.triangles()
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", triples, word); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day23(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
