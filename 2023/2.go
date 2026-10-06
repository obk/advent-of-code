// Day 2: Cube Conundrum
//
// Each line of the input is a game: handfuls of red, green and blue cubes
// shown from a bag, "Game 1: 3 blue, 4 red; 1 red, 2 green".
//
// Part 1 sums the IDs of games possible with 12 red, 13 green and 14 blue.
// Part 2 sums, over games, the product of the fewest cubes of each colour
// that make the game possible.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 16
	maxGames      = 1_000
)

// fewest returns a game's ID and the most cubes of each colour shown.
func fewest(line string) (int, [3]int, error) {
	var most [3]int
	head, body, found := strings.Cut(line, ": ")
	id, err := strconv.Atoi(strings.TrimPrefix(head, "Game "))
	if !found || err != nil {
		return 0, most, fmt.Errorf("malformed game %q", line)
	}
	for _, shown := range strings.Split(strings.NewReplacer(";", ",").Replace(body), ",") {
		n, colour, ok := strings.Cut(strings.TrimSpace(shown), " ")
		v, err := strconv.Atoi(n)
		if !ok || err != nil || v < 0 {
			return 0, most, fmt.Errorf("malformed handful %q", shown)
		}
		c := -1
		switch colour {
		case "red":
			c = 0
		case "green":
			c = 1
		case "blue":
			c = 2
		}
		if c < 0 {
			return 0, most, fmt.Errorf("unknown colour %q", colour)
		}
		most[c] = max(most[c], v)
	}
	return id, most, nil
}

func day2(path string) error {
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
	if len(lines) > maxGames {
		return fmt.Errorf("%s: more than %d games", path, maxGames)
	}
	possible, power := 0, 0
	for _, l := range lines {
		id, most, err := fewest(l)
		if err != nil {
			return err
		}
		if most[0] <= 12 && most[1] <= 13 && most[2] <= 14 {
			possible += id
		}
		power += most[0] * most[1] * most[2]
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", possible, power); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
