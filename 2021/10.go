// Day 10: Syntax Scoring
//
// Each line of the input is a chunk of brackets (), [], {}, <>.
//
// Part 1 sums the scores (3, 57, 1197, 25137) of the first wrong closing
// bracket in each corrupted line.
// Part 2 completes each incomplete line: the score of a completion is
// built as score*5 + (1, 2, 3, 4) per bracket; the answer is the median.
//
// Each line is checked with a stack of open brackets; the scores are
// sorted with an insertion sort.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	maxLine       = 256
)

// check returns a line's corruption score, or (if it is only incomplete)
// its completion score.
func check(line string) (corrupt, complete int, err error) {
	if len(line) > maxLine {
		return 0, 0, fmt.Errorf("line of %d is longer than %d", len(line), maxLine)
	}
	var stack [maxLine]byte
	n := 0
	for k := range len(line) {
		c := line[k]
		if open := strings.IndexByte("([{<", c); open >= 0 {
			stack[n], n = c, n+1
			continue
		}
		closing := strings.IndexByte(")]}>", c)
		if closing < 0 {
			return 0, 0, fmt.Errorf("unexpected %q", c)
		}
		if n == 0 || stack[n-1] != "([{<"[closing] {
			return [4]int{3, 57, 1197, 25137}[closing], 0, nil
		}
		n--
	}
	for k := n; k > 0; k-- {
		complete = complete*5 + strings.IndexByte("([{<", stack[k-1]) + 1
	}
	return 0, complete, nil
}

func day10(path string) error {
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
	lines := strings.Fields(string(content))
	if len(lines) > maxLines {
		return fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	scores := make([]int, 0, len(lines))
	syntax := 0
	for _, line := range lines {
		bad, done, err := check(line)
		if err != nil {
			return err
		}
		syntax += bad
		if done == 0 {
			continue
		}
		scores = append(scores, done)
		for k := len(scores) - 1; k > 0 && scores[k-1] > scores[k]; k-- {
			scores[k-1], scores[k] = scores[k], scores[k-1]
		}
	}
	if len(scores)%2 == 0 {
		return fmt.Errorf("%d incomplete lines; the median needs an odd count", len(scores))
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", syntax, scores[len(scores)/2]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
