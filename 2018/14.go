// Day 14: Chocolate Charts
//
// A scoreboard starts with recipes 3 and 7, and two elves on them. Each
// round the elves add the digits of their recipes' sum as new recipes, then
// each steps forward 1 + its recipe's score. The input is a number n.
//
// Part 1 is the ten scores after the first n recipes.
// Part 2 reads the input as a sequence of digits and asks how many recipes
// come before it first appears on the scoreboard.
//
// The scoreboard is one fixed byte array, filled as the rounds go.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 10
	maxRecipes    = 50_000_000
)

// board is the scoreboard and the elves' positions.
type board struct {
	scores []byte
	a, b   int
}

// readCount returns the input number, both as a number and as its digits.
func readCount(path string) (int, string, error) {
	if path == "" {
		return 0, "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	text := strings.TrimSpace(string(content))
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, "", fmt.Errorf("%s: invalid number: %w", path, err)
	}
	if n < 1 || n+10 > maxRecipes {
		return 0, "", fmt.Errorf("%s: %d is outside 1..%d", path, n, maxRecipes-10)
	}
	return n, text, nil
}

// grow runs rounds until the board holds at least count recipes.
func (bd *board) grow(count int) error {
	if count < 0 || len(bd.scores) < 2 {
		return fmt.Errorf("cannot grow a board of %d to %d recipes", len(bd.scores), count)
	}
	if count > cap(bd.scores) {
		return fmt.Errorf("board of %d cannot hold %d recipes", cap(bd.scores), count)
	}
	for i := 0; i < maxRecipes && len(bd.scores) < count && len(bd.scores)+2 <= cap(bd.scores); i++ {
		sum := bd.scores[bd.a] + bd.scores[bd.b]
		if sum >= 10 {
			bd.scores = append(bd.scores, 1)
		}
		bd.scores = append(bd.scores, sum%10)
		bd.a = (bd.a + 1 + int(bd.scores[bd.a])) % len(bd.scores)
		bd.b = (bd.b + 1 + int(bd.scores[bd.b])) % len(bd.scores)
	}
	if len(bd.scores) < count {
		return fmt.Errorf("the board filled up before %d recipes", count)
	}
	if bd.a < 0 || bd.b < 0 || bd.a >= len(bd.scores) || bd.b >= len(bd.scores) {
		return fmt.Errorf("elves at %d and %d are off the board", bd.a, bd.b)
	}
	return nil
}

// findSequence grows the board until the digits of seq appear, and returns
// how many recipes come before them.
func (bd *board) findSequence(seq string) (int, error) {
	if seq == "" || len(seq) > 20 || strings.Trim(seq, "0123456789") != "" {
		return 0, fmt.Errorf("%q is not a short run of digits", seq)
	}
	if len(bd.scores) < 2 {
		return 0, errors.New("the board is not set up")
	}
	checked := 0 // every start position below this has been ruled out
	for round := 0; round < maxRecipes; round++ {
		for k := 0; k < maxRecipes && checked+len(seq) <= len(bd.scores); k++ {
			match := true
			for d := 0; d < len(seq) && match; d++ {
				match = bd.scores[checked+d] == seq[d]-'0'
			}
			if match {
				return checked, nil
			}
			checked++
		}
		if err := bd.grow(len(bd.scores) + 1); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("%s does not appear within %d recipes", seq, maxRecipes)
}

func day14(path string) error {
	n, digits, err := readCount(path)
	if err != nil {
		return err
	}
	// The board starts with 3 and 7 and has room for maxRecipes.
	bd := board{scores: make([]byte, 2, maxRecipes), a: 0, b: 1}
	bd.scores[0], bd.scores[1] = 3, 7
	if err := bd.grow(n + 10); err != nil {
		return err
	}
	if n+10 > len(bd.scores) {
		return fmt.Errorf("board holds %d recipes, need %d", len(bd.scores), n+10)
	}
	var next [10]byte
	for i := range next {
		next[i] = '0' + bd.scores[n+i]
	}

	before, err := bd.findSequence(digits)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", next[:], before); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
