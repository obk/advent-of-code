// Day 9: Stream Processing
//
// The input is a stream of nested groups {...} separated by commas, with
// garbage <...> mixed in. Inside garbage, '!' cancels the next character and
// nothing else is special until the closing '>'.
//
// Part 1 is the total score: each group scores one more than the group
// around it, the outermost scoring 1.
// Part 2 counts the characters inside garbage, not counting the brackets,
// the '!'s or the characters they cancel.
//
// One pass with a depth counter and a "in garbage" flag.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 20
)

// readStream returns the stream.
func readStream(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	stream := strings.TrimSpace(string(content))
	if stream == "" {
		return "", fmt.Errorf("%s: empty stream", path)
	}
	return stream, nil
}

// score scans the stream and returns the total group score and the amount
// of garbage.
func score(stream string) (total, junk int, err error) {
	if stream == "" {
		return 0, 0, errors.New("empty stream")
	}
	depth, garbage := 0, false
	for i := 0; i < len(stream); i++ {
		c := stream[i]
		switch {
		case garbage && c == '!':
			i++ // skip the cancelled character
		case garbage && c == '>':
			garbage = false
		case garbage:
			junk++
		case c == '<':
			garbage = true
		case c == '{':
			depth++
			total += depth
		case c == '}':
			if depth == 0 {
				return 0, 0, fmt.Errorf("unmatched '}' at %d", i)
			}
			depth--
		case c == ',':
		default:
			return 0, 0, fmt.Errorf("unexpected %q outside garbage at %d", c, i)
		}
	}
	if depth != 0 || garbage {
		return 0, 0, fmt.Errorf("stream ends inside %d groups (garbage=%v)", depth, garbage)
	}
	return total, junk, nil
}

func day9(path string) error {
	stream, err := readStream(path)
	if err != nil {
		return err
	}
	total, junk, err := score(stream)
	if err != nil {
		return err
	}

	if junk > len(stream) {
		return fmt.Errorf("%d garbage characters in a stream of %d", junk, len(stream))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, junk); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
