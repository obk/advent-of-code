// Day 6: Tuning Trouble
//
// The input is a datastream of letters.
//
// Part 1 is how many characters have been read when the last four were all
// different (a start-of-packet marker).
// Part 2 is the same for fourteen (a start-of-message marker).
//
// A count per letter in the current window, and a count of letters seen
// more than once, slide along the stream.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 14
)

// marker returns how many characters are read when the last size differ.
func marker(stream string, size int) (int, error) {
	if size < 1 || size > 26 {
		return 0, fmt.Errorf("window %d is outside 1..26", size)
	}
	var count [26]int
	repeats := 0
	for i := range len(stream) {
		c := stream[i] - 'a'
		if c >= 26 {
			return 0, fmt.Errorf("character %q is not a-z", stream[i])
		}
		if count[c]++; count[c] == 2 {
			repeats++
		}
		if i >= size {
			old := stream[i-size] - 'a'
			if count[old]--; count[old] == 1 {
				repeats--
			}
		}
		if repeats < 0 {
			return 0, errors.New("negative repeat count")
		}
		if i >= size-1 && repeats == 0 {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("no window of %d distinct letters", size)
}

func day6(path string) error {
	if path == "" {
		return errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	stream := strings.TrimSpace(string(content))
	packet, err := marker(stream, 4)
	if err != nil {
		return err
	}
	message, err := marker(stream, 14)
	if err != nil {
		return err
	}
	if message < packet {
		return fmt.Errorf("message marker %d comes before packet marker %d", message, packet)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", packet, message); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
