// Day 14: Extended Polymerization
//
// The input is a polymer template, then rules "CH -> B": between each
// adjacent C and H, insert B (all insertions in a step happen at once).
//
// Part 1 takes the most common element's count minus the least common's
// after 10 steps.
// Part 2 does the same after 40.
//
// Only the count of each adjacent pair matters: pair CH becomes CB and BH.
// Each element is counted as the first letter of its pairs, plus the
// template's last letter, which never changes.
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
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 12
	letters       = 26
)

// polymer is the template and the insertion rules (0 for none).
type polymer struct {
	template string
	insert   [letters][letters]byte
}

// load parses the template and rules.
func load(path string) (*polymer, error) {
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
	template, rules, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found || len(template) < 2 || strings.Trim(template, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return nil, fmt.Errorf("%s: malformed template", path)
	}
	p := &polymer{template: template}
	lines := strings.Split(rules, "\n")
	if len(lines) > letters*letters {
		return nil, fmt.Errorf("%s: %d rules is more than there are pairs", path, len(lines))
	}
	for _, line := range lines {
		pair, ins, ok := strings.Cut(line, " -> ")
		if !ok || len(pair) != 2 || len(ins) != 1 || strings.Trim(pair+ins, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			return nil, fmt.Errorf("%s: malformed rule %q", path, line)
		}
		p.insert[pair[0]-'A'][pair[1]-'A'] = ins[0]
	}
	return p, nil
}

// spread runs the steps and returns most minus least common counts.
func (p *polymer) spread(steps int) (int, error) {
	if steps < 0 || steps > 100 {
		return 0, fmt.Errorf("%d steps is out of range", steps)
	}
	var pairs, next [letters][letters]int
	for i := 0; i+1 < len(p.template); i++ {
		pairs[p.template[i]-'A'][p.template[i+1]-'A']++
	}
	for range steps {
		next = [letters][letters]int{}
		for a := range letters {
			for b := range letters {
				c := p.insert[a][b]
				if c == 0 {
					next[a][b] += pairs[a][b]
					continue
				}
				next[a][c-'A'] += pairs[a][b]
				next[c-'A'][b] += pairs[a][b]
			}
		}
		pairs = next
	}
	var count [letters]int
	count[p.template[len(p.template)-1]-'A']++
	for a := range letters {
		for b := range letters {
			count[a] += pairs[a][b]
		}
	}
	most, least, total := 0, math.MaxInt, 0
	for _, c := range count {
		total += c
		if c > 0 {
			most, least = max(most, c), min(least, c)
		}
	}
	if least == math.MaxInt || least < 0 {
		return 0, errors.New("no elements, or the counts overflowed")
	}
	if total < len(p.template) {
		return 0, fmt.Errorf("the polymer shrank to %d elements", total)
	}
	return most - least, nil
}

func day14(path string) error {
	p, err := load(path)
	if err != nil {
		return err
	}
	short, err := p.spread(10)
	if err != nil {
		return err
	}
	long, err := p.spread(40)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", short, long); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
