// Day 16: Ticket Translation
//
// The input is ticket field rules ("class: 1-3 or 5-7"), your ticket, and
// nearby tickets, each a comma-separated list of values.
//
// Part 1 sums the nearby values that fit no rule at all.
// Part 2 drops the tickets with such values, works out which position is
// which field, and multiplies your ticket's six "departure" fields.
//
// For each field, a bit set holds the positions where every valid ticket
// fits its rule. Repeatedly, a field left with one position takes it and
// that position is removed from the others.
package main

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 16
	maxFields     = 32
	maxTickets    = 1_000
	maxValue      = 1_000_000
)

// rule is a field's name and its two ranges.
type rule struct {
	name           string
	a1, b1, a2, b2 int
}

// notes is the parsed input.
type notes struct {
	rules  []rule
	mine   []int
	nearby [][]int
}

// parseTicket reads one ticket of n values.
func parseTicket(text string, n int) ([]int, error) {
	f := strings.Split(text, ",")
	if len(f) != n {
		return nil, fmt.Errorf("ticket %q has %d values, not %d", text, len(f), n)
	}
	values := make([]int, n)
	for i, v := range f {
		x, err := strconv.Atoi(v)
		if err != nil || x < 0 || x > maxValue {
			return nil, fmt.Errorf("invalid value %q", v)
		}
		values[i] = x
	}
	return values, nil
}

// load parses the three sections.
func load(path string) (*notes, error) {
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
	sections := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(sections) != 3 {
		return nil, fmt.Errorf("%s: %d sections, expected 3", path, len(sections))
	}
	ruleLines := strings.Split(sections[0], "\n")
	if len(ruleLines) > maxFields {
		return nil, fmt.Errorf("%s: more than %d fields", path, maxFields)
	}
	n := &notes{rules: make([]rule, len(ruleLines))}
	for i, l := range ruleLines {
		name, ranges, found := strings.Cut(l, ": ")
		r := rule{name: name}
		k, err := fmt.Sscanf(ranges, "%d-%d or %d-%d", &r.a1, &r.b1, &r.a2, &r.b2)
		if !found || err != nil || k != 4 {
			return nil, fmt.Errorf("%s: malformed rule %q", path, l)
		}
		n.rules[i] = r
	}
	mine := strings.Split(sections[1], "\n")
	near := strings.Split(sections[2], "\n")
	if len(mine) != 2 || len(near) < 2 || len(near) > maxTickets {
		return nil, fmt.Errorf("%s: malformed ticket sections", path)
	}
	if n.mine, err = parseTicket(mine[1], len(n.rules)); err != nil {
		return nil, err
	}
	n.nearby = make([][]int, len(near)-1)
	for i, t := range near[1:] {
		if n.nearby[i], err = parseTicket(t, len(n.rules)); err != nil {
			return nil, err
		}
	}
	return n, nil
}

// fits reports whether v fits rule r.
func (r rule) fits(v int) bool {
	return (v >= r.a1 && v <= r.b1) || (v >= r.a2 && v <= r.b2)
}

// scan sums the values fitting no rule, and marks tickets with any.
func (n *notes) scan(bad []bool) (int, error) {
	if len(bad) != len(n.nearby) {
		return 0, errors.New("scratch size does not match the tickets")
	}
	sum := 0
	for t, ticket := range n.nearby {
		for _, v := range ticket {
			ok := false
			for _, r := range n.rules {
				ok = ok || r.fits(v)
			}
			if !ok {
				sum, bad[t] = sum+v, true
			}
		}
	}
	return sum, nil
}

// assign works out each field's position, into pos.
func (n *notes) assign(bad []bool, pos []int) error {
	if len(pos) != len(n.rules) || len(bad) != len(n.nearby) {
		return errors.New("scratch sizes do not match")
	}
	var possible [maxFields]uint32
	all := uint32(1)<<len(n.rules) - 1
	for f, r := range n.rules {
		possible[f] = all
		for t, ticket := range n.nearby {
			for p, v := range ticket {
				if !bad[t] && !r.fits(v) {
					possible[f] &^= 1 << p
				}
			}
		}
	}
	for i := range pos {
		pos[i] = -1
	}
	for range n.rules {
		for f := range n.rules {
			if pos[f] >= 0 || bits.OnesCount32(possible[f]) != 1 {
				continue
			}
			pos[f] = bits.TrailingZeros32(possible[f])
			for g := range n.rules {
				possible[g] &^= 1 << pos[f]
			}
		}
	}
	used := uint32(0)
	for f, p := range pos {
		if p < 0 {
			return fmt.Errorf("field %q has no single position", n.rules[f].name)
		}
		if used>>p&1 == 1 {
			return fmt.Errorf("two fields share position %d", p)
		}
		used |= 1 << p
	}
	return nil
}

func day16(path string) error {
	n, err := load(path)
	if err != nil {
		return err
	}
	bad := make([]bool, len(n.nearby))
	rate, err := n.scan(bad)
	if err != nil {
		return err
	}
	pos := make([]int, len(n.rules))
	if err := n.assign(bad, pos); err != nil {
		return err
	}
	product, found := 1, 0
	for f, r := range n.rules {
		if strings.HasPrefix(r.name, "departure") {
			if v := n.mine[pos[f]]; v > 0 && product > math.MaxInt64/v {
				return errors.New("the product overflows")
			}
			product, found = product*n.mine[pos[f]], found+1
		}
	}
	if found != 6 {
		return fmt.Errorf("%d departure fields, expected 6", found)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", rate, product); err != nil {
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
