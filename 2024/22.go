// Day 22: Monkey Market
//
// Each line of the input is a buyer's initial secret number. Each step,
// the secret is mixed with (xor) and pruned (mod 16777216) after
// multiplying by 64, dividing by 32, and multiplying by 2048. A buyer's
// price is the secret's last digit.
//
// Part 1 sums every buyer's 2000th new secret.
// Part 2: a monkey sells at each buyer's first price that follows a chosen
// sequence of four price changes. The answer is the most bananas any one
// sequence earns.
//
// A sequence of four changes (each -9..9) is a number below 19^4; a table
// totals each sequence's first price per buyer, and a stamp per sequence
// marks which buyer last saw it.
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
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 16
	steps         = 2000
	prune         = 1 << 24
	sequences     = 19 * 19 * 19 * 19
)

// next returns the next secret.
func next(s int) int {
	s = (s ^ s*64) % prune
	s = (s ^ s/32) % prune
	return (s ^ s*2048) % prune
}

// market totals bananas per sequence for one buyer (number id).
func market(secret, id int, total, seen []int) (int, error) {
	if len(total) != sequences || len(seen) != sequences || secret < 0 {
		return 0, errors.New("wrong table sizes or a negative secret")
	}
	key, price := 0, secret%10
	for k := 1; k <= steps; k++ {
		secret = next(secret)
		p := secret % 10
		key = (key*19 + p - price + 9) % sequences
		price = p
		if k >= 4 && seen[key] != id {
			seen[key], total[key] = id, total[key]+p
		}
	}
	if secret < 0 || secret >= prune {
		return 0, fmt.Errorf("secret %d escaped the pruning", secret)
	}
	return secret, nil
}

func day22(path string) error {
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
	total, seen := make([]int, sequences), make([]int, sequences)
	sum, buyers := 0, strings.Fields(string(content))
	if len(buyers) == 0 {
		return fmt.Errorf("%s: no buyers", path)
	}
	for i, f := range buyers {
		s, err := strconv.Atoi(f)
		if err != nil || s < 0 || s >= prune {
			return fmt.Errorf("%s: invalid secret %q", path, f)
		}
		last, err := market(s, i+1, total, seen)
		if err != nil {
			return err
		}
		sum += last
	}
	best := 0
	for _, t := range total {
		best = max(best, t)
	}

	if sum < 1 {
		return errors.New("every secret ended at zero")
	}
	if best > 9*len(buyers) {
		return fmt.Errorf("%d bananas from %d buyers", best, len(buyers))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, best); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
