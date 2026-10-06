// Day 5: If You Give A Seed A Fertilizer
//
// The input is a list of seeds, then seven maps, each lines of "dest src
// len": numbers src..src+len-1 map to dest..; other numbers map to
// themselves.
//
// Part 1 maps every seed through all seven maps and gives the lowest
// location.
// Part 2 reads the seeds as pairs (start, length) of ranges, and gives the
// lowest location of any seed in them.
//
// Part 2 maps whole ranges: each range is split at the rules' edges, the
// pieces inside a rule shift, and the rest pass through. Ranges live in
// two fixed buffers, swapped between maps.
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
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 14
	maxRules      = 64
	maxRanges     = 4_096
	maps          = 7
)

// rule maps src..src+n-1 to dst..
type rule struct {
	dst, src, n int
}

// almanac is the seeds and the maps.
type almanac struct {
	seeds []int
	rules [maps][maxRules]rule
	count [maps]int
}

// load parses the almanac.
func load(path string) (*almanac, error) {
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
	blocks := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(blocks) != maps+1 || !strings.HasPrefix(blocks[0], "seeds: ") {
		return nil, fmt.Errorf("%s: expected seeds and %d maps", path, maps)
	}
	f := strings.Fields(strings.TrimPrefix(blocks[0], "seeds: "))
	a := &almanac{seeds: make([]int, len(f))}
	for i, s := range f {
		if a.seeds[i], err = strconv.Atoi(s); err != nil || a.seeds[i] < 0 {
			return nil, fmt.Errorf("%s: invalid seed %q", path, s)
		}
	}
	for m, b := range blocks[1:] {
		lines := strings.Split(b, "\n")[1:]
		if len(lines) > maxRules {
			return nil, fmt.Errorf("%s: map %d has more than %d rules", path, m, maxRules)
		}
		for i, l := range lines {
			r := &a.rules[m][i]
			if n, err := fmt.Sscanf(l, "%d %d %d", &r.dst, &r.src, &r.n); err != nil || n != 3 || r.n < 1 {
				return nil, fmt.Errorf("%s: invalid rule %q", path, l)
			}
		}
		a.count[m] = len(lines)
	}
	return a, nil
}

// mapRanges sends the ranges in from through map m into out.
func (a *almanac) mapRanges(m int, from, out [][2]int) ([][2]int, error) {
	if m < 0 || m >= maps {
		return nil, fmt.Errorf("map %d does not exist", m)
	}
	out = out[:0]
	for _, r := range from {
		lo, hi := r[0], r[1] // [lo, hi)
		for k := 0; k <= a.count[m] && lo < hi; k++ {
			// The rule covering lo, or the nearest rule starting after it.
			cover, next := -1, hi
			for i, ru := range a.rules[m][:a.count[m]] {
				if ru.src <= lo && lo < ru.src+ru.n {
					cover = i
				} else if ru.src > lo {
					next = min(next, ru.src)
				}
			}
			if len(out) == cap(out) {
				return nil, errors.New("too many ranges")
			}
			if cover < 0 { // unmapped up to the next rule
				out = append(out, [2]int{lo, next})
				lo = next
				continue
			}
			ru := a.rules[m][cover]
			end := min(hi, ru.src+ru.n)
			out = append(out, [2]int{lo - ru.src + ru.dst, end - ru.src + ru.dst})
			lo = end
		}
	}
	return out, nil
}

// lowest maps ranges through every map and returns the lowest start.
func (a *almanac) lowest(ranges, spare [][2]int) (int, error) {
	if len(ranges) == 0 {
		return 0, errors.New("no seeds")
	}
	var err error
	for m := range maps {
		if spare, err = a.mapRanges(m, ranges, spare); err != nil {
			return 0, err
		}
		ranges, spare = spare, ranges
	}
	best := math.MaxInt
	for _, r := range ranges {
		best = min(best, r[0])
	}
	return best, nil
}

func day5(path string) error {
	a, err := load(path)
	if err != nil {
		return err
	}
	if len(a.seeds)%2 != 0 {
		return fmt.Errorf("%s: an odd number of seeds cannot be ranges", path)
	}
	ranges, spare := make([][2]int, 0, maxRanges), make([][2]int, 0, maxRanges)
	for _, s := range a.seeds {
		ranges = append(ranges, [2]int{s, s + 1})
	}
	single, err := a.lowest(ranges, spare)
	if err != nil {
		return err
	}
	ranges = ranges[:0]
	for i := 0; i < len(a.seeds); i += 2 {
		ranges = append(ranges, [2]int{a.seeds[i], a.seeds[i] + a.seeds[i+1]})
	}
	ranged, err := a.lowest(ranges, spare)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", single, ranged); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
