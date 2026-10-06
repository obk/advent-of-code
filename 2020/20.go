// Day 20: Jurassic Jigsaw
//
// The input is square image tiles, each 10x10 with an ID, scattered and
// rotated or flipped. Neighbouring tiles share their border.
//
// Part 1 multiplies the IDs of the four corner tiles.
// Part 2 assembles the image, drops each tile's border, finds the sea
// monsters (a fixed 20x3 pattern) in whichever orientation has them, and
// counts the '#' cells not part of any monster.
//
// An edge, read either way, matches at most one other tile. Corner tiles
// are those with two edges matching nothing. Assembly starts from a corner
// turned so its unmatched edges face up and left, then fills the grid row
// by row, trying each unused tile in all eight orientations against its
// left and upper neighbours.
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
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 16
	maxTiles      = 256
	n             = 10 // tile side
	maxSide       = 16 // tiles per image side
	maxImage      = maxSide * (n - 2)
)

// The sea monster.
const monster = "                  # \n#    ##    ##    ###\n #  #  #  #  #  #   "

// tile is one piece of the picture.
type tile struct {
	id   int
	grid [n][n]byte
}

// src maps (r, c) in orientation o (0-7) of a side x side square back to
// the unoriented square: rotations, then a mirror for o >= 4.
func src(o, side, r, c int) (int, int) {
	if o >= 4 {
		c = side - 1 - c
	}
	for k := 0; k < o%4; k++ {
		r, c = side-1-c, r
	}
	return r, c
}

// at returns cell (r, c) of tile t in orientation o.
func (t *tile) at(o, r, c int) byte {
	sr, sc := src(o, n, r, c)
	return t.grid[sr][sc]
}

// puzzle is the tiles and the assembly.
type puzzle struct {
	tiles  []tile
	count  map[int]int // how many tile edges have each value
	side   int
	place  [maxSide][maxSide][2]int // tile index and orientation
	used   []bool
	image  [maxImage][maxImage]byte
	marked [maxImage][maxImage]bool
}

// load parses the tiles.
func load(path string) (*puzzle, error) {
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
	if len(blocks) > maxTiles {
		return nil, fmt.Errorf("%s: more than %d tiles", path, maxTiles)
	}
	side := int(math.Sqrt(float64(len(blocks))))
	if side*side != len(blocks) || side < 2 || side > maxSide {
		return nil, fmt.Errorf("%s: %d tiles do not make a square", path, len(blocks))
	}
	p := &puzzle{tiles: make([]tile, len(blocks)), count: make(map[int]int, 8*len(blocks)), side: side, used: make([]bool, len(blocks))}
	for i, b := range blocks {
		lines := strings.Split(b, "\n")
		id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(lines[0], "Tile "), ":"))
		if err != nil || id < 1 || len(lines) != n+1 {
			return nil, fmt.Errorf("%s: malformed tile %q", path, lines[0])
		}
		for _, earlier := range p.tiles[:i] {
			if earlier.id == id {
				return nil, fmt.Errorf("%s: tile %d appears twice", path, id)
			}
		}
		p.tiles[i].id = id
		for r, row := range lines[1:] {
			if len(row) != n || strings.Trim(row, ".#") != "" {
				return nil, fmt.Errorf("%s: tile %d row %d is malformed", path, id, r)
			}
			copy(p.tiles[i].grid[r][:], row)
		}
	}
	for i := range p.tiles {
		e, err := p.edges(i)
		if err != nil {
			return nil, err
		}
		for _, code := range e {
			p.count[code]++
		}
	}
	for code, k := range p.count {
		if k > 2 {
			return nil, fmt.Errorf("%s: edge %b is on %d tiles; matches are not unique", path, code, k)
		}
	}
	return p, nil
}

// edgeCode returns the top (or left) edge of tile i in orientation o as a
// number, the same whichever way round the edge is read.
func (p *puzzle) edgeCode(i, o int, top bool) (int, error) {
	if i < 0 || i >= len(p.tiles) {
		return 0, fmt.Errorf("tile %d does not exist", i)
	}
	if o < 0 || o > 7 {
		return 0, fmt.Errorf("orientation %d does not exist", o)
	}
	fwd, back := 0, 0
	for k := range n {
		c := p.tiles[i].at(o, 0, k)
		if !top {
			c = p.tiles[i].at(o, k, 0)
		}
		bit := 0
		if c == '#' {
			bit = 1
		}
		fwd, back = fwd<<1|bit, back|bit<<k
	}
	return min(fwd, back), nil
}

// edges returns tile i's four edge codes: the top and left edges as it is,
// and as turned upside down (its bottom and right).
func (p *puzzle) edges(i int) ([4]int, error) {
	var out [4]int
	if i < 0 || i >= len(p.tiles) {
		return out, fmt.Errorf("tile %d does not exist", i)
	}
	for k := range 4 {
		var err error
		if out[k], err = p.edgeCode(i, 2*(k/2), k%2 == 0); err != nil {
			return out, err
		}
		if out[k] >= 1<<n {
			return out, fmt.Errorf("edge code %d is wider than %d bits", out[k], n)
		}
	}
	return out, nil
}

// unmatched counts tile i's edges that match no other tile.
func (p *puzzle) unmatched(i int) (int, error) {
	e, err := p.edges(i)
	if err != nil {
		return 0, err
	}
	free := 0
	for _, code := range e {
		if p.count[code] == 1 {
			free++
		}
	}
	if free > 2 {
		return 0, fmt.Errorf("tile %d has %d edges matching nothing", p.tiles[i].id, free)
	}
	return free, nil
}

// fits reports whether tile i in orientation o matches the tiles placed
// left of and above (r, c).
func (p *puzzle) fits(i, o, r, c int) (bool, error) {
	if i < 0 || i >= len(p.tiles) || o < 0 || o > 7 {
		return false, fmt.Errorf("tile %d orientation %d does not exist", i, o)
	}
	if r < 0 || c < 0 || r >= p.side || c >= p.side {
		return false, fmt.Errorf("place %d,%d is outside the grid", r, c)
	}
	t := &p.tiles[i]
	for k := 0; c > 0 && k < n; k++ {
		left := &p.tiles[p.place[r][c-1][0]]
		if t.at(o, k, 0) != left.at(p.place[r][c-1][1], k, n-1) {
			return false, nil
		}
	}
	for k := 0; r > 0 && k < n; k++ {
		up := &p.tiles[p.place[r-1][c][0]]
		if t.at(o, 0, k) != up.at(p.place[r-1][c][1], n-1, k) {
			return false, nil
		}
	}
	return true, nil
}

// cornerOrientation turns corner tile i so that its unmatched edges face
// up and left.
func (p *puzzle) cornerOrientation(i int) (int, error) {
	if i < 0 || i >= len(p.tiles) {
		return 0, fmt.Errorf("tile %d does not exist", i)
	}
	for o := range 8 {
		top, err := p.edgeCode(i, o, true)
		if err != nil {
			return 0, err
		}
		left, err := p.edgeCode(i, o, false)
		if err != nil {
			return 0, err
		}
		if p.count[top] == 1 && p.count[left] == 1 {
			return o, nil
		}
	}
	return 0, fmt.Errorf("corner %d cannot face its open edges outwards", p.tiles[i].id)
}

// corners returns the product of the corner IDs and one corner tile.
func (p *puzzle) corners() (product, first int, err error) {
	if len(p.tiles) < 4 {
		return 0, 0, errors.New("fewer than four tiles")
	}
	product, first, found := 1, -1, 0
	for i := range p.tiles {
		free, err := p.unmatched(i)
		if err != nil {
			return 0, 0, err
		}
		if free != 2 {
			continue
		}
		if p.tiles[i].id > 0 && product > math.MaxInt64/p.tiles[i].id {
			return 0, 0, errors.New("the product of corner IDs overflows")
		}
		product, found = product*p.tiles[i].id, found+1
		if first < 0 {
			first = i
		}
	}
	if found != 4 {
		return 0, 0, fmt.Errorf("%d corner tiles, expected 4", found)
	}
	return product, first, nil
}

// assemble places every tile, returning the product of the corner IDs.
func (p *puzzle) assemble() (int, error) {
	product, first, err := p.corners()
	if err != nil {
		return 0, err
	}
	o, err := p.cornerOrientation(first)
	if err != nil {
		return 0, err
	}
	p.place[0][0], p.used[first] = [2]int{first, o}, true
	for pos := 1; pos < p.side*p.side; pos++ {
		r, c := pos/p.side, pos%p.side
		found := false
		for i := 0; i < len(p.tiles) && !found; i++ {
			for o := 0; o < 8 && !found && !p.used[i]; o++ {
				ok, err := p.fits(i, o, r, c)
				if err != nil {
					return 0, err
				}
				if ok {
					p.place[r][c], p.used[i], found = [2]int{i, o}, true, true
				}
			}
		}
		if !found {
			return 0, fmt.Errorf("no tile fits at row %d, column %d", r, c)
		}
	}
	for i, u := range p.used {
		if !u {
			return 0, fmt.Errorf("tile %d was left over", p.tiles[i].id)
		}
	}
	return product, nil
}

// build copies the assembled tiles, without borders, into the image.
func (p *puzzle) build() error {
	size := p.side * (n - 2)
	if size > maxImage {
		return fmt.Errorf("image of %d is larger than %d", size, maxImage)
	}
	for y := range size {
		for x := range size {
			pl := p.place[y/(n-2)][x/(n-2)]
			if pl[0] < 0 || pl[0] >= len(p.tiles) || !p.used[pl[0]] {
				return fmt.Errorf("no tile placed at row %d, column %d", y/(n-2), x/(n-2))
			}
			p.image[y][x] = p.tiles[pl[0]].at(pl[1], 1+y%(n-2), 1+x%(n-2))
			if p.image[y][x] != '#' && p.image[y][x] != '.' {
				return fmt.Errorf("image cell %d,%d is %q", x, y, p.image[y][x])
			}
		}
	}
	return nil
}

// roughness finds the orientation with sea monsters and counts the '#'
// cells outside them.
func (p *puzzle) roughness() (int, error) {
	size := p.side * (n - 2)
	if size < 1 || size > maxImage {
		return 0, fmt.Errorf("image size %d is out of range", size)
	}
	pattern := strings.Split(monster, "\n")
	if len(pattern) != 3 {
		return 0, errors.New("malformed monster pattern")
	}
	for o := range 8 {
		clear(p.marked[:])
		found := 0
		for y := 0; y+len(pattern) <= size; y++ {
			for x := 0; x+len(pattern[0]) <= size; x++ {
				here, err := p.monsterAt(o, size, y, x, pattern, false)
				if err != nil {
					return 0, err
				}
				if !here {
					continue
				}
				if _, err := p.monsterAt(o, size, y, x, pattern, true); err != nil {
					return 0, err
				}
				found++
			}
		}
		if found == 0 {
			continue
		}
		rough, waves := 0, 0
		for y := range size {
			for x := range size {
				if p.image[y][x] == '#' {
					waves++
				}
				if p.image[y][x] == '#' && !p.marked[y][x] {
					rough++
				}
			}
		}
		if rough >= waves {
			return 0, errors.New("monsters were found but none of their cells were marked")
		}
		return rough, nil
	}
	return 0, errors.New("no orientation shows a sea monster")
}

// monsterAt reports whether a monster sits at (y, x) of the image in
// orientation o; with mark set, it marks the monster's cells (in the
// unoriented image).
func (p *puzzle) monsterAt(o, size, y, x int, pattern []string, mark bool) (bool, error) {
	if o < 0 || o > 7 || y < 0 || x < 0 || y+len(pattern) > size || x+len(pattern[0]) > size {
		return false, fmt.Errorf("monster at %d,%d in orientation %d leaves the image", x, y, o)
	}
	for dy, row := range pattern {
		for dx := range len(row) {
			if row[dx] != '#' {
				continue
			}
			sy, sx := src(o, size, y+dy, x+dx)
			if p.image[sy][sx] != '#' {
				return false, nil
			}
			if mark {
				p.marked[sy][sx] = true
			}
		}
	}
	return true, nil
}

func day20(path string) error {
	p, err := load(path)
	if err != nil {
		return err
	}
	corners, err := p.assemble()
	if err != nil {
		return err
	}
	if err := p.build(); err != nil {
		return err
	}
	rough, err := p.roughness()
	if err != nil {
		return err
	}

	if corners < 1 {
		return fmt.Errorf("corner product %d is not positive", corners)
	}
	if rough < 1 {
		return errors.New("the sea is all monster")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", corners, rough); err != nil {
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
