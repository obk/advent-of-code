// Day 13: Knights of the Dinner Table
//
// Each line of the input says how much one guest's happiness changes when
// seated next to another, as "Alice would gain 54 happiness units by sitting
// next to Bob." (or "lose"). Everyone sits at one round table.
//
// Part 1 is the best possible total change in happiness.
// Part 2 adds yourself, at zero happiness either way with everyone.
//
// Only the sum of the two directions matters for a pair of neighbours. With
// a round table one guest can be fixed in seat 0, which leaves a Held-Karp
// dynamic programme over the rest: best[set][last] is the happiest chain from
// guest 0 through exactly the guests in set, ending at last. Closing the
// circle adds the pair (last, 0).
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
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxGuests     = 12 // including yourself
	maxHappiness  = 10_000
)

// table holds the combined happiness of every pair of guests.
type table struct {
	n    int
	pair [maxGuests][maxGuests]int
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

// guestID returns the number of the named guest, numbering them if new.
func guestID(ids map[string]int, name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty guest name")
	}
	if id, ok := ids[name]; ok {
		return id, nil
	}
	if len(ids) >= maxGuests-1 { // leave a seat for yourself
		return 0, fmt.Errorf("more than %d guests", maxGuests-1)
	}
	ids[name] = len(ids)
	return ids[name], nil
}

func parseInput(path string) (table, error) {
	lines, err := readLines(path)
	if err != nil {
		return table{}, err
	}

	var t table
	pairings := 0
	ids := make(map[string]int, maxGuests)
	for _, line := range lines {
		words := strings.Fields(strings.TrimSuffix(strings.TrimSpace(line), "."))
		if len(words) == 0 {
			continue
		}
		if len(words) != 11 || (words[2] != "gain" && words[2] != "lose") {
			return table{}, fmt.Errorf("malformed line %q", line)
		}
		a, err := guestID(ids, words[0])
		if err != nil {
			return table{}, err
		}
		b, err := guestID(ids, words[10])
		if err != nil {
			return table{}, err
		}
		units, err := strconv.Atoi(words[3])
		if err != nil {
			return table{}, fmt.Errorf("invalid amount in %q: %w", line, err)
		}
		if units < 0 || units > maxHappiness || a == b {
			return table{}, fmt.Errorf("invalid pairing %q", line)
		}
		if words[2] == "lose" {
			units = -units
		}
		pairings++
		t.pair[a][b] += units
		t.pair[b][a] += units
	}

	t.n = len(ids)
	if pairings != t.n*(t.n-1) {
		return table{}, fmt.Errorf("%s: %d lines for %d guests, expected one per ordered pair", path, pairings, t.n)
	}
	if t.n < 2 {
		return table{}, fmt.Errorf("%s: need at least two guests, found %d", path, t.n)
	}
	return t, nil
}

// bestSeating returns the happiest total for a round table of the first n
// guests. best is scratch space for the Held-Karp table.
func (t *table) bestSeating(n int, best *[1 << maxGuests][maxGuests]int) (int, error) {
	if n < 2 || n > maxGuests {
		return 0, fmt.Errorf("%d guests is outside 2..%d", n, maxGuests)
	}
	if n > t.n+1 {
		return 0, fmt.Errorf("%d seats for %d guests and yourself", n, t.n)
	}

	// Sets are of guests 1..n-1, as bit (guest-1); guest 0 is always first.
	const unreachable = -1 << 40
	all := 1<<(n-1) - 1
	for set := 1; set <= all; set++ {
		for last := 1; last < n; last++ {
			bit := 1 << (last - 1)
			best[set][last] = unreachable
			if set&bit == 0 {
				continue
			}
			rest := set &^ bit
			if rest == 0 {
				best[set][last] = t.pair[0][last]
				continue
			}
			for prev := 1; prev < n; prev++ {
				if rest>>(prev-1)&1 == 1 && best[rest][prev] > unreachable {
					best[set][last] = max(best[set][last], best[rest][prev]+t.pair[prev][last])
				}
			}
		}
	}

	happiest := unreachable
	for last := 1; last < n; last++ {
		if best[all][last] > unreachable {
			happiest = max(happiest, best[all][last]+t.pair[last][0])
		}
	}
	if happiest == unreachable {
		return 0, errors.New("no seating found")
	}
	return happiest, nil
}

func day13(path string) error {
	t, err := parseInput(path)
	if err != nil {
		return err
	}

	best := new([1 << maxGuests][maxGuests]int)
	withoutYou, err := t.bestSeating(t.n, best)
	if err != nil {
		return err
	}
	// You are guest n, with every pair involving you still zero.
	withYou, err := t.bestSeating(t.n+1, best)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", withoutYou, withYou); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
