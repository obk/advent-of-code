// Day 3: Spiral Memory
//
// Squares are numbered 1, 2, 3, ... in a spiral out from the centre:
//
//	17  16  15  14  13
//	18   5   4   3  12
//	19   6   1   2  11
//	20   7   8   9  10
//	21  22  23---> ...
//
// The input is a square's number.
//
// Part 1 is its taxicab distance from square 1.
// Part 2 fills the spiral again, each square getting the sum of its already
// filled neighbours (diagonals included), square 1 starting at 1, and asks
// for the first value written that is larger than the input.
//
// Ring k of the spiral (k = 1, 2, ...) ends with square (2k+1)^2 and has four
// sides of 2k squares. A square's distance is k plus how far it sits from
// the middle of its side, so no walk is needed.
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
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 10
	maxSquare     = 1_000_000_000_000
)

// readSquare returns the square's number.
func readSquare(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid number: %w", path, err)
	}
	if n < 1 || n > maxSquare {
		return 0, fmt.Errorf("%s: square %d is outside 1..%d", path, n, maxSquare)
	}
	return n, nil
}

// distance returns how far square n is from square 1.
func distance(n int) (int, error) {
	if n < 1 || n > maxSquare {
		return 0, fmt.Errorf("square %d is outside 1..%d", n, maxSquare)
	}
	if n == 1 {
		return 0, nil
	}
	k := 1
	for i := 0; i < 1_000_000 && (2*k+1)*(2*k+1) < n; i++ {
		k++
	}
	if (2*k+1)*(2*k+1) < n {
		return 0, fmt.Errorf("square %d is beyond ring %d", n, k)
	}
	if (2*k-1)*(2*k-1) >= n {
		return 0, fmt.Errorf("square %d belongs to an inner ring, not ring %d", n, k)
	}
	// Squares of ring k run from (2k-1)^2+1; each side's middle is k-1
	// squares into the side.
	along := (n - (2*k-1)*(2*k-1) - 1) % (2 * k)
	off := along - (k - 1)
	if off < -k || off > k {
		return 0, fmt.Errorf("square %d is %d from its side's middle", n, off)
	}
	return k + max(off, -off), nil
}

// firstSumAbove walks the spiral on a small fixed grid, writing neighbour
// sums, until one exceeds limit. The sums at least double each ring, so a
// grid of side gridSide is far more than enough for any int limit.
func firstSumAbove(limit int) (int, error) {
	if limit < 1 || limit > maxSquare {
		return 0, fmt.Errorf("limit %d is outside 1..%d", limit, maxSquare)
	}
	const gridSide = 41
	var grid [gridSide][gridSide]int
	x, y := gridSide/2, gridSide/2
	grid[y][x] = 1
	dx, dy := 1, 0 // start by heading right, then turn left whenever possible
	for range gridSide * gridSide {
		x, y = x+dx, y+dy
		if x < 1 || y < 1 || x >= gridSide-1 || y >= gridSide-1 {
			return 0, errors.New("the spiral left the grid")
		}
		sum := 0
		for ny := y - 1; ny <= y+1; ny++ {
			for nx := x - 1; nx <= x+1; nx++ {
				sum += grid[ny][nx]
			}
		}
		grid[y][x] = sum
		if sum > limit {
			return sum, nil
		}
		// Turning left is (dx, dy) -> (dy, -dx); turn if that square is free.
		if grid[y-dx][x+dy] == 0 {
			dx, dy = dy, -dx
		}
	}
	return 0, errors.New("the spiral filled the grid")
}

func day3(path string) error {
	n, err := readSquare(path)
	if err != nil {
		return err
	}
	steps, err := distance(n)
	if err != nil {
		return err
	}
	sum, err := firstSumAbove(n)
	if err != nil {
		return err
	}
	if sum <= n {
		return fmt.Errorf("value %d written is not larger than %d", sum, n)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, sum); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
