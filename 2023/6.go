// Day 6: Wait For It
//
// The input is boat races: their times and record distances. Holding the
// button for h of a race's t milliseconds travels h*(t-h).
//
// Part 1 multiplies, over races, the number of hold times that beat the
// record.
// Part 2 reads each line as one number (ignoring spaces) for a single race.
//
// h*(t-h) rises to t/2 and falls symmetrically, so the winning holds run
// from the first winning h (found by binary search below t/2) to t minus
// it.
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
	maxInputBytes = 1 << 8
	maxRaces      = 8
)

// ways counts the winning hold times of a race.
func ways(t, record int) (int, error) {
	if t < 1 || record < 0 || t > 1<<31 {
		return 0, fmt.Errorf("race of %d ms with record %d is out of range", t, record)
	}
	if (t/2)*(t-t/2) <= record {
		return 0, nil
	}
	lo, hi := 0, t/2 // lo loses, hi wins
	for k := 0; k < 64 && hi-lo > 1; k++ {
		mid := (lo + hi) / 2
		if mid*(t-mid) > record {
			hi = mid
		} else {
			lo = mid
		}
	}
	return t - 2*hi + 1, nil
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		return fmt.Errorf("%s: expected a time line and a distance line", path)
	}
	times, dists := strings.Fields(lines[0])[1:], strings.Fields(lines[1])[1:]
	if len(times) != len(dists) || len(times) > maxRaces {
		return fmt.Errorf("%s: times and distances do not pair up", path)
	}
	product := 1
	for i := range times {
		t, errT := strconv.Atoi(times[i])
		d, errD := strconv.Atoi(dists[i])
		if errT != nil || errD != nil {
			return fmt.Errorf("%s: invalid race %q %q", path, times[i], dists[i])
		}
		w, err := ways(t, d)
		if err != nil {
			return err
		}
		product *= w
	}
	t, errT := strconv.Atoi(strings.Join(times, ""))
	d, errD := strconv.Atoi(strings.Join(dists, ""))
	if errT != nil || errD != nil {
		return fmt.Errorf("%s: the joined race is not a number", path)
	}
	long, err := ways(t, d)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", product, long); err != nil {
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
