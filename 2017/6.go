// Day 6: Memory Reallocation
//
// The input is the number of blocks in each of up to sixteen memory banks.
// Each cycle empties the fullest bank (the lowest-numbered one on a tie) and
// deals its blocks out one at a time to the following banks, wrapping round.
//
// Part 1 is how many cycles happen before a configuration repeats.
// Part 2 is the length of the loop the configurations then go round.
//
// The sequence of configurations is eventually periodic: mu cycles lead into
// a loop of length lambda, and the first repeat comes after mu + lambda
// cycles. Brent's cycle-finding algorithm finds both using only a couple of
// configurations, with no record of every configuration seen.
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
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 10
	maxBanks      = 16
	maxBlocks     = 1_000_000
	maxCycles     = 100_000_000
)

// banks is one configuration of the memory.
type banks struct {
	blocks [maxBanks]int
	n      int
}

// readBanks returns the starting configuration.
func readBanks(path string) (banks, error) {
	if path == "" {
		return banks{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return banks{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return banks{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	fields := strings.Fields(string(content))
	if len(fields) == 0 || len(fields) > maxBanks {
		return banks{}, fmt.Errorf("%s: %d banks, expected 1..%d", path, len(fields), maxBanks)
	}
	var b banks
	b.n = len(fields)
	for i, f := range fields {
		if b.blocks[i], err = strconv.Atoi(f); err != nil {
			return banks{}, fmt.Errorf("%s: invalid block count %q: %w", path, f, err)
		}
		if b.blocks[i] < 0 || b.blocks[i] > maxBlocks {
			return banks{}, fmt.Errorf("%s: block count %d is outside 0..%d", path, b.blocks[i], maxBlocks)
		}
	}
	return b, nil
}

// redistribute performs one cycle and returns the new configuration.
func redistribute(b banks) (banks, error) {
	if b.n < 1 || b.n > maxBanks {
		return banks{}, fmt.Errorf("%d banks is outside 1..%d", b.n, maxBanks)
	}
	fullest := 0
	for i := 1; i < b.n; i++ {
		if b.blocks[i] > b.blocks[fullest] {
			fullest = i
		}
	}
	blocks := b.blocks[fullest]
	b.blocks[fullest] = 0
	// Every bank gets blocks/n, and the first blocks%n after the fullest one
	// more each.
	for k := 1; k <= b.n; k++ {
		i := (fullest + k) % b.n
		b.blocks[i] += blocks / b.n
		if k <= blocks%b.n {
			b.blocks[i]++
		}
	}
	return b, nil
}

// findCycle returns the length of the lead-in (mu) and of the loop (lambda).
func findCycle(start banks) (mu, lambda int, err error) {
	if start.n < 1 || start.n > maxBanks {
		return 0, 0, fmt.Errorf("%d banks is outside 1..%d", start.n, maxBanks)
	}
	// Brent: grow a window in powers of two until the hare meets the tortoise.
	power, lambda := 1, 1
	tortoise := start
	hare, err := redistribute(start)
	if err != nil {
		return 0, 0, err
	}
	for steps := 0; tortoise != hare; steps++ {
		if steps > maxCycles {
			return 0, 0, fmt.Errorf("no repeat within %d cycles", maxCycles)
		}
		if power == lambda {
			tortoise, power, lambda = hare, power*2, 0
		}
		if hare, err = redistribute(hare); err != nil {
			return 0, 0, err
		}
		lambda++
	}

	// Start the hare lambda ahead, then step both until they meet: that is mu.
	tortoise, hare = start, start
	for range lambda {
		if hare, err = redistribute(hare); err != nil {
			return 0, 0, err
		}
	}
	for mu = 0; tortoise != hare; mu++ {
		if mu > maxCycles {
			return 0, 0, errors.New("lead-in did not end")
		}
		if tortoise, err = redistribute(tortoise); err != nil {
			return 0, 0, err
		}
		if hare, err = redistribute(hare); err != nil {
			return 0, 0, err
		}
	}
	return mu, lambda, nil
}

func day6(path string) error {
	start, err := readBanks(path)
	if err != nil {
		return err
	}
	mu, lambda, err := findCycle(start)
	if err != nil {
		return err
	}

	if lambda < 1 {
		return fmt.Errorf("loop length %d is not positive", lambda)
	}
	if mu < 0 {
		return fmt.Errorf("lead-in %d is negative", mu)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", mu+lambda, lambda); err != nil {
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
