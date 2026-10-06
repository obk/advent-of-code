// Day 3: Lobby
//
// Each line of the input is a bank of digits. A code is read off a bank by
// deleting digits and keeping what is left in its original order; the code
// wanted is always the largest one of a given length.
//
// Part 1 reads a 2-digit code from every bank and sums them.
// Part 2 reads a 12-digit code instead, which pushes the total into the
// trillions and so needs int64.
//
// Both parts use the same greedy scan rather than trying every combination.
// The first digit of the code should be as big as possible, and the largest
// digit that still leaves enough digits after it to finish the code is the
// one to take - the leftmost such digit if there is a tie, which keeps the
// most options open. Each later digit is picked the same way from what comes
// after the one before it.
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
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxBankLen    = 1_000
)

// Part 1 reads two digits off each bank, part 2 reads twelve.
const (
	shortCode = 2
	longCode  = 12
)

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

// parseInput returns the banks, each checked to hold only digits.
func parseInput(path string) ([]string, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	banks := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > maxBankLen {
			return nil, fmt.Errorf("bank of %d digits is longer than %d", len(line), maxBankLen)
		}
		if strings.Trim(line, "0123456789") != "" {
			return nil, fmt.Errorf("bank %q holds something other than digits", line)
		}
		banks = append(banks, line)
	}
	if len(banks) == 0 {
		return nil, fmt.Errorf("%s: no banks found", path)
	}
	return banks, nil
}

// largestSubsequence returns the biggest number that can be formed by deleting
// digits from bank while keeping the rest in order, leaving exactly k of them.
//
// Each digit of the answer is the largest one available between just after
// the previous pick and the last position that still leaves enough digits to
// finish the code.
func largestSubsequence(bank string, k int) (int64, error) {
	if k < 1 || k > longCode {
		return 0, fmt.Errorf("code length %d is outside 1..%d", k, longCode)
	}
	if len(bank) < k || len(bank) > maxBankLen {
		return 0, fmt.Errorf("bank of %d digits cannot give a %d-digit code", len(bank), k)
	}

	var value int64
	from := 0
	for pick := range k {
		// The digits after this pick must still cover the k-pick-1 to come.
		last := len(bank) - (k - pick)
		best := from
		for i := from + 1; i <= last; i++ {
			if bank[i] > bank[best] {
				best = i
			}
		}
		if bank[best] < '0' || bank[best] > '9' {
			return 0, fmt.Errorf("bank %q holds a non-digit", bank)
		}
		value = value*10 + int64(bank[best]-'0')
		from = best + 1
	}
	return value, nil
}

func day3(path string) error {
	banks, err := parseInput(path)
	if err != nil {
		return err
	}

	// Part 1 sums stay small; part 2 reaches into the trillions.
	var part1Total, part2Total int64
	for _, bank := range banks {
		if len(bank) >= shortCode {
			value, err := largestSubsequence(bank, shortCode)
			if err != nil {
				return err
			}
			part1Total += value
		}

		if len(bank) >= longCode {
			value, err := largestSubsequence(bank, longCode)
			if err != nil {
				return err
			}
			if part2Total > math.MaxInt64-value {
				return errors.New("part 2 total overflows int64")
			}
			part2Total += value
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", part1Total, part2Total); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day3(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
