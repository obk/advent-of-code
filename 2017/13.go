// Day 13: Packet Scanners
//
// Each line of the input is a firewall layer, "depth: range". Each layer's
// scanner sweeps down and back up its range, one step a picosecond, and a
// packet moving one layer a picosecond is caught at a layer if that layer's
// scanner is at the top just as the packet arrives.
//
// Part 1 is the trip's severity, the sum of depth*range over the layers where
// the packet is caught when it leaves at once.
// Part 2 is the smallest delay before setting off that gets the packet
// through without being caught at all.
//
// A scanner of range r is back at the top every 2(r-1) picoseconds, and the
// packet reaches layer d at time d, so it is caught there exactly when d is
// a multiple of 2(r-1).
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
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxValue      = 10_000
	maxDelay      = 100_000_000
)

// layer is one firewall layer.
type layer struct {
	depth, span int
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

func parseInput(path string) ([]layer, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	layers := make([]layer, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		d, r, found := strings.Cut(line, ": ")
		if !found {
			return nil, fmt.Errorf("malformed layer %q", line)
		}
		depth, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return nil, fmt.Errorf("invalid depth in %q: %w", line, err)
		}
		span, err := strconv.Atoi(strings.TrimSpace(r))
		if err != nil {
			return nil, fmt.Errorf("invalid range in %q: %w", line, err)
		}
		if depth < 0 || depth > maxValue || span < 1 || span > maxValue {
			return nil, fmt.Errorf("layer %q is out of range", line)
		}
		layers = append(layers, layer{depth, span})
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("%s: no layers", path)
	}
	return layers, nil
}

// caught reports whether a packet sent at time delay is caught by l.
func caught(l layer, delay int) (bool, error) {
	if l.span < 1 || delay < 0 {
		return false, fmt.Errorf("layer %+v or delay %d is invalid", l, delay)
	}
	if l.depth < 0 {
		return false, fmt.Errorf("layer %+v has a negative depth", l)
	}
	period := 2 * (l.span - 1)
	if period == 0 {
		return true, nil // a scanner of range 1 never leaves the top
	}
	return (l.depth+delay)%period == 0, nil
}

// severity sums depth*range over the layers that catch a packet sent at once.
func severity(layers []layer) (int, error) {
	if len(layers) == 0 {
		return 0, errors.New("no layers")
	}
	if len(layers) > maxLines {
		return 0, fmt.Errorf("%d layers, more than %d", len(layers), maxLines)
	}
	total := 0
	for _, l := range layers {
		c, err := caught(l, 0)
		if err != nil {
			return 0, err
		}
		if c {
			total += l.depth * l.span
		}
	}
	return total, nil
}

// safeDelay returns the smallest delay at which no layer catches the packet.
func safeDelay(layers []layer) (int, error) {
	if len(layers) == 0 {
		return 0, errors.New("no layers")
	}
	if len(layers) > maxLines {
		return 0, fmt.Errorf("%d layers, more than %d", len(layers), maxLines)
	}
	for delay := range maxDelay {
		safe := true
		for _, l := range layers {
			c, err := caught(l, delay)
			if err != nil {
				return 0, err
			}
			if c {
				safe = false
				break
			}
		}
		if safe {
			return delay, nil
		}
	}
	return 0, fmt.Errorf("no safe delay below %d", maxDelay)
}

func day13(path string) error {
	layers, err := parseInput(path)
	if err != nil {
		return err
	}
	total, err := severity(layers)
	if err != nil {
		return err
	}

	delay, err := safeDelay(layers)
	if err != nil {
		return err
	}
	if total < 0 {
		return fmt.Errorf("severity %d is negative", total)
	}
	if delay == 0 && total > 0 {
		return errors.New("leaving at once both is and is not caught")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, delay); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
