// Day 7: Camel Cards
//
// Each line of the input is a hand of five cards and a bid. Hands rank by
// type (five of a kind down to high card), then card by card (A high to 2
// low, with T for ten).
//
// Part 1 sums each bid times its hand's rank (1 for the weakest).
// Part 2 makes J a joker: it counts as whatever card makes the best type,
// but is the weakest card when breaking ties.
//
// Each hand becomes one number, type then cards in base 16, and the hands
// are sorted by insertion sort.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 14
	maxHands      = 2_000
)

// strength returns a hand's sort key.
func strength(hand string, jokers bool) (int, error) {
	order := "23456789TJQKA"
	if jokers {
		order = "J23456789TQKA"
	}
	var count [13]int
	key := 0
	for k := range len(hand) {
		v := strings.IndexByte(order, hand[k])
		if v < 0 || len(hand) != 5 {
			return 0, fmt.Errorf("invalid hand %q", hand)
		}
		count[v]++
		key = key*16 + v
	}
	wild := 0
	if jokers {
		wild, count[0] = count[0], 0
	}
	first, second := 0, 0
	for _, c := range count {
		if c > first {
			first, second = c, first
		} else if c > second {
			second = c
		}
	}
	first += wild
	// Types from weakest: high card, pair, two pair, three, full house, four, five.
	// Indexed by the two largest counts: (1,1) high card, (2,1) pair, (2,2)
	// two pair, (3,1) three, (3,2) full house, (4,1) four, (5,0) five.
	if first < 1 || first > 5 || second > 5-first {
		return 0, fmt.Errorf("hand %q has counts %d and %d", hand, first, second)
	}
	kind := [6][3]int{{}, {0, 0}, {1, 1, 2}, {3, 3, 4}, {5, 5}, {6}}[first][second]
	return kind<<20 | key, nil
}

// winnings sorts the hands by strength and sums bid times rank.
func winnings(hands []string, bids []int, jokers bool, keys []int, idx []int) (int, error) {
	if len(hands) != len(bids) || len(keys) < len(hands) || len(idx) < len(hands) {
		return 0, errors.New("hands, bids and scratch sizes differ")
	}
	for i, h := range hands {
		k, err := strength(h, jokers)
		if err != nil {
			return 0, err
		}
		keys[i], idx[i] = k, i
		for j := i; j > 0 && keys[idx[j-1]] > keys[idx[j]]; j-- {
			idx[j-1], idx[j] = idx[j], idx[j-1]
		}
	}
	if len(hands) == 0 {
		return 0, errors.New("no hands")
	}
	total := 0
	for rank, i := range idx[:len(hands)] {
		total += (rank + 1) * bids[i]
	}
	return total, nil
}

func day7(path string) error {
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
	if len(lines) > maxHands {
		return fmt.Errorf("%s: more than %d hands", path, maxHands)
	}
	hands, bids := make([]string, len(lines)), make([]int, len(lines))
	for i, l := range lines {
		h, b, ok := strings.Cut(l, " ")
		bid, err := strconv.Atoi(b)
		if !ok || err != nil || bid < 0 {
			return fmt.Errorf("%s: malformed hand %q", path, l)
		}
		hands[i], bids[i] = h, bid
	}
	keys, idx := make([]int, len(lines)), make([]int, len(lines))
	plain, err := winnings(hands, bids, false, keys, idx)
	if err != nil {
		return err
	}
	wild, err := winnings(hands, bids, true, keys, idx)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", plain, wild); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
