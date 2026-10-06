// Day 9: All in a Single Night
//
// Each line of the input gives the distance between two locations, as
// "London to Dublin = 464". Santa must visit every location exactly once,
// starting and ending anywhere.
//
// Part 1 is the shortest such route, part 2 the longest.
//
// With only a handful of locations this is a small travelling salesman
// problem, solved exactly with the Held-Karp dynamic programme: best[set][end]
// is the shortest (or longest) route that visits exactly the locations in set
// and finishes at end. Sets are bitmasks, so processing them in increasing
// order means every smaller set is done before it is needed.
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
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxPlaces     = 12
	maxDistance   = 1_000_000
)

// network is the distance table between every pair of locations.
type network struct {
	n    int
	dist [maxPlaces][maxPlaces]int
}

// tables holds the Held-Karp tables for both parts.
type tables struct {
	shortest [1 << maxPlaces][maxPlaces]int
	longest  [1 << maxPlaces][maxPlaces]int
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

// placeID returns the number of the named location, numbering it if new.
func placeID(ids map[string]int, name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty location name")
	}
	if id, ok := ids[name]; ok {
		return id, nil
	}
	if len(ids) >= maxPlaces {
		return 0, fmt.Errorf("more than %d locations", maxPlaces)
	}
	ids[name] = len(ids)
	return ids[name], nil
}

func parseInput(path string) (network, error) {
	lines, err := readLines(path)
	if err != nil {
		return network{}, err
	}

	var net network
	ids := make(map[string]int, maxPlaces)
	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		if len(words) != 5 || words[1] != "to" || words[3] != "=" {
			return network{}, fmt.Errorf("malformed line %q", line)
		}
		a, err := placeID(ids, words[0])
		if err != nil {
			return network{}, err
		}
		b, err := placeID(ids, words[2])
		if err != nil {
			return network{}, err
		}
		d, err := strconv.Atoi(words[4])
		if err != nil {
			return network{}, fmt.Errorf("invalid distance in %q: %w", line, err)
		}
		if d < 0 || d > maxDistance || a == b {
			return network{}, fmt.Errorf("invalid route %q", line)
		}
		if net.dist[a][b] != 0 && net.dist[a][b] != d {
			return network{}, fmt.Errorf("two distances given for %q", line)
		}
		net.dist[a][b], net.dist[b][a] = d, d
	}

	net.n = len(ids)
	if net.n < 2 {
		return network{}, fmt.Errorf("%s: need at least two locations, found %d", path, net.n)
	}
	return net, nil
}

// routes fills the Held-Karp tables and returns the shortest and longest
// routes through every location. A route of n places has n-1 legs, so -1
// marks a table entry no route reaches.
func (t *tables) routes(net network) (shortest, longest int, err error) {
	if net.n < 2 || net.n > maxPlaces {
		return 0, 0, fmt.Errorf("%d locations is outside 2..%d", net.n, maxPlaces)
	}

	all := 1<<net.n - 1
	for set := 1; set <= all; set++ {
		for end := range net.n {
			t.shortest[set][end], t.longest[set][end] = -1, -1
			if set>>end&1 == 0 {
				continue
			}
			rest := set &^ (1 << end)
			if rest == 0 { // a route of just this location
				t.shortest[set][end], t.longest[set][end] = 0, 0
				continue
			}
			for prev := range net.n {
				if rest>>prev&1 == 0 || t.shortest[rest][prev] < 0 {
					continue
				}
				short := t.shortest[rest][prev] + net.dist[prev][end]
				long := t.longest[rest][prev] + net.dist[prev][end]
				if t.shortest[set][end] < 0 || short < t.shortest[set][end] {
					t.shortest[set][end] = short
				}
				t.longest[set][end] = max(t.longest[set][end], long)
			}
		}
	}

	shortest, longest = -1, -1
	for end := range net.n {
		if shortest < 0 || t.shortest[all][end] < shortest {
			shortest = t.shortest[all][end]
		}
		longest = max(longest, t.longest[all][end])
	}
	if shortest < 0 || longest < shortest {
		return 0, 0, fmt.Errorf("no route found (shortest %d, longest %d)", shortest, longest)
	}
	return shortest, longest, nil
}

func day9(path string) error {
	net, err := parseInput(path)
	if err != nil {
		return err
	}
	shortest, longest, err := new(tables).routes(net)
	if err != nil {
		return err
	}
	if longest > (net.n-1)*maxDistance {
		return fmt.Errorf("route of %d is longer than %d legs can be", longest, net.n-1)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", shortest, longest); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
