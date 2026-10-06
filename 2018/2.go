// Day 2: Inventory Management System
//
// Each line of the input is a box ID of lowercase letters.
//
// Part 1 is a checksum: the number of IDs with some letter exactly twice,
// times the number with some letter exactly three times.
// Part 2 finds the two IDs that differ in exactly one position and gives the
// letters they share.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
)

// readIDs returns the box IDs.
func readIDs(path string) ([]string, error) {
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
	ids := strings.Fields(string(content))
	if len(ids) == 0 || len(ids) > maxLines {
		return nil, fmt.Errorf("%s: %d IDs, expected 1..%d", path, len(ids), maxLines)
	}
	for _, id := range ids {
		if strings.Trim(id, "abcdefghijklmnopqrstuvwxyz") != "" {
			return nil, fmt.Errorf("%s: ID %q is not lowercase letters", path, id)
		}
	}
	return ids, nil
}

// twosAndThrees reports whether some letter appears exactly twice, and
// whether some letter appears exactly three times.
func twosAndThrees(id string) (two, three bool, err error) {
	if id == "" {
		return false, false, errors.New("empty ID")
	}
	var counts [26]int
	for i := range len(id) {
		if id[i] < 'a' || id[i] > 'z' {
			return false, false, fmt.Errorf("ID %q holds %q", id, id[i])
		}
		counts[id[i]-'a']++
	}
	for _, n := range counts {
		two, three = two || n == 2, three || n == 3
	}
	return two, three, nil
}

// common finds the one pair of IDs differing in a single position and writes
// their shared letters into out.
func common(ids []string, out []byte) ([]byte, error) {
	if len(ids) < 2 {
		return nil, fmt.Errorf("%d IDs cannot hold a pair", len(ids))
	}
	for i, a := range ids {
		for _, b := range ids[i+1:] {
			if len(a) != len(b) {
				continue
			}
			diff, at := 0, 0
			for k := 0; k < len(a) && diff < 2; k++ {
				if a[k] != b[k] {
					diff, at = diff+1, k
				}
			}
			if diff != 1 {
				continue
			}
			if cap(out) < len(a)-1 {
				return nil, fmt.Errorf("room for %d letters, need %d", cap(out), len(a)-1)
			}
			out = append(append(out[:0], a[:at]...), a[at+1:]...)
			if len(out) != len(a)-1 {
				return nil, fmt.Errorf("kept %d letters of %d", len(out), len(a))
			}
			return out, nil
		}
	}
	return nil, errors.New("no two IDs differ in exactly one position")
}

func day2(path string) error {
	ids, err := readIDs(path)
	if err != nil {
		return err
	}
	twos, threes := 0, 0
	for _, id := range ids {
		two, three, err := twosAndThrees(id)
		if err != nil {
			return err
		}
		if two {
			twos++
		}
		if three {
			threes++
		}
	}

	if twos > len(ids) || threes > len(ids) {
		return fmt.Errorf("%d and %d matches among %d IDs", twos, threes, len(ids))
	}
	shared, err := common(ids, make([]byte, 0, len(ids[0])))
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", twos*threes, shared); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
