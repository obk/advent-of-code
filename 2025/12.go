// Day 12: Christmas Tree Farm
//
// The input first defines the presents, each an ID line followed by a drawing
// of the cells it covers, then lists the regions to fill:
//
//	0:
//	###
//	.#.
//
//	12x5: 1 0 2
//
// A region line gives its width and height and then how many of each present
// it has to hold, in present ID order. The answer is how many regions can take
// all of their presents at once, with no overlaps and nothing hanging over the
// edge. Presents may be turned and flipped freely.
//
// This is exact cover, and it is solved by backtracking: place a present in
// every orientation at every position that fits, move on to the next, and undo
// the placement when the rest cannot be made to work. The presents placed so
// far live on an explicit stack, so backtracking is popping it. Three things
// keep that in budget - the orientations of each present are worked out once
// up front rather than per attempt; a region whose presents cover more cells
// than it has is rejected without any search at all; and the presents are
// placed biggest first, because a large present has the fewest positions
// available and so fails fastest when the region was never going to work.
//
// A present's drawing is kept as a bitmask over a small square, which makes
// turning, flipping and comparing orientations plain integer operations.
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
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxShapes     = 10
	shapeSide     = 5 // a drawing fits in a shapeSide x shapeSide square
	maxShapeCells = shapeSide * shapeSide
	maxRegionSide = 100
	maxPresents   = 1_000
	// A cap on the placements tried per region. The real regions either
	// fail the area check or pack without a single step back.
	maxSearchSteps = 10_000_000
)

// orientations is the most distinct ways a present can lie: four turns, each
// optionally mirrored.
const orientations = 8

// shape is one orientation of a present: the cells it occupies, relative to
// its top-left corner and in row-major order, and the box it fills.
type shape struct {
	rows, cols    [maxShapeCells]int
	cells         int
	height, width int
}

// shapeSet is every distinct orientation of one present.
type shapeSet struct {
	shapes [orientations]shape
	n      int
}

// query is one region to fill: its size and how many of each present it must
// hold.
type query struct {
	width, height int
	counts        [maxShapes]int
}

// frame is one present on the search stack: the next placement to try for it,
// and the placement currently on the grid (-1 if none). A placement is an
// index into every orientation at every grid position.
type frame struct {
	cursor int
	placed int
}

// workspace holds every buffer the search needs, sized for the largest region
// the limits allow.
type workspace struct {
	covered  [maxRegionSide * maxRegionSide]bool
	presents [maxPresents]int
	stack    [maxPresents + 1]frame
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

// addDrawingRow adds one row of a present's drawing to its bitmask, where the
// cell at (row, col) is bit row*shapeSide + col.
func addDrawingRow(mask uint32, row int, line string) (uint32, error) {
	if row < 0 || row >= shapeSide {
		return 0, fmt.Errorf("drawing has more than %d rows", shapeSide)
	}
	if len(line) > shapeSide {
		return 0, fmt.Errorf("drawing row %q is wider than %d", line, shapeSide)
	}
	if strings.Trim(line, ".#") != "" {
		return 0, fmt.Errorf("drawing row %q holds something other than '.' and '#'", line)
	}

	for col := range len(line) {
		if line[col] == '#' {
			mask |= 1 << (row*shapeSide + col)
		}
	}
	return mask, nil
}

// parseQuery reads a line such as "12x5: 1 0 2", where the counts are listed in
// present ID order, one for each of the numShapes presents.
func parseQuery(line string, numShapes int) (query, error) {
	size, counts, _ := strings.Cut(line, ":")
	widthText, heightText, found := strings.Cut(size, "x")
	if !found {
		return query{}, fmt.Errorf("invalid region size %q", line)
	}
	width, err := strconv.Atoi(strings.TrimSpace(widthText))
	if err != nil {
		return query{}, fmt.Errorf("invalid region width %q: %w", line, err)
	}
	height, err := strconv.Atoi(strings.TrimSpace(heightText))
	if err != nil {
		return query{}, fmt.Errorf("invalid region height %q: %w", line, err)
	}
	if width < 1 || width > maxRegionSide || height < 1 || height > maxRegionSide {
		return query{}, fmt.Errorf("region %q is outside 1..%d on a side", line, maxRegionSide)
	}

	fields := strings.Fields(counts)
	if len(fields) != numShapes {
		return query{}, fmt.Errorf("region %q lists %d counts for %d presents", line, len(fields), numShapes)
	}
	q := query{width: width, height: height}
	for id, countText := range fields {
		if q.counts[id], err = strconv.Atoi(countText); err != nil {
			return query{}, fmt.Errorf("invalid present count %q: %w", line, err)
		}
		if q.counts[id] < 0 || q.counts[id] > maxPresents {
			return query{}, fmt.Errorf("present count in %q is outside 0..%d", line, maxPresents)
		}
	}
	return q, nil
}

// parseInput reads the present drawings, as bitmasks indexed by ID, and the
// region queries. Presents are introduced by an ID line ("3:") followed by
// their drawing, and must be numbered 0, 1, 2, ... in order.
func parseInput(path string) (masks []uint32, queries []query, err error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, nil, err
	}

	masks = make([]uint32, 0, maxShapes)
	queries = make([]query, 0, len(lines))
	row := 0 // the next row of the drawing being read
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
		case strings.Contains(line, "x"):
			q, err := parseQuery(line, len(masks))
			if err != nil {
				return nil, nil, err
			}
			queries = append(queries, q)
			row = -1 // no drawing rows may follow a region
		case strings.HasSuffix(line, ":"):
			id, err := strconv.Atoi(strings.TrimSuffix(line, ":"))
			if err != nil || id != len(masks) || id >= maxShapes {
				return nil, nil, fmt.Errorf("present %q is out of sequence or invalid", line)
			}
			masks = append(masks, 0)
			row = 0
		case len(masks) > 0:
			if masks[len(masks)-1], err = addDrawingRow(masks[len(masks)-1], row, line); err != nil {
				return nil, nil, err
			}
			row++
		default:
			return nil, nil, fmt.Errorf("drawing row %q comes before any present ID", line)
		}
	}

	if len(masks) == 0 || len(queries) == 0 {
		return nil, nil, fmt.Errorf("%s: %d presents and %d regions", path, len(masks), len(queries))
	}
	for id, mask := range masks {
		if mask == 0 {
			return nil, nil, fmt.Errorf("%s: present %d has an empty drawing", path, id)
		}
	}
	return masks, queries, nil
}

// normalize slides a drawing up and left until it touches both edges of its
// square, so drawings that differ only by position compare equal.
func normalize(mask uint32) (uint32, error) {
	if mask == 0 || mask >= 1<<maxShapeCells {
		return 0, fmt.Errorf("drawing %b is empty or too big", mask)
	}

	var firstCol uint32 // the bits of column 0, in every row
	for row := range shapeSide {
		firstCol |= 1 << (row * shapeSide)
	}
	for range shapeSide {
		if mask&(1<<shapeSide-1) != 0 {
			break
		}
		mask >>= shapeSide // drop the empty top row
	}
	for range shapeSide {
		if mask&firstCol != 0 {
			break
		}
		mask >>= 1 // drop the empty left column
	}
	return mask, nil
}

// transform turns a drawing 90 degrees, (row, col) -> (col, side-1-row), or
// mirrors it, (row, col) -> (row, side-1-col), and normalizes the result.
func transform(mask uint32, turn bool) (uint32, error) {
	if mask == 0 || mask >= 1<<maxShapeCells {
		return 0, fmt.Errorf("drawing %b is empty or too big", mask)
	}

	var moved uint32
	for row := range shapeSide {
		for col := range shapeSide {
			if mask>>(row*shapeSide+col)&1 == 0 {
				continue
			}
			if turn {
				moved |= 1 << (col*shapeSide + shapeSide - 1 - row)
			} else {
				moved |= 1 << (row*shapeSide + shapeSide - 1 - col)
			}
		}
	}
	return normalize(moved)
}

// toShape lists a normalized drawing's cells in row-major order.
func toShape(mask uint32) (shape, error) {
	if mask == 0 || mask >= 1<<maxShapeCells {
		return shape{}, fmt.Errorf("drawing %b is empty or too big", mask)
	}

	var s shape
	for bit := range maxShapeCells {
		if mask>>bit&1 == 0 {
			continue
		}
		s.rows[s.cells], s.cols[s.cells] = bit/shapeSide, bit%shapeSide
		s.height = max(s.height, s.rows[s.cells]+1)
		s.width = max(s.width, s.cols[s.cells]+1)
		s.cells++
	}
	if s.cells == 0 || s.cells > maxShapeCells {
		return shape{}, fmt.Errorf("drawing %b has %d cells", mask, s.cells)
	}
	return s, nil
}

// buildVariations returns the present's distinct orientations: its four
// rotations and the mirror image of each, minus any duplicates a symmetric
// present produces. A square block collapses to one, and dropping the
// duplicates here saves the search from retrying placements it has already
// ruled out.
func buildVariations(mask uint32) (shapeSet, error) {
	current, err := normalize(mask)
	if err != nil {
		return shapeSet{}, err
	}

	var seen [orientations]uint32
	var set shapeSet
	for range orientations / 2 {
		mirrored, err := transform(current, false)
		if err != nil {
			return shapeSet{}, err
		}
		for _, candidate := range [2]uint32{current, mirrored} {
			duplicate := false
			for _, m := range seen[:set.n] {
				duplicate = duplicate || m == candidate
			}
			if duplicate {
				continue
			}
			if set.shapes[set.n], err = toShape(candidate); err != nil {
				return shapeSet{}, err
			}
			seen[set.n] = candidate
			set.n++
		}
		if current, err = transform(current, true); err != nil {
			return shapeSet{}, err
		}
	}

	if set.n < 1 || set.n > orientations {
		return shapeSet{}, fmt.Errorf("drawing %b has %d orientations", mask, set.n)
	}
	return set, nil
}

// orderPresents writes the query's presents into w.presents, biggest first,
// and returns how many there are.
func (w *workspace) orderPresents(q query, sets []shapeSet) (int, error) {
	if len(sets) > maxShapes {
		return 0, fmt.Errorf("%d presents, more than the %d allowed", len(sets), maxShapes)
	}

	for id := len(sets); id < maxShapes; id++ {
		if q.counts[id] != 0 {
			return 0, fmt.Errorf("region asks for present %d, which is not defined", id)
		}
	}

	n, want := 0, 0
	for _, count := range q.counts {
		want += count
	}
	for cells := maxShapeCells; cells > 0; cells-- {
		for id, set := range sets {
			if set.shapes[0].cells != cells {
				continue
			}
			if n+q.counts[id] > maxPresents {
				return 0, fmt.Errorf("region holds more than %d presents", maxPresents)
			}
			for range q.counts[id] {
				w.presents[n] = id
				n++
			}
		}
	}
	if n != want {
		return 0, fmt.Errorf("ordered %d presents, expected %d", n, want)
	}
	return n, nil
}

// placementAt decodes a placement index into the orientation and the grid
// position of its top-left corner.
func placementAt(set shapeSet, index, width, height int) (s shape, row, col int, err error) {
	positions := width * height
	if set.n < 1 || positions < 1 {
		return shape{}, 0, 0, fmt.Errorf("nothing to place: %d orientations on %d cells", set.n, positions)
	}
	if index < 0 || index >= set.n*positions {
		return shape{}, 0, 0, fmt.Errorf("placement %d is outside 0..%d", index, set.n*positions-1)
	}
	s = set.shapes[index/positions]
	row, col = index%positions/width, index%width
	if row+s.height > height || col+s.width > width {
		return shape{}, 0, 0, fmt.Errorf("placement %d hangs over the edge", index)
	}
	return s, row, col, nil
}

// nextPlacement finds the first placement at or after index from where the
// shape lies wholly on the grid without covering anything, reporting false if
// there is none.
func (w *workspace) nextPlacement(set shapeSet, from, width, height int) (int, bool, error) {
	positions := width * height
	if set.n < 1 || positions < 1 {
		return 0, false, fmt.Errorf("nothing to place: %d orientations on %d cells", set.n, positions)
	}
	if positions > len(w.covered) {
		return 0, false, fmt.Errorf("region of %d cells is bigger than the grid buffer", positions)
	}
	for index := max(from, 0); index < set.n*positions; index++ {
		s := set.shapes[index/positions]
		row, col := index%positions/width, index%width
		if row+s.height > height || col+s.width > width {
			continue
		}
		free := true
		for i := range s.cells {
			free = free && !w.covered[(row+s.rows[i])*width+col+s.cols[i]]
		}
		if free {
			return index, true, nil
		}
	}
	return 0, false, nil
}

// setPlacement covers (or uncovers) the cells of the placement at index. Every
// cell must currently be in the opposite state.
func (w *workspace) setPlacement(set shapeSet, index, width, height int, covered bool) error {
	s, row, col, err := placementAt(set, index, width, height)
	if err != nil {
		return err
	}
	for i := range s.cells {
		cell := (row+s.rows[i])*width + col + s.cols[i]
		if w.covered[cell] == covered {
			return fmt.Errorf("cell %d is already in the state being set", cell)
		}
		w.covered[cell] = covered
	}
	return nil
}

// search reports whether the n presents in w.presents can all be placed on
// an empty width x height grid, trying each orientation at each position and
// backtracking on a dead end. The stack holds one frame per present placed,
// so the grid always reflects the path currently being explored.
func (w *workspace) search(n int, sets []shapeSet, width, height int) (bool, error) {
	if n < 0 || n > maxPresents {
		return false, fmt.Errorf("%d presents is outside 0..%d", n, maxPresents)
	}
	if width < 1 || width > maxRegionSide || height < 1 || height > maxRegionSide {
		return false, fmt.Errorf("region %dx%d is outside 1..%d on a side", width, height, maxRegionSide)
	}

	if len(sets) == 0 {
		return false, errors.New("no presents defined")
	}

	clear(w.covered[:width*height])
	top := 0
	w.stack[0] = frame{placed: -1}
	for range maxSearchSteps {
		if top == n || top < 0 {
			return top == n, nil
		}
		f := &w.stack[top]
		set := sets[w.presents[top]]
		if f.placed >= 0 { // back from a dead end: lift this present off again
			if err := w.setPlacement(set, f.placed, width, height, false); err != nil {
				return false, err
			}
			f.placed = -1
		}

		index, found, err := w.nextPlacement(set, f.cursor, width, height)
		if err != nil {
			return false, err
		}
		if !found {
			top-- // no placement left for this present: backtrack
			continue
		}
		if err := w.setPlacement(set, index, width, height, true); err != nil {
			return false, err
		}
		f.placed, f.cursor = index, index+1
		top++
		w.stack[top] = frame{placed: -1}
	}
	return false, fmt.Errorf("search exceeded %d steps", maxSearchSteps)
}

// canFillRegion reports whether the query's presents can all be packed into
// its region.
func (w *workspace) canFillRegion(q query, sets []shapeSet) (bool, error) {
	if q.width*q.height > len(w.covered) {
		return false, fmt.Errorf("region %dx%d is bigger than the grid buffer", q.width, q.height)
	}
	n, err := w.orderPresents(q, sets)
	if err != nil {
		return false, err
	}

	// The presents cannot possibly fit if they cover more cells than the
	// region has.
	cells := 0
	for _, id := range w.presents[:n] {
		cells += sets[id].shapes[0].cells
	}
	if cells > q.width*q.height {
		return false, nil
	}
	return w.search(n, sets, q.width, q.height)
}

func day12(path string) error {
	masks, queries, err := parseInput(path)
	if err != nil {
		return err
	}

	// Every orientation of every present is needed over and over, so work
	// them out once up front.
	sets := make([]shapeSet, len(masks))
	for id, mask := range masks {
		if sets[id], err = buildVariations(mask); err != nil {
			return err
		}
	}

	w := new(workspace)
	solved := 0
	for _, q := range queries {
		fits, err := w.canFillRegion(q, sets)
		if err != nil {
			return err
		}
		if fits {
			solved++
		}
	}

	if _, err := fmt.Printf("Part 1: %d\n", solved); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day12(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
