// Day 7: The Sum of Its Parts
//
// Each line of the input says one step must finish before another can begin:
// "Step C must be finished before step A can begin." Steps are letters.
//
// Part 1 is the order the steps are done in, always taking the
// alphabetically first step whose prerequisites are all done.
// Part 2 has five workers and steps taking 60 seconds plus their letter's
// number (A=1); it asks how long everything takes.
//
// Prerequisites are bitmasks over the 26 letters, so a step is ready when its
// mask is a subset of the done mask.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	letters       = 26
	workers       = 5
	baseTime      = 60
	maxSeconds    = letters * (baseTime + letters)
)

// steps holds each step's prerequisites and which steps appear at all.
type steps struct {
	needs   [letters]uint32
	present uint32
}

// readSteps parses the requirements.
func readSteps(path string) (steps, error) {
	if path == "" {
		return steps{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return steps{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return steps{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxLines {
		return steps{}, fmt.Errorf("%s: %d lines, more than %d", path, len(lines), maxLines)
	}
	var s steps
	for _, line := range lines {
		w := strings.Fields(line)
		if len(w) != 10 || len(w[1]) != 1 || len(w[7]) != 1 {
			return steps{}, fmt.Errorf("%s: malformed line %q", path, line)
		}
		before, after := int(w[1][0]-'A'), int(w[7][0]-'A')
		if before < 0 || before >= letters || after < 0 || after >= letters || before == after {
			return steps{}, fmt.Errorf("%s: invalid steps in %q", path, line)
		}
		s.needs[after] |= 1 << before
		s.present |= 1<<before | 1<<after
	}
	return s, nil
}

// order returns the steps in the order they are done, into out.
func (s *steps) order(out []byte) ([]byte, error) {
	if s.present == 0 {
		return nil, errors.New("no steps")
	}
	out = out[:0]
	var done uint32
	for range letters {
		next := -1
		for i := range letters {
			if s.present>>i&1 == 1 && done>>i&1 == 0 && s.needs[i]&^done == 0 {
				next = i
				break
			}
		}
		if next < 0 {
			break
		}
		done |= 1 << next
		out = append(out, byte('A'+next))
	}
	if done != s.present {
		return nil, errors.New("the requirements go round in a circle")
	}
	if len(out) > letters {
		return nil, fmt.Errorf("ordered %d steps, more than %d letters", len(out), letters)
	}
	return out, nil
}

// timed simulates the workers second by second and returns the total time.
func (s *steps) timed(crew, base int) (int, error) {
	if crew < 1 || crew > letters || base < 0 {
		return 0, fmt.Errorf("invalid crew %d or base time %d", crew, base)
	}
	var job, left [letters]int // per worker: step (-1 idle) and seconds left
	for w := range crew {
		job[w] = -1
	}
	var done, started uint32
	for t := 0; t <= maxSeconds; t++ {
		for w := range crew {
			if job[w] >= 0 && left[w] == 0 {
				done, job[w] = done|1<<job[w], -1
			}
		}
		if done == s.present {
			return t, nil
		}
		for w := range crew {
			for i := 0; i < letters && job[w] < 0; i++ {
				free := s.present>>i&1 == 1 && started>>i&1 == 0 && s.needs[i]&^done == 0
				if free {
					job[w], left[w], started = i, base+i+1, started|1<<i
				}
			}
			if job[w] >= 0 {
				left[w]--
			}
		}
	}
	return 0, fmt.Errorf("the steps took longer than %d seconds", maxSeconds)
}

func day7(path string) error {
	s, err := readSteps(path)
	if err != nil {
		return err
	}
	order, err := s.order(make([]byte, 0, letters))
	if err != nil {
		return err
	}

	seconds, err := s.timed(workers, baseTime)
	if err != nil {
		return err
	}
	if seconds < len(order) {
		return fmt.Errorf("%d steps cannot all finish in %d seconds", len(order), seconds)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", order, seconds); err != nil {
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
