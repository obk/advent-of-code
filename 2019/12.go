// Day 12: The N-Body Problem
//
// The input is the positions of four moons, "<x=-1, y=0, z=2>". Each step,
// every pair of moons pulls each other's velocity one unit closer on every
// axis, then every moon moves by its velocity.
//
// Part 1 is the total energy after 1000 steps: for each moon, the sum of
// its absolute position values times the sum of its absolute velocities.
// Part 2 is the number of steps until the moons first repeat a state.
//
// The axes move independently, and each step can be undone, so each axis
// returns to its starting state after its own period. The whole system
// repeats after the least common multiple of the three periods.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 10
	moons         = 4
	maxCoord      = 1_000_000
	maxPeriod     = 10_000_000
)

// system is every moon's position and velocity, axis by axis.
type system struct {
	pos, vel [3][moons]int
}

// load parses the four positions.
func load(path string) (system, error) {
	if path == "" {
		return system{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return system{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return system{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != moons {
		return system{}, fmt.Errorf("%s: %d moons, expected %d", path, len(lines), moons)
	}
	var s system
	clean := strings.NewReplacer("<", "", ">", "", "x=", "", "y=", "", "z=", "", ",", " ")
	for i, line := range lines {
		f := strings.Fields(clean.Replace(line))
		if len(f) != 3 {
			return system{}, fmt.Errorf("%s: malformed moon %q", path, line)
		}
		for axis, text := range f {
			v, err := strconv.Atoi(text)
			if err != nil || v < -maxCoord || v > maxCoord {
				return system{}, fmt.Errorf("%s: invalid coordinate in %q", path, line)
			}
			s.pos[axis][i] = v
		}
	}
	return s, nil
}

// stepAxis advances one axis by one step.
func (s *system) stepAxis(axis int) error {
	if axis < 0 || axis > 2 {
		return fmt.Errorf("axis %d does not exist", axis)
	}
	p, v := &s.pos[axis], &s.vel[axis]
	for i := range moons {
		for j := range moons {
			switch {
			case p[j] > p[i]:
				v[i]++
			case p[j] < p[i]:
				v[i]--
			}
		}
	}
	momentum := 0
	for i := range moons {
		p[i] += v[i]
		momentum += v[i]
	}
	if momentum != 0 {
		return fmt.Errorf("axis %d gained momentum %d", axis, momentum)
	}
	return nil
}

// energy is the system's total energy.
func (s *system) energy() (int, error) {
	total := 0
	for i := range moons {
		pot, kin := 0, 0
		for axis := range 3 {
			pot += max(s.pos[axis][i], -s.pos[axis][i])
			kin += max(s.vel[axis][i], -s.vel[axis][i])
		}
		if pot > 6*maxCoord || kin > 6*maxCoord {
			return 0, fmt.Errorf("moon %d has run away", i)
		}
		total += pot * kin
	}
	if total < 0 {
		return 0, errors.New("total energy overflowed")
	}
	return total, nil
}

// simulate runs the given number of steps on every axis.
func (s *system) simulate(steps int) error {
	if steps < 0 || steps > maxPeriod {
		return fmt.Errorf("step count %d is outside 0..%d", steps, maxPeriod)
	}
	for range steps {
		for axis := range 3 {
			if err := s.stepAxis(axis); err != nil {
				return err
			}
		}
	}
	return nil
}

// period counts the steps until an axis is back where it began.
func (s *system) period(axis int) (int, error) {
	if axis < 0 || axis > 2 {
		return 0, fmt.Errorf("axis %d does not exist", axis)
	}
	start, startVel := s.pos[axis], s.vel[axis]
	for steps := 1; steps <= maxPeriod; steps++ {
		if err := s.stepAxis(axis); err != nil {
			return 0, err
		}
		if s.pos[axis] == start && s.vel[axis] == startVel {
			return steps, nil
		}
	}
	return 0, fmt.Errorf("axis %d did not repeat within %d steps", axis, maxPeriod)
}

// lcm is the least common multiple of two positive numbers, guarding
// against overflow.
func lcm(a, b int) (int, error) {
	if a < 1 || b < 1 {
		return 0, fmt.Errorf("lcm of %d and %d", a, b)
	}
	x, y := a, b
	for i := 0; i < 128 && y != 0; i++ {
		x, y = y, x%y
	}
	if a/x > math.MaxInt64/b {
		return 0, fmt.Errorf("lcm of %d and %d overflows", a, b)
	}
	result := a / x * b
	if result%a != 0 || result%b != 0 {
		return 0, fmt.Errorf("%d is not a multiple of %d and %d", result, a, b)
	}
	return result, nil
}

func day12(path string) error {
	s, err := load(path)
	if err != nil {
		return err
	}
	if s.vel != [3][moons]int{} {
		return errors.New("moons must start at rest")
	}
	initial := s
	if err := s.simulate(1000); err != nil {
		return err
	}
	energy, err := s.energy()
	if err != nil {
		return err
	}
	repeat := 1
	for axis := range 3 {
		p, err := initial.period(axis)
		if err != nil {
			return err
		}
		if repeat, err = lcm(repeat, p); err != nil {
			return err
		}
	}
	if repeat < 1 {
		return fmt.Errorf("the moons repeat after %d steps", repeat)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", energy, repeat); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
