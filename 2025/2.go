// Day 2: Gift Shop
//
// The input is a single line of comma-separated inclusive ID ranges, such as
// "11-22,95-115". An ID is invalid when its digits are one sequence written
// over and over, and the answer is the sum of the invalid IDs.
//
// Part 1 counts an ID as invalid when the sequence appears exactly twice:
// 55 and 123123 qualify, 101 and 12123 do not.
// Part 2 loosens that to two or more repeats, so 123123123 and 1111111 count
// as well.
//
// Every ID in every range is tested. A repeated block has to divide the digit
// count exactly, so a candidate block length k only needs one comparison: take
// the leading k digits and write them out len/k times. That is done with
// arithmetic rather than strings, so testing an ID allocates nothing.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 16
	maxRanges     = 1_000
	maxDigits     = 15
	maxID         = 999_999_999_999_999 // the largest ID of maxDigits digits
	maxRangeSpan  = 10_000_000
)

// idRange is an inclusive range of gift IDs, written "start-end" in the input.
type idRange struct {
	start int64
	end   int64
}

// parseRange reads a range written as "11-22".
func parseRange(field string) (idRange, error) {
	start, end, found := strings.Cut(field, "-")
	if !found {
		return idRange{}, fmt.Errorf("malformed range %q", field)
	}

	lo, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return idRange{}, fmt.Errorf("invalid range %q: %w", field, err)
	}
	hi, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return idRange{}, fmt.Errorf("invalid range %q: %w", field, err)
	}

	if lo < 1 || hi > maxID {
		return idRange{}, fmt.Errorf("range %q is outside 1..%d", field, int64(maxID))
	}
	if lo > hi {
		return idRange{}, fmt.Errorf("range %q runs backwards", field)
	}
	if hi-lo >= maxRangeSpan {
		return idRange{}, fmt.Errorf("range %q spans more than %d IDs", field, maxRangeSpan)
	}
	return idRange{start: lo, end: hi}, nil
}

// parseInput reads the single comma-separated line of ID ranges.
func parseInput(path string) ([]idRange, error) {
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

	fields := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(fields) > maxRanges {
		return nil, fmt.Errorf("%s: %d ranges, more than the %d allowed", path, len(fields), maxRanges)
	}

	ranges := make([]idRange, len(fields))
	for i, field := range fields {
		ranges[i], err = parseRange(field)
		if err != nil {
			return nil, err
		}
	}
	return ranges, nil
}

// digitCount returns how many decimal digits id has.
func digitCount(id int64) (int, error) {
	if id < 1 || id > maxID {
		return 0, fmt.Errorf("ID %d is outside 1..%d", id, int64(maxID))
	}

	count := 0
	for range maxDigits {
		if id == 0 {
			break
		}
		id /= 10
		count++
	}
	if id != 0 {
		return 0, fmt.Errorf("ID has more than %d digits", maxDigits)
	}
	return count, nil
}

// repeatsBlock reports whether id, which has the given number of digits,
// consists solely of copies of its leading blockLen digits: 123123 repeats a
// block of length 3.
func repeatsBlock(id int64, digits, blockLen int) (bool, error) {
	if digits < 1 || digits > maxDigits {
		return false, fmt.Errorf("digit count %d is outside 1..%d", digits, maxDigits)
	}
	if blockLen < 1 || blockLen > digits {
		return false, fmt.Errorf("block length %d is outside 1..%d", blockLen, digits)
	}
	if digits%blockLen != 0 {
		return false, nil
	}

	// blockScale = 10^blockLen; dropScale = 10^(digits-blockLen) leaves just
	// the leading block when id is divided by it.
	var blockScale, dropScale int64 = 1, 1
	for i := range digits {
		if i < blockLen {
			blockScale *= 10
		} else {
			dropScale *= 10
		}
	}

	block := id / dropScale
	var rebuilt int64
	for range digits / blockLen {
		rebuilt = rebuilt*blockScale + block
	}
	return rebuilt == id, nil
}

// classify reports whether id is one sequence written exactly twice (part 1)
// and whether it is one sequence written two or more times (part 2).
func classify(id int64) (doubled, repeated bool, err error) {
	digits, err := digitCount(id)
	if err != nil {
		return false, false, err
	}

	if digits%2 == 0 {
		doubled, err = repeatsBlock(id, digits, digits/2)
		if err != nil {
			return false, false, err
		}
	}

	// A block that repeats at least twice can be at most half as long as id.
	for blockLen := 1; blockLen <= digits/2 && !repeated; blockLen++ {
		repeated, err = repeatsBlock(id, digits, blockLen)
		if err != nil {
			return false, false, err
		}
	}

	if doubled && !repeated {
		return false, false, fmt.Errorf("ID %d is doubled but not repeated", id)
	}
	return doubled, repeated, nil
}

// sumInvalid adds up the invalid IDs in one range, by both parts' rules.
func sumInvalid(r idRange) (doubledSum, repeatedSum int64, err error) {
	if r.start < 1 || r.start > r.end || r.end > maxID {
		return 0, 0, fmt.Errorf("invalid range %d-%d", r.start, r.end)
	}
	if r.end-r.start >= maxRangeSpan {
		return 0, 0, fmt.Errorf("range %d-%d is too wide", r.start, r.end)
	}

	for id := r.start; id <= r.end; id++ {
		doubled, repeated, err := classify(id)
		if err != nil {
			return 0, 0, err
		}
		if doubled {
			doubledSum += id
		}
		if repeatedSum > math.MaxInt64-id {
			return 0, 0, fmt.Errorf("sum of range %d-%d overflows int64", r.start, r.end)
		}
		if repeated {
			repeatedSum += id
		}
	}
	return doubledSum, repeatedSum, nil
}

func day2(path string) error {
	ranges, err := parseInput(path)
	if err != nil {
		return err
	}

	// The doubled sum never exceeds the repeated one, so guarding the
	// repeated sum against overflow covers both.
	var doubledSum, repeatedSum int64
	for _, r := range ranges {
		doubled, repeated, err := sumInvalid(r)
		if err != nil {
			return err
		}
		if repeatedSum > math.MaxInt64-repeated {
			return errors.New("part 2 sum overflows int64")
		}
		doubledSum += doubled
		repeatedSum += repeated
	}

	if doubledSum > repeatedSum {
		return fmt.Errorf("doubled sum %d exceeds repeated sum %d", doubledSum, repeatedSum)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", doubledSum, repeatedSum); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day2(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
