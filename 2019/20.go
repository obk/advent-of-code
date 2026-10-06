// Day 20: Donut Maze
//
// The input is a donut-shaped maze of walls '#' and open tiles '.', with
// two-letter labels beside some open tiles on its outer and inner edges.
// Tiles with the same label are portals to each other, one step apart. The
// walk starts at the tile labelled AA and ends at ZZ.
//
// Part 1 is the fewest steps from AA to ZZ.
// Part 2 makes the maze recursive: inner portals lead into a copy of the
// maze one level deeper, outer portals back out a level, and AA and ZZ only
// exist at the outermost level. The answer is again the fewest steps.
//
// A breadth-first search over tiles, where a portal tile also neighbours
// its partner. In part 2 a state is a tile and a level; the levels are
// capped at maxLevel.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	cells         = maxSide * maxSide
	labels        = 26 * 26
	maxLevel      = 128
)

// maze is the map, its portals and the search's buffers.
type maze struct {
	grid       [][]byte
	w, h       int
	portal     [cells]int // partner tile, or -1
	outer      [cells]bool
	first      [labels]int // first tile seen with each label, or -1
	start, end int
	dist       []int32
	queue      []int32
}

// load reads the maze, keeping its spaces.
func load(path string) (*maze, error) {
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
	rows := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	mz := &maze{h: len(rows), grid: make([][]byte, len(rows))}
	for _, row := range rows {
		mz.w = max(mz.w, len(row))
	}
	if mz.w > maxSide || mz.h > maxSide || mz.w < 5 || mz.h < 5 {
		return nil, fmt.Errorf("%s: maze of %dx%d is outside 5..%d", path, mz.w, mz.h, maxSide)
	}
	for y, row := range rows {
		mz.grid[y] = make([]byte, mz.w) // pad short rows with spaces
		for x := range mz.w {
			mz.grid[y][x] = ' '
			if x < len(row) {
				mz.grid[y][x] = row[x]
			}
		}
	}
	mz.dist = make([]int32, cells*maxLevel)
	mz.queue = make([]int32, cells*maxLevel)
	return mz, mz.findPortals()
}

// letter reports whether the tile at x, y is a label letter.
func (mz *maze) letter(x, y int) bool {
	return x >= 0 && y >= 0 && x < mz.w && y < mz.h && mz.grid[y][x] >= 'A' && mz.grid[y][x] <= 'Z'
}

// findPortals pairs up the labelled tiles.
func (mz *maze) findPortals() error {
	for i := range mz.portal {
		mz.portal[i] = -1
	}
	for i := range mz.first {
		mz.first[i] = -1
	}
	mz.start, mz.end = -1, -1
	for y := range mz.h {
		for x := range mz.w {
			for _, d := range [2][2]int{{1, 0}, {0, 1}} {
				if !mz.letter(x, y) || !mz.letter(x+d[0], y+d[1]) {
					continue
				}
				// The open tile is just before or just after the pair.
				tx, ty := x-d[0], y-d[1]
				if tx < 0 || ty < 0 || mz.grid[ty][tx] != '.' {
					tx, ty = x+2*d[0], y+2*d[1]
				}
				if tx >= mz.w || ty >= mz.h || mz.grid[ty][tx] != '.' {
					return fmt.Errorf("label at %d,%d is next to no open tile", x, y)
				}
				label := int(mz.grid[y][x]-'A')*26 + int(mz.grid[y+d[1]][x+d[0]]-'A')
				if err := mz.addPortal(label, ty*mz.w+tx); err != nil {
					return err
				}
			}
		}
	}
	if mz.start < 0 || mz.end < 0 || mz.start == mz.end {
		return errors.New("the maze needs separate AA and ZZ tiles")
	}
	for label, tile := range mz.first {
		if tile >= 0 && mz.portal[tile] < 0 {
			return fmt.Errorf("label %c%c has no partner", 'A'+label/26, 'A'+label%26)
		}
	}
	return nil
}

// addPortal records one labelled tile.
func (mz *maze) addPortal(label, tile int) error {
	if label < 0 || label >= labels || tile < 0 || tile >= mz.w*mz.h {
		return fmt.Errorf("label %d at tile %d is out of range", label, tile)
	}
	x, y := tile%mz.w, tile/mz.w
	mz.outer[tile] = x == 2 || y == 2 || x == mz.w-3 || y == mz.h-3
	switch {
	case label == 0: // AA
		mz.start = tile
	case label == labels-1: // ZZ
		mz.end = tile
	case mz.first[label] < 0:
		mz.first[label] = tile
	case mz.portal[mz.first[label]] >= 0:
		return fmt.Errorf("label %c%c appears more than twice", 'A'+label/26, 'A'+label%26)
	default:
		mz.portal[tile], mz.portal[mz.first[label]] = mz.first[label], tile
	}
	return nil
}

// shortest returns the fewest steps from AA to ZZ. With recursion, inner
// portals lead a level down and outer ones a level up (closed at level 0),
// and ZZ counts only at level 0.
func (mz *maze) shortest(recursive bool) (int, error) {
	if mz.start < 0 || mz.end < 0 {
		return 0, errors.New("the maze has no AA or no ZZ")
	}
	levels := 1
	if recursive {
		levels = maxLevel
	}
	if len(mz.dist) < cells*levels || len(mz.queue) < cells*levels {
		return 0, errors.New("search buffers are too small")
	}
	for i := range mz.dist {
		mz.dist[i] = -1
	}
	mz.dist[mz.start], mz.queue[0] = 0, int32(mz.start)
	for head, tail := 0, 1; head < tail; head++ {
		state := int(mz.queue[head])
		tile, level := state%cells, state/cells
		if tile == mz.end && level == 0 {
			return int(mz.dist[state]), nil
		}
		var next [5]int
		for k, d := range [4]int{-mz.w, mz.w, -1, 1} {
			next[k] = state + d
		}
		next[4] = -1
		if p := mz.portal[tile]; p >= 0 {
			if mz.grid[p/mz.w][p%mz.w] != '.' {
				return 0, fmt.Errorf("portal from tile %d leads to a wall", tile)
			}
			nl := level
			if recursive && mz.outer[tile] {
				nl--
			} else if recursive {
				nl++
			}
			if nl >= 0 && nl < levels {
				next[4] = nl*cells + p
			}
		}
		for _, n := range next {
			if n < 0 || n%cells >= mz.w*mz.h || mz.grid[n%cells/mz.w][n%cells%mz.w] != '.' || mz.dist[n] >= 0 {
				continue
			}
			if tail == len(mz.queue) {
				return 0, errors.New("search queue is full")
			}
			mz.dist[n] = mz.dist[state] + 1
			mz.queue[tail], tail = int32(n), tail+1
		}
	}
	return 0, errors.New("ZZ cannot be reached")
}

func day20(path string) error {
	mz, err := load(path)
	if err != nil {
		return err
	}
	steps, err := mz.shortest(false)
	if err != nil {
		return err
	}

	deep, err := mz.shortest(true)
	if err != nil {
		return err
	}
	if steps < 1 {
		return fmt.Errorf("AA to ZZ took %d steps", steps)
	}
	// A recursive path is also a path through the flat maze.
	if deep < steps {
		return fmt.Errorf("recursive path of %d is shorter than the flat %d", deep, steps)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, deep); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
