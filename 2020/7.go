// Day 7: Handy Haversacks
//
// Each line of the input is a rule: "light red bags contain 1 bright white
// bag, 2 muted yellow bags." (or "contain no other bags.").
//
// Part 1 counts the colours that can eventually hold a shiny gold bag.
// Part 2 counts the bags inside one shiny gold bag.
//
// Both are worked out by passes over the rules until nothing changes: a
// colour holds gold once any colour inside it does, and a colour's total
// is known once those of everything inside it are. The rules form no
// cycle, so there are at most as many passes as colours.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 16
	maxColours    = 1_000
	maxInside     = 8
)

// content is a number of bags of one colour.
type content struct {
	count, colour int
}

// rules holds what each colour contains.
type rules struct {
	ids    map[string]int
	inside [][]content
}

// id returns a colour's number, numbering it if new.
func (r *rules) id(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty colour")
	}
	if i, ok := r.ids[name]; ok {
		return i, nil
	}
	if len(r.inside) == cap(r.inside) {
		return 0, fmt.Errorf("more than %d colours", cap(r.inside))
	}
	r.ids[name] = len(r.inside)
	r.inside = append(r.inside, nil)
	return r.ids[name], nil
}

// addRule parses one rule.
func (r *rules) addRule(line string) error {
	outer, rest, found := strings.Cut(line, " bags contain ")
	if !found {
		return fmt.Errorf("malformed rule %q", line)
	}
	o, err := r.id(outer)
	if err != nil {
		return err
	}
	if r.inside[o] != nil {
		return fmt.Errorf("two rules for %q", outer)
	}
	r.inside[o] = make([]content, 0, maxInside)
	if rest == "no other bags." {
		return nil
	}
	for _, part := range strings.Split(strings.TrimSuffix(rest, "."), ", ") {
		f := strings.Fields(part)
		if len(f) != 4 || len(r.inside[o]) == maxInside {
			return fmt.Errorf("malformed contents %q", part)
		}
		n, err := strconv.Atoi(f[0])
		if err != nil || n < 1 {
			return fmt.Errorf("invalid count in %q", part)
		}
		c, err := r.id(f[1] + " " + f[2])
		if err != nil {
			return err
		}
		r.inside[o] = append(r.inside[o], content{n, c})
	}
	return nil
}

// load parses every rule.
func load(path string) (*rules, error) {
	if path == "" {
		return nil, errors.New("no input path given")
	}
	text, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(text) == 0 || len(text) > maxInputBytes {
		return nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(text), maxInputBytes)
	}
	r := &rules{ids: make(map[string]int, maxColours), inside: make([][]content, 0, maxColours)}
	for _, line := range strings.Split(strings.TrimSpace(string(text)), "\n") {
		if err := r.addRule(line); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	if len(r.inside) == 0 {
		return nil, fmt.Errorf("%s: no rules", path)
	}
	return r, nil
}

// holders counts colours that eventually hold the target.
func (r *rules) holders(target int, holds []bool) (int, error) {
	if target < 0 || target >= len(r.inside) || len(holds) != len(r.inside) {
		return 0, errors.New("unknown colour or wrong scratch size")
	}
	for range len(r.inside) {
		changed := false
		for c, list := range r.inside {
			for _, in := range list {
				if !holds[c] && (in.colour == target || holds[in.colour]) {
					holds[c], changed = true, true
				}
			}
		}
		if !changed {
			break
		}
	}
	count := 0
	for _, h := range holds {
		if h {
			count++
		}
	}
	if holds[target] {
		return 0, errors.New("the target can hold itself")
	}
	return count, nil
}

// total counts the bags inside the target, using totals as scratch (-1
// for not yet known).
func (r *rules) total(target int, totals []int) (int, error) {
	if target < 0 || target >= len(r.inside) || len(totals) != len(r.inside) {
		return 0, errors.New("unknown colour or wrong scratch size")
	}
	for i := range totals {
		totals[i] = -1
	}
	for range len(r.inside) + 1 {
		for c, list := range r.inside {
			sum := 0
			for _, in := range list {
				if sum >= 0 && totals[in.colour] >= 0 {
					sum += in.count * (1 + totals[in.colour])
					if sum < 0 {
						return 0, errors.New("the bag count overflowed")
					}
				} else {
					sum = -1
				}
			}
			totals[c] = sum
		}
		if totals[target] >= 0 {
			return totals[target], nil
		}
	}
	return 0, errors.New("the rules for the target form a cycle")
}

func day7(path string) error {
	r, err := load(path)
	if err != nil {
		return err
	}
	gold, ok := r.ids["shiny gold"]
	if !ok {
		return errors.New("no rule mentions shiny gold")
	}
	if r.inside[gold] == nil {
		return errors.New("no rule says what shiny gold bags contain")
	}
	holding, err := r.holders(gold, make([]bool, len(r.inside)))
	if err != nil {
		return err
	}
	inside, err := r.total(gold, make([]int, len(r.inside)))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", holding, inside); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
