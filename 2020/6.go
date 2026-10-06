// Day 6: Custom Customs
//
// The input is groups separated by blank lines; each line is the questions
// (a-z) one person answered "yes" to.
//
// Part 1 sums, over groups, the questions anyone answered yes to.
// Part 2 sums the questions everyone in the group answered yes to.
//
// Each person's answers are a 26-bit set: the group's union and
// intersection are then OR and AND.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 16
	allQuestions  = 1<<26 - 1
)

// group returns the union and intersection of a group's answers.
func group(text string) (anyone, everyone uint32, err error) {
	people := strings.Fields(text)
	if len(people) == 0 {
		return 0, 0, errors.New("empty group")
	}
	everyone = allQuestions
	for _, p := range people {
		if len(p) > 26 {
			return 0, 0, fmt.Errorf("answers %q repeat a question", p)
		}
		if strings.Trim(p, "abcdefghijklmnopqrstuvwxyz") != "" {
			return 0, 0, fmt.Errorf("answers %q are not a-z", p)
		}
		var set uint32
		for k := range len(p) {
			set |= 1 << (p[k] - 'a')
		}
		anyone, everyone = anyone|set, everyone&set
	}
	return anyone, everyone, nil
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
	sumAny, sumAll := 0, 0
	for _, g := range strings.Split(strings.TrimSpace(string(content)), "\n\n") {
		anyone, everyone, err := group(g)
		if err != nil {
			return err
		}
		if everyone&^anyone != 0 {
			return errors.New("everyone answered a question nobody did")
		}
		sumAny += bits.OnesCount32(anyone)
		sumAll += bits.OnesCount32(everyone)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sumAny, sumAll); err != nil {
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
