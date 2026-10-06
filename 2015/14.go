// Day 14: Reindeer Olympics
//
// Each line of the input describes a reindeer: "Comet can fly 14 km/s for 10
// seconds, but then must rest for 127 seconds." Reindeer alternate between
// flying at top speed and resting, starting with a burst of flight.
//
// Part 1 is the distance the winner has covered after 2503 seconds.
// Part 2 awards a point every second to every reindeer in the lead at that
// moment, and asks for the winner's points.
//
// The race is simulated second by second; with a handful of reindeer and
// 2503 seconds that is a few tens of thousands of steps.
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
	maxInputBytes = 1 << 20
	maxLines      = 100
	maxValue      = 10_000 // for speeds and times
	raceSeconds   = 2503
)

// reindeer is one racer and its state during the race.
type reindeer struct {
	speed, flyTime, restTime int
	distance, points         int
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

// parseReindeer reads one description; the numbers are words 3, 6 and 13.
func parseReindeer(line string) (reindeer, error) {
	words := strings.Fields(line)
	if len(words) != 15 || words[1] != "can" || words[2] != "fly" {
		return reindeer{}, fmt.Errorf("malformed line %q", line)
	}
	if words[4] != "km/s" || words[14] != "seconds." {
		return reindeer{}, fmt.Errorf("unexpected units in %q", line)
	}

	var values [3]int
	for i, w := range [3]int{3, 6, 13} {
		v, err := strconv.Atoi(words[w])
		if err != nil {
			return reindeer{}, fmt.Errorf("invalid number in %q: %w", line, err)
		}
		if v < 1 || v > maxValue {
			return reindeer{}, fmt.Errorf("number in %q is outside 1..%d", line, maxValue)
		}
		values[i] = v
	}
	return reindeer{speed: values[0], flyTime: values[1], restTime: values[2]}, nil
}

func parseInput(path string) ([]reindeer, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	herd := make([]reindeer, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		r, err := parseReindeer(line)
		if err != nil {
			return nil, err
		}
		herd = append(herd, r)
	}
	if len(herd) == 0 {
		return nil, fmt.Errorf("%s: no reindeer found", path)
	}
	return herd, nil
}

// race runs the herd for the given number of seconds and returns the leader's
// distance and the most points any reindeer scored.
func race(herd []reindeer, seconds int) (distance, points int, err error) {
	if len(herd) == 0 {
		return 0, 0, errors.New("no reindeer to race")
	}
	if seconds < 1 || seconds > raceSeconds {
		return 0, 0, fmt.Errorf("race of %d seconds is outside 1..%d", seconds, raceSeconds)
	}

	for _, r := range herd {
		if r.speed < 1 || r.flyTime < 1 || r.restTime < 0 {
			return 0, 0, fmt.Errorf("reindeer %+v cannot race", r)
		}
	}
	for t := range seconds {
		lead := 0
		for i := range herd {
			r := &herd[i]
			if t%(r.flyTime+r.restTime) < r.flyTime {
				r.distance += r.speed
			}
			lead = max(lead, r.distance)
		}
		for i := range herd {
			if herd[i].distance == lead {
				herd[i].points++
			}
		}
	}

	for _, r := range herd {
		distance = max(distance, r.distance)
		points = max(points, r.points)
	}
	if distance < 1 {
		return 0, 0, errors.New("no reindeer moved")
	}
	if points < 1 || points > seconds {
		return 0, 0, fmt.Errorf("winner scored %d points in %d seconds", points, seconds)
	}
	return distance, points, nil
}

func day14(path string) error {
	herd, err := parseInput(path)
	if err != nil {
		return err
	}
	distance, points, err := race(herd, raceSeconds)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", distance, points); err != nil {
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
