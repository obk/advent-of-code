// Day 5: Cafeteria
//
// The input has two sections separated by a blank line: first the inclusive ID
// ranges of the ingredients that are still fresh, then the IDs actually sitting
// in the pantry, one per line.
//
// Part 1 counts the pantry IDs that fall inside at least one fresh range.
// Part 2 ignores the pantry and asks how many distinct IDs the ranges cover
// between them, which is far more than the ranges' sizes added up because they
// overlap heavily.
//
// Part 2 sorts the ranges by their start and merges each one into the previous
// where they touch, leaving a set of disjoint ranges whose sizes can simply be
// summed. Ranges that only sit next to each other merge too: the IDs are whole
// numbers, so 3-5 and 6-8 leave no gap between them. The sort and the merge
// both work in place, so part 2 allocates nothing.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxRanges     = 1_000 // keeps the quadratic insertion sort cheap
)

// idRange is an inclusive range of fresh ingredient IDs.
type idRange struct {
	start int64
	end   int64
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

// parseRange reads a range written as "3-5".
func parseRange(line string) (idRange, error) {
	start, end, found := strings.Cut(line, "-")
	if !found {
		return idRange{}, fmt.Errorf("invalid range format: %s", line)
	}

	lo, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return idRange{}, fmt.Errorf("invalid range %q: %w", line, err)
	}
	hi, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return idRange{}, fmt.Errorf("invalid range %q: %w", line, err)
	}
	if lo < 0 || hi == math.MaxInt64 {
		return idRange{}, fmt.Errorf("range %q is outside 0..%d", line, int64(math.MaxInt64-1))
	}
	if lo > hi {
		return idRange{}, fmt.Errorf("range %q runs backwards", line)
	}
	return idRange{start: lo, end: hi}, nil
}

// parseInput reads the two sections of the input: the fresh ID ranges, then a
// blank line, then the IDs found in the pantry.
func parseInput(path string) (ranges []idRange, ids []int64, err error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, nil, err
	}

	// Either section can be at most the whole file long, so neither slice
	// ever has to grow.
	ranges = make([]idRange, 0, len(lines))
	ids = make([]int64, 0, len(lines))
	readingRanges := true

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			// The one blank line ends the range section.
			if !readingRanges {
				return nil, nil, fmt.Errorf("%s: more than one blank line", path)
			}
			readingRanges = false
		case readingRanges:
			r, err := parseRange(line)
			if err != nil {
				return nil, nil, err
			}
			ranges = append(ranges, r)
		default:
			id, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid ingredient ID %q: %w", line, err)
			}
			ids = append(ids, id)
		}
	}

	if len(ranges) == 0 || len(ranges) > maxRanges {
		return nil, nil, fmt.Errorf("%s: %d ranges, expected 1..%d", path, len(ranges), maxRanges)
	}
	if len(ids) == 0 {
		return nil, nil, fmt.Errorf("%s: no pantry IDs found", path)
	}
	return ranges, ids, nil
}

// countFresh counts the listed IDs that fall inside at least one fresh range.
func countFresh(ids []int64, ranges []idRange) (int, error) {
	if len(ranges) == 0 {
		return 0, errors.New("no ranges to check against")
	}
	if len(ranges) > maxRanges {
		return 0, fmt.Errorf("%d ranges, more than the %d allowed", len(ranges), maxRanges)
	}

	fresh := 0
	for _, id := range ids {
		for _, r := range ranges {
			if id >= r.start && id <= r.end {
				fresh++
				break
			}
		}
	}

	if fresh > len(ids) {
		return 0, fmt.Errorf("counted %d fresh IDs out of %d", fresh, len(ids))
	}
	return fresh, nil
}

// sortByStart sorts the ranges in place by their start, with an insertion
// sort: there are few ranges, and it needs neither recursion nor a comparison
// function.
func sortByStart(ranges []idRange) error {
	if len(ranges) > maxRanges {
		return fmt.Errorf("%d ranges, more than the %d allowed", len(ranges), maxRanges)
	}

	for i := 1; i < len(ranges); i++ {
		for j := i; j > 0 && ranges[j].start < ranges[j-1].start; j-- {
			ranges[j], ranges[j-1] = ranges[j-1], ranges[j]
		}
	}

	for i := 1; i < len(ranges); i++ {
		if ranges[i].start < ranges[i-1].start {
			return fmt.Errorf("ranges out of order at %d after sorting", i)
		}
	}
	return nil
}

// mergeRanges collapses overlapping and adjacent ranges into as few as
// possible, so their sizes can simply be added up. 3-5 and 4-6 overlap; 3-5 and
// 6-8 are adjacent, and over integers that is just as continuous. The merged
// ranges are written over the front of the slice, which is returned shortened.
func mergeRanges(ranges []idRange) ([]idRange, error) {
	if len(ranges) == 0 {
		return nil, errors.New("no ranges to merge")
	}
	if err := sortByStart(ranges); err != nil {
		return nil, err
	}

	last := 0 // index of the merged range currently being extended
	for _, next := range ranges[1:] {
		if next.start < ranges[last].start {
			return nil, fmt.Errorf("range %d-%d is out of order", next.start, next.end)
		}
		if next.start > ranges[last].end+1 {
			last++
			ranges[last] = next
			continue
		}
		if next.end > ranges[last].end {
			ranges[last].end = next.end
		}
	}

	merged := ranges[:last+1]
	for i := 1; i < len(merged); i++ {
		if merged[i].start <= merged[i-1].end+1 {
			return nil, fmt.Errorf("merged ranges %d and %d still touch", i-1, i)
		}
	}
	return merged, nil
}

// countCovered counts how many distinct IDs the ranges cover in total. It
// reorders the ranges.
func countCovered(ranges []idRange) (int64, error) {
	merged, err := mergeRanges(ranges)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, r := range merged {
		size := r.end - r.start + 1
		if total > math.MaxInt64-size {
			return 0, errors.New("covered ID count overflows int64")
		}
		total += size
	}
	return total, nil
}

func day5(path string) error {
	ranges, ids, err := parseInput(path)
	if err != nil {
		return err
	}

	// Part 1 must run first: part 2 merges the ranges in place.
	fresh, err := countFresh(ids, ranges)
	if err != nil {
		return err
	}
	covered, err := countCovered(ranges)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", fresh, covered); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day5(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
