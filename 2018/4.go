// Day 4: Repose Record
//
// Each line of the input is a timestamped record, in no particular order:
// a guard beginning a shift, falling asleep, or waking up. All sleeping
// happens within the midnight hour.
//
// Part 1 picks the guard who sleeps the most minutes in total, and the
// minute that guard is most often asleep, and multiplies the two.
// Part 2 picks the guard and minute with the most nights asleep at that
// minute, and multiplies those.
//
// The records are sorted by their timestamps (which sort as plain text) with
// an insertion sort, then walked once to fill a table of how often each
// guard is asleep in each minute.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 20
	maxLines      = 5_000
	maxGuards     = 1_000
	maxGuardID    = 100_000
)

// guards holds the sleep table: asleep[g][m] counts the nights guard g was
// asleep at minute m.
type guards struct {
	ids    [maxGuards]int
	asleep [maxGuards][60]int
	n      int
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

// sortRecords sorts the records by timestamp in place.
func sortRecords(lines []string) error {
	if len(lines) > maxLines {
		return fmt.Errorf("%d records, more than %d", len(lines), maxLines)
	}
	for i := 1; i < len(lines); i++ {
		for j := i; j > 0 && lines[j] < lines[j-1]; j-- {
			lines[j], lines[j-1] = lines[j-1], lines[j]
		}
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			return fmt.Errorf("records out of order at %d", i)
		}
	}
	return nil
}

// index returns the table row of a guard, adding it if new.
func (g *guards) index(id int) (int, error) {
	if id < 0 || id > maxGuardID {
		return 0, fmt.Errorf("guard ID %d is outside 0..%d", id, maxGuardID)
	}
	for i := range g.n {
		if g.ids[i] == id {
			return i, nil
		}
	}
	if g.n == maxGuards {
		return 0, fmt.Errorf("more than %d guards", maxGuards)
	}
	g.ids[g.n] = id
	g.n++
	return g.n - 1, nil
}

// record walks the sorted records and fills the sleep table.
func (g *guards) record(lines []string) error {
	if len(lines) == 0 {
		return errors.New("no records")
	}
	current, since := -1, -1
	for _, line := range lines {
		if len(line) < 19 || line[0] != '[' || line[17] != ']' {
			return fmt.Errorf("malformed record %q", line)
		}
		minute, err := strconv.Atoi(line[15:17])
		if err != nil || minute < 0 || minute > 59 {
			return fmt.Errorf("invalid minute in %q", line)
		}
		event := line[19:]
		switch {
		case strings.HasPrefix(event, "Guard #"):
			if len(strings.Fields(event)) < 2 {
				return fmt.Errorf("malformed shift record %q", line)
			}
			id, err := strconv.Atoi(strings.Fields(event)[1][1:])
			if err != nil {
				return fmt.Errorf("invalid guard in %q: %w", line, err)
			}
			if current, err = g.index(id); err != nil {
				return err
			}
		case event == "falls asleep" && current >= 0:
			since = minute
		case event == "wakes up" && current >= 0 && since >= 0 && since < minute:
			for m := since; m < minute; m++ {
				g.asleep[current][m]++
			}
			since = -1
		default:
			return fmt.Errorf("record %q does not fit the sequence", line)
		}
	}
	return nil
}

// sleepiest returns the guard with the most minutes asleep, and that
// guard's most-slept minute.
func (g *guards) sleepiest() (id, minute int, err error) {
	if g.n == 0 {
		return 0, 0, errors.New("no guards")
	}
	best, bestTotal := 0, -1
	for i := range g.n {
		total := 0
		for _, n := range g.asleep[i] {
			total += n
		}
		if total > bestTotal {
			best, bestTotal = i, total
		}
	}
	if bestTotal < 1 {
		return 0, 0, errors.New("no guard ever sleeps")
	}
	for m := range 60 {
		if g.asleep[best][m] > g.asleep[best][minute] {
			minute = m
		}
	}
	return g.ids[best], minute, nil
}

// mostRegular returns the guard and minute pair with the highest count.
func (g *guards) mostRegular() (id, minute int, err error) {
	if g.n == 0 {
		return 0, 0, errors.New("no guards")
	}
	best, bestCount := 0, -1
	for i := range g.n {
		for m, n := range g.asleep[i] {
			if n > bestCount {
				best, minute, bestCount = i, m, n
			}
		}
	}
	if bestCount < 1 {
		return 0, 0, errors.New("no guard ever sleeps")
	}
	return g.ids[best], minute, nil
}

func day4(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	if err := sortRecords(lines); err != nil {
		return err
	}
	g := new(guards)
	if err := g.record(lines); err != nil {
		return err
	}
	id, minute, err := g.sleepiest()
	if err != nil {
		return err
	}

	id2, minute2, err := g.mostRegular()
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", id*minute, id2*minute2); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
