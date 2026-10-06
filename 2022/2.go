// Day 2: Rock Paper Scissors
//
// Each line of the input is a round: the opponent's shape (A, B, C for
// rock, paper, scissors) and a code (X, Y, Z). A round scores the shape
// played (1-3) plus 0, 3 or 6 for a loss, draw or win.
//
// Part 1 reads X, Y, Z as the shape to play.
// Part 2 reads them as the result wanted: lose, draw or win.
// Each answer is the total score.
//
// With shapes numbered 0-2, the shape that beats s is (s+1) mod 3.
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
	maxInputBytes = 1 << 16
	maxRounds     = 10_000
)

// score is the score of playing me against them (both 0-2).
func score(them, me int) (int, error) {
	if them < 0 || them > 2 || me < 0 || me > 2 {
		return 0, fmt.Errorf("shapes %d and %d are not 0-2", them, me)
	}
	outcome := (me - them + 4) % 3 // 0 loss, 1 draw, 2 win
	return me + 1 + 3*outcome, nil
}

func day2(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxRounds {
		return fmt.Errorf("%s: more than %d rounds", path, maxRounds)
	}
	asShape, asResult := 0, 0
	for _, l := range lines {
		if len(l) != 3 || l[0] < 'A' || l[0] > 'C' || l[2] < 'X' || l[2] > 'Z' {
			return fmt.Errorf("%s: malformed round %q", path, l)
		}
		them, code := int(l[0]-'A'), int(l[2]-'X')
		a, err := score(them, code)
		if err != nil {
			return err
		}
		b, err := score(them, (them+code+2)%3) // lose = them-1, draw = them, win = them+1
		if err != nil {
			return err
		}
		asShape, asResult = asShape+a, asResult+b
	}

	if asShape < len(lines) || asResult < len(lines) {
		return errors.New("a round scored below 1")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", asShape, asResult); err != nil {
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
