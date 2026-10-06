// Day 19: Medicine for Rudolph
//
// The input lists replacement rules such as "H => HO", then, after a blank
// line, the medicine molecule. A step replaces one occurrence of a rule's
// left side with its right side.
//
// Part 1 counts the distinct molecules one step can make from the medicine.
// Part 2 is the fewest steps to build the medicine from the single electron
// "e".
//
// Part 1 writes every one-step result into a single preallocated arena and
// counts those not equal to an earlier one; a hash of each result means only
// genuine candidates are compared byte for byte.
//
// Part 2 relies on the shape of the rules rather than a search. Molecules are
// sequences of elements (an uppercase letter plus an optional lowercase one).
// Every rule turns one element into either two elements, or into a form
// X Rn ... Ar with one, two or three elements inside separated by Y. So each
// step adds one element, except that Rn and Ar come for free and each Y brings
// one extra element along with it for free. Counting backwards from the
// medicine, the steps are
//
//	elements - count(Rn) - count(Ar) - 2*count(Y) - 1
//
// which is checked against the rules: every rule must have that shape.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath      = "inputs/19.txt"
	maxInputBytes  = 1 << 20
	maxLines       = 1_000
	maxMoleculeLen = 2_000
)

// rule is one replacement.
type rule struct {
	from, to string
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

func parseInput(path string) ([]rule, string, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, "", err
	}

	rules := make([]rule, 0, len(lines))
	molecule := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		from, to, found := strings.Cut(line, " => ")
		if !found {
			if molecule != "" {
				return nil, "", fmt.Errorf("%s: more than one molecule", path)
			}
			molecule = line
			continue
		}
		if from == "" || to == "" {
			return nil, "", fmt.Errorf("malformed rule %q", line)
		}
		rules = append(rules, rule{from: from, to: to})
	}
	if len(rules) == 0 || molecule == "" || len(molecule) > maxMoleculeLen {
		return nil, "", fmt.Errorf("%s: need rules and a molecule of 1..%d letters", path, maxMoleculeLen)
	}
	return rules, molecule, nil
}

// countOneStep counts the distinct molecules one replacement makes. arena and
// starts/hashes are scratch space big enough for every possible result.
func countOneStep(rules []rule, molecule string, arena []byte, starts []int, hashes []uint64) (int, error) {
	if len(starts) != len(hashes) || len(starts) < 2 {
		return 0, errors.New("scratch space is malformed")
	}

	n := 0 // results stored so far; result k is arena[starts[k]:starts[k+1]]
	distinct := 0
	for _, r := range rules {
		for i := range len(molecule) - len(r.from) + 1 {
			if molecule[i:i+len(r.from)] != r.from {
				continue
			}
			if n+1 >= len(starts) || len(arena)+len(molecule)+len(r.to) > cap(arena) {
				return 0, fmt.Errorf("more than %d one-step results", len(starts)-1)
			}
			arena = append(arena, molecule[:i]...)
			arena = append(arena, r.to...)
			arena = append(arena, molecule[i+len(r.from):]...)
			starts[n+1] = len(arena)
			result := arena[starts[n]:starts[n+1]]
			h := fnv.New64a()
			_, _ = h.Write(result) // writing to a hash never fails
			hashes[n] = h.Sum64()

			seen := false
			for k := range n {
				if hashes[k] == hashes[n] && bytes.Equal(arena[starts[k]:starts[k+1]], result) {
					seen = true
					break
				}
			}
			if !seen {
				distinct++
			}
			n++
		}
	}
	if distinct > n {
		return 0, fmt.Errorf("%d distinct molecules out of %d results", distinct, n)
	}
	return distinct, nil
}

// elementCounts splits a molecule into elements and counts them, along with
// the Rn, Ar and Y elements among them.
func elementCounts(molecule string) (elements, rn, ar, y int, err error) {
	for i := range len(molecule) {
		c := molecule[i]
		if c >= 'a' && c <= 'z' {
			if i == 0 || molecule[i-1] < 'A' || molecule[i-1] > 'Z' {
				return 0, 0, 0, 0, fmt.Errorf("%q has a lowercase letter not after an uppercase one", molecule)
			}
			continue
		}
		if c < 'A' || c > 'Z' {
			return 0, 0, 0, 0, fmt.Errorf("%q holds %q", molecule, c)
		}
		elements++
		switch {
		case strings.HasPrefix(molecule[i:], "Rn"):
			rn++
		case strings.HasPrefix(molecule[i:], "Ar"):
			ar++
		case c == 'Y':
			y++
		}
	}
	return elements, rn, ar, y, nil
}

// stepsFromElectron applies the counting formula, after checking that every
// rule has the shape the formula depends on.
func stepsFromElectron(rules []rule, molecule string) (int, error) {
	if len(rules) == 0 || molecule == "" {
		return 0, errors.New("need rules and a molecule")
	}
	for _, r := range rules {
		elements, rn, ar, y, err := elementCounts(r.to)
		if err != nil {
			return 0, err
		}
		// Each rule must add exactly one element net of its free ones.
		if r.from != "e" && elements-rn-ar-2*y != 2 {
			return 0, fmt.Errorf("rule %s => %s does not fit the counting argument", r.from, r.to)
		}
	}

	elements, rn, ar, y, err := elementCounts(molecule)
	if err != nil {
		return 0, err
	}
	steps := elements - rn - ar - 2*y - 1
	if steps < 1 {
		return 0, fmt.Errorf("medicine needs %d steps", steps)
	}
	return steps, nil
}

func day19(path string) error {
	rules, molecule, err := parseInput(path)
	if err != nil {
		return err
	}

	maxResults := len(rules) * len(molecule)
	longest := 0
	for _, r := range rules {
		longest = max(longest, len(r.to))
	}
	arena := make([]byte, 0, maxResults*(len(molecule)+longest))
	starts := make([]int, maxResults+1)
	hashes := make([]uint64, maxResults+1)

	oneStep, err := countOneStep(rules, molecule, arena, starts, hashes)
	if err != nil {
		return err
	}
	steps, err := stepsFromElectron(rules, molecule)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", oneStep, steps); err != nil {
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
