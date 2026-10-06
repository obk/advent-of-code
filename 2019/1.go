// Day 1: The Tyranny of the Rocket Equation
//
// Each line of the input is a module's mass. A module needs fuel equal to
// its mass divided by three, rounded down, minus two.
//
// Part 1 is the total fuel for all modules.
// Part 2 also fuels the fuel: each amount of fuel needs fuel by the same
// rule, until the extra comes to nothing.
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
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
	maxModules    = 10_000
	maxMass       = 1_000_000_000
)

// readMasses parses one mass per line.
func readMasses(path string) ([]int, error) {
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
	fields := strings.Fields(string(content))
	if len(fields) == 0 || len(fields) > maxModules {
		return nil, fmt.Errorf("%s: %d modules, expected 1..%d", path, len(fields), maxModules)
	}
	masses := make([]int, len(fields))
	for i, f := range fields {
		if masses[i], err = strconv.Atoi(f); err != nil {
			return nil, fmt.Errorf("%s: invalid mass %q: %w", path, f, err)
		}
		if masses[i] < 0 || masses[i] > maxMass {
			return nil, fmt.Errorf("%s: mass %d is outside 0..%d", path, masses[i], maxMass)
		}
	}
	return masses, nil
}

// fuel is the fuel a mass needs, never below zero.
func fuel(mass int) (int, error) {
	if mass < 0 || mass > maxMass {
		return 0, fmt.Errorf("mass %d is outside 0..%d", mass, maxMass)
	}
	f := max(mass/3-2, 0)
	if f > mass {
		return 0, fmt.Errorf("fuel %d exceeds mass %d", f, mass)
	}
	return f, nil
}

func day1(path string) error {
	masses, err := readMasses(path)
	if err != nil {
		return err
	}
	total, full := 0, 0
	for _, m := range masses {
		f, err := fuel(m)
		if err != nil {
			return err
		}
		total += f
		// Each round divides by three, so 64 rounds reach zero from any int.
		for extra, round := f, 0; extra > 0 && round < 64; round++ {
			full += extra
			if extra, err = fuel(extra); err != nil {
				return err
			}
		}
	}

	if 2*full > 3*total {
		return fmt.Errorf("fuel for fuel (%d) is more than half again the fuel (%d)", full, total)
	}
	if full < total {
		return fmt.Errorf("fuel for fuel (%d) came to less than the fuel alone (%d)", full, total)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, full); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
