// Day 10: Balance Bots
//
// The input says which bot each starting chip goes to ("value 5 goes to bot
// 2") and, for every bot, where it sends its lower and higher chip once it
// holds two ("bot 2 gives low to bot 1 and high to output 0").
//
// Part 1 is the number of the bot that compares the chips 61 and 17.
// Part 2 is the product of the chips that end up in outputs 0, 1 and 2.
//
// Bots that hold two chips wait in a queue. Each one handed out empties a bot
// for good (every bot's two chips arrive exactly once), so the queue sees each
// bot at most once and is a fixed-size array.
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
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxBots       = 1_000
	maxValue      = 1_000_000
)

// target is where a bot sends a chip: another bot or an output bin.
type target struct {
	output bool
	id     int
}

// bot is one robot's rules and the chips it holds.
type bot struct {
	low, high target
	hasRules  bool
	chips     [2]int
	held      int
}

// factory holds every bot and output bin.
type factory struct {
	bots    [maxBots]bot
	outputs [maxBots]int
	queue   [maxBots]int
}

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// number parses a non-negative number below limit.
func number(text string, limit int) (int, error) {
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", text, err)
	}
	if n < 0 || n >= limit {
		return 0, fmt.Errorf("number %d is outside 0..%d", n, limit-1)
	}
	return n, nil
}

// give hands a chip to a bot; it reports whether the bot now holds two.
func (f *factory) give(id, chip int) (bool, error) {
	if id < 0 || id >= maxBots {
		return false, fmt.Errorf("bot %d does not exist", id)
	}
	b := &f.bots[id]
	if b.held >= 2 {
		return false, fmt.Errorf("bot %d already holds two chips", id)
	}
	b.chips[b.held] = chip
	b.held++
	return b.held == 2, nil
}

// load reads every line into the factory, queueing bots that start with two
// chips. It returns how many are queued.
func (f *factory) load(lines []string) (int, error) {
	if len(lines) == 0 {
		return 0, errors.New("no instructions")
	}
	queued := 0
	for _, line := range lines {
		w := strings.Fields(line)
		switch {
		case len(w) == 0:
		case len(w) == 6 && w[0] == "value":
			chip, err := number(w[1], maxValue)
			if err != nil {
				return 0, err
			}
			id, err := number(w[5], maxBots)
			if err != nil {
				return 0, err
			}
			full, err := f.give(id, chip)
			if err != nil {
				return 0, err
			}
			if full {
				f.queue[queued], queued = id, queued+1
			}
		case len(w) == 12 && w[0] == "bot":
			id, err := number(w[1], maxBots)
			if err != nil {
				return 0, err
			}
			low, err := number(w[6], maxBots)
			if err != nil {
				return 0, err
			}
			high, err := number(w[11], maxBots)
			if err != nil {
				return 0, err
			}
			f.bots[id].low = target{output: w[5] == "output", id: low}
			f.bots[id].high = target{output: w[10] == "output", id: high}
			f.bots[id].hasRules = true
		default:
			return 0, fmt.Errorf("malformed line %q", line)
		}
	}
	return queued, nil
}

// run hands chips along until every bot is done, returning the bot that
// compared the two given chips (-1 if none did).
func (f *factory) run(queued, chipA, chipB int) (int, error) {
	if queued < 0 || queued > maxBots {
		return 0, fmt.Errorf("%d bots queued", queued)
	}
	if chipA == chipB {
		return 0, fmt.Errorf("cannot look for a comparison of chip %d with itself", chipA)
	}

	comparer := -1
	for head := 0; head < queued; head++ {
		id := f.queue[head]
		b := &f.bots[id]
		if b.held != 2 || !b.hasRules {
			return 0, fmt.Errorf("bot %d has %d chips and rules=%v", id, b.held, b.hasRules)
		}
		lo, hi := min(b.chips[0], b.chips[1]), max(b.chips[0], b.chips[1])
		if lo == min(chipA, chipB) && hi == max(chipA, chipB) {
			comparer = id
		}
		b.held = 0
		for _, send := range [2]struct {
			to   target
			chip int
		}{{b.low, lo}, {b.high, hi}} {
			if send.to.output {
				// Chip values start at 1 here, so 0 marks an empty bin.
				f.outputs[send.to.id] = send.chip + 1
				continue
			}
			full, err := f.give(send.to.id, send.chip)
			if err != nil {
				return 0, err
			}
			if full {
				if queued >= maxBots {
					return 0, errors.New("bot queue overflowed")
				}
				f.queue[queued], queued = send.to.id, queued+1
			}
		}
	}
	return comparer, nil
}

func day10(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	f := new(factory)
	queued, err := f.load(lines)
	if err != nil {
		return err
	}
	comparer, err := f.run(queued, 61, 17)
	if err != nil {
		return err
	}

	if comparer < 0 {
		return errors.New("part 1: no bot compared chips 61 and 17")
	}
	// Bins hold chip+1, so that 0 means empty.
	if f.outputs[0] == 0 || f.outputs[1] == 0 || f.outputs[2] == 0 {
		return errors.New("part 2: outputs 0, 1 and 2 are not all filled")
	}
	product := (f.outputs[0] - 1) * (f.outputs[1] - 1) * (f.outputs[2] - 1)
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", comparer, product); err != nil {
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
