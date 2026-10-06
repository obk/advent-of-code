// Day 5: Doesn't He Have Intern-Elves For This?
//
// Each line of the input is a string to be judged naughty or nice.
//
// Part 1 calls a string nice when it has at least three vowels, at least one
// letter doubled in a row, and none of "ab", "cd", "pq" or "xy".
// Part 2 calls it nice when some pair of letters appears twice without
// overlapping, and some letter repeats with exactly one letter between.
//
// For the non-overlapping pair, the first position of every pair is recorded
// in a 26x26 table: a pair is repeated without overlap when it turns up again
// at least two positions after that first sighting.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxLineLen    = 1_000
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

// checkWord rejects anything that is not a non-empty run of lowercase letters.
func checkWord(s string) error {
	if s == "" || len(s) > maxLineLen {
		return fmt.Errorf("string of %d letters is outside 1..%d", len(s), maxLineLen)
	}
	if strings.Trim(s, "abcdefghijklmnopqrstuvwxyz") != "" {
		return fmt.Errorf("%q holds something other than lowercase letters", s)
	}
	return nil
}

// isNiceOld applies part 1's rules.
func isNiceOld(s string) (bool, error) {
	if s == "" {
		return false, errors.New("empty string")
	}

	vowels, doubled := 0, false
	for i := range len(s) {
		if strings.IndexByte("aeiou", s[i]) >= 0 {
			vowels++
		}
		if i == 0 {
			continue
		}
		switch s[i-1 : i+1] {
		case "ab", "cd", "pq", "xy":
			return false, nil
		}
		doubled = doubled || s[i] == s[i-1]
	}

	if vowels > len(s) {
		return false, fmt.Errorf("counted %d vowels in %d letters", vowels, len(s))
	}
	return vowels >= 3 && doubled, nil
}

// isNiceNew applies part 2's rules.
func isNiceNew(s string) (bool, error) {
	if s == "" || len(s) > maxLineLen {
		return false, fmt.Errorf("string of %d letters is outside 1..%d", len(s), maxLineLen)
	}

	// firstAt[a][b] is one more than where the pair ab first starts, so zero
	// means not seen yet.
	var firstAt [26][26]int
	pairTwice, sandwich := false, false
	for i := 1; i < len(s); i++ {
		a, b := s[i-1]-'a', s[i]-'a'
		if a >= 26 || b >= 26 {
			return false, fmt.Errorf("%q holds a non-letter", s)
		}
		switch first := firstAt[a][b]; {
		case first == 0:
			firstAt[a][b] = i
		case i-first >= 2:
			pairTwice = true
		}
		sandwich = sandwich || (i >= 2 && s[i] == s[i-2])
	}
	// Two non-overlapping pairs take at least four letters.
	if pairTwice && len(s) < 4 {
		return false, fmt.Errorf("found a repeated pair in only %d letters", len(s))
	}
	return pairTwice && sandwich, nil
}

func day5(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	niceOld, niceNew, words := 0, 0, 0
	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word == "" {
			continue
		}
		words++
		if err := checkWord(word); err != nil {
			return err
		}
		old, err := isNiceOld(word)
		if err != nil {
			return err
		}
		updated, err := isNiceNew(word)
		if err != nil {
			return err
		}
		if old {
			niceOld++
		}
		if updated {
			niceNew++
		}
	}

	if words == 0 {
		return fmt.Errorf("%s: no strings found", path)
	}
	if niceOld > words || niceNew > words {
		return fmt.Errorf("more nice strings than the %d strings", words)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", niceOld, niceNew); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
