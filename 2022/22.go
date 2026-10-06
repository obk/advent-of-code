// Day 22: Monkey Map
//
// The input is a map of open tiles '.', walls '#' and nothing ' ', then a
// path: numbers of steps and turns L and R. You start on the leftmost open
// tile of the top row, facing right; walls stop a move.
//
// Part 1 wraps a move off the map to the far side of the same row or
// column.
// Part 2 folds the map into a cube and wraps over the cube's edges.
// Each answer is 1000*row + 4*column + facing (right 0, down 1, left 2,
// up 3), counting rows and columns from 1.
//
// For the cube, each face gets a 3D frame: the directions its map right
// (R) and down (D) point, and its outward normal (N), found by walking the
// net from the first face. Stepping right onto a face folds it so that
// its normal is the old right, and so on. In doubled coordinates, with
// the cube centred on the origin, a cell's centre is
// s*N + (2i-s+1)*R + (2j-s+1)*D; stepping over an edge in direction v
// gives the neighbouring cell's centre by adding v - N, on the face whose
// normal is v, and the walker then faces -N.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 15
	maxSide       = 256
	maxPath       = 8_192
)

// vec is a 3D direction.
type vec [3]int

// add returns a + b; dot is the dot product.
func (a vec) add(b vec) vec { return vec{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func (a vec) dot(b vec) int { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

// scale returns k * a.
func (a vec) scale(k int) vec { return vec{k * a[0], k * a[1], k * a[2]} }

// face is one cube face: where it sits on the map and its 3D frame.
type face struct {
	row, col int // top-left tile
	r, d, n  vec
	placed   bool
}

// board is the map, the faces and the path.
type board struct {
	rows     [maxSide][maxSide]byte
	w, h     int
	s        int       // face size
	faceAt   [8][8]int // face number by face-grid position, or -1
	faces    [6]face
	path     string
	startCol int
}

// tile returns the map tile at (x, y), or ' ' off the map.
func (b *board) tile(x, y int) byte {
	if x < 0 || y < 0 || x >= b.w || y >= b.h {
		return ' '
	}
	return b.rows[y][x]
}

// load parses the map and path.
func load(path string) (*board, error) {
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
	drawing, route, found := strings.Cut(strings.TrimRight(string(content), "\n"), "\n\n")
	lines := strings.Split(drawing, "\n")
	if !found || len(lines) > maxSide || len(route) > maxPath {
		return nil, fmt.Errorf("%s: malformed map or path", path)
	}
	b := &board{h: len(lines), path: strings.TrimSpace(route)}
	if b.path == "" {
		return nil, fmt.Errorf("%s: empty path", path)
	}
	tiles := 0
	for y, l := range lines {
		if len(l) > maxSide || strings.Trim(l, " .#") != "" {
			return nil, fmt.Errorf("%s: row %d is malformed", path, y)
		}
		b.w = max(b.w, len(l))
		for x := range maxSide {
			b.rows[y][x] = ' '
			if x < len(l) {
				b.rows[y][x] = l[x]
			}
		}
		tiles += len(strings.TrimSpace(l))
	}
	for b.s = 1; b.s*b.s*6 < tiles && b.s < maxSide; b.s++ {
	}
	if tiles == 0 {
		return nil, fmt.Errorf("%s: empty map", path)
	}
	if b.s*b.s*6 != tiles || b.w/b.s > 8 || b.h/b.s > 8 {
		return nil, fmt.Errorf("%s: %d tiles do not make six square faces", path, tiles)
	}
	b.startCol = strings.IndexByte(lines[0], '.')
	return b, b.fold()
}

// fold numbers the faces and gives each its 3D frame, walking the net
// breadth-first from the first face.
func (b *board) fold() error {
	n := 0
	for fr := range 8 {
		for fc := range 8 {
			b.faceAt[fr][fc] = -1
			if b.tile(fc*b.s, fr*b.s) == ' ' {
				continue
			}
			if n == 6 {
				return errors.New("more than six faces")
			}
			b.faces[n] = face{row: fr * b.s, col: fc * b.s}
			b.faceAt[fr][fc], n = n, n+1
		}
	}
	if n != 6 {
		return fmt.Errorf("%d faces, expected 6", n)
	}
	b.faces[0].r, b.faces[0].d, b.faces[0].n, b.faces[0].placed = vec{1, 0, 0}, vec{0, 1, 0}, vec{0, 0, 1}, true
	var queue [6]int
	tail := 1
	for head := 0; head < tail; head++ {
		f := b.faces[queue[head]]
		fr, fc := f.row/b.s, f.col/b.s
		// Right, down, left, up neighbours on the net, and their frames.
		next := [4][2]int{{fr, fc + 1}, {fr + 1, fc}, {fr, fc - 1}, {fr - 1, fc}}
		frames := [4][3]vec{{f.n.scale(-1), f.d, f.r}, {f.r, f.n.scale(-1), f.d}, {f.n, f.d, f.r.scale(-1)}, {f.r, f.n, f.d.scale(-1)}}
		for k, pos := range next {
			if pos[0] < 0 || pos[1] < 0 || pos[0] >= 8 || pos[1] >= 8 || b.faceAt[pos[0]][pos[1]] < 0 {
				continue
			}
			g := &b.faces[b.faceAt[pos[0]][pos[1]]]
			if g.placed {
				continue
			}
			g.r, g.d, g.n, g.placed = frames[k][0], frames[k][1], frames[k][2], true
			queue[tail], tail = b.faceAt[pos[0]][pos[1]], tail+1
		}
	}
	if tail != 6 {
		return errors.New("the net is not connected")
	}
	for i, f := range b.faces {
		// Each frame is right-handed: N = R x D.
		cross := vec{f.r[1]*f.d[2] - f.r[2]*f.d[1], f.r[2]*f.d[0] - f.r[0]*f.d[2], f.r[0]*f.d[1] - f.r[1]*f.d[0]}
		if cross != f.n {
			return fmt.Errorf("face %d has a mirrored frame", i)
		}
		for j := range i {
			if b.faces[j].n == f.n {
				return fmt.Errorf("faces %d and %d both face %v", j, i, f.n)
			}
		}
	}
	return nil
}

// wrapFlat returns where a step off the map from (x, y) in direction dir
// lands: the far end of the same row or column.
func (b *board) wrapFlat(x, y, dir int) (int, int, error) {
	if dir < 0 || dir > 3 {
		return 0, 0, fmt.Errorf("direction %d does not exist", dir)
	}
	dx, dy := [4]int{1, 0, -1, 0}[dir], [4]int{0, 1, 0, -1}[dir]
	for k := 0; k < maxSide && b.tile(x-dx, y-dy) != ' '; k++ {
		x, y = x-dx, y-dy
	}
	if b.tile(x, y) == ' ' {
		return 0, 0, errors.New("wrapped onto nothing")
	}
	return x, y, nil
}

// wrapCube returns where a step off a face from (x, y) in direction dir
// lands on the cube, and the new direction.
func (b *board) wrapCube(x, y, dir int) (int, int, int, error) {
	if dir < 0 || dir > 3 || b.tile(x, y) == ' ' {
		return 0, 0, 0, fmt.Errorf("cannot step from %d,%d in direction %d", x, y, dir)
	}
	if b.faceAt[y/b.s][x/b.s] < 0 {
		return 0, 0, 0, fmt.Errorf("%d,%d is not on a face", x, y)
	}
	f := b.faces[b.faceAt[y/b.s][x/b.s]]
	moves := [4]vec{f.r, f.d, f.r.scale(-1), f.d.scale(-1)}
	v, s := moves[dir], b.s
	p := f.n.scale(s).add(f.r.scale(2*(x-f.col) - s + 1)).add(f.d.scale(2*(y-f.row) - s + 1))
	p = p.add(v).add(f.n.scale(-1))
	for _, g := range b.faces {
		if g.n != v {
			continue
		}
		i, j := (p.dot(g.r)+s-1)/2, (p.dot(g.d)+s-1)/2
		if i < 0 || j < 0 || i >= s || j >= s {
			return 0, 0, 0, fmt.Errorf("stepped off the cube at %d,%d", i, j)
		}
		for nd, m := range [4]vec{g.r, g.d, g.r.scale(-1), g.d.scale(-1)} {
			if m == f.n.scale(-1) && b.tile(g.col+i, g.row+j) == ' ' {
				return 0, 0, 0, errors.New("wrapped onto nothing")
			}
			if m == f.n.scale(-1) {
				return g.col + i, g.row + j, nd, nil
			}
		}
	}
	return 0, 0, 0, errors.New("no face lies over the edge")
}

// walk follows the path and returns the password.
func (b *board) walk(cube bool) (int, error) {
	x, y, dir := b.startCol, 0, 0
	if x < 0 {
		return 0, errors.New("the top row has no open tile")
	}
	for k := 0; k < len(b.path); k++ {
		switch c := b.path[k]; c {
		case 'L':
			dir = (dir + 3) % 4
			continue
		case 'R':
			dir = (dir + 1) % 4
			continue
		}
		steps, last := 0, k
		for j := k; j < len(b.path) && b.path[j] >= '0' && b.path[j] <= '9'; j++ {
			steps, last = steps*10+int(b.path[j]-'0'), j
		}
		if last == k && (b.path[k] < '0' || b.path[k] > '9') {
			return 0, fmt.Errorf("unexpected %q in the path", b.path[k])
		}
		k = last
		for range steps {
			nx, ny, nd := x+[4]int{1, 0, -1, 0}[dir], y+[4]int{0, 1, 0, -1}[dir], dir
			var err error
			if b.tile(nx, ny) == ' ' && cube {
				nx, ny, nd, err = b.wrapCube(x, y, dir)
			} else if b.tile(nx, ny) == ' ' {
				nx, ny, err = b.wrapFlat(x, y, dir)
			}
			if err != nil {
				return 0, err
			}
			if b.tile(nx, ny) == '#' {
				break
			}
			x, y, dir = nx, ny, nd
		}
	}
	if b.tile(x, y) != '.' {
		return 0, fmt.Errorf("the walk ended on %q", b.tile(x, y))
	}
	return 1000*(y+1) + 4*(x+1) + dir, nil
}

func day22(path string) error {
	b, err := load(path)
	if err != nil {
		return err
	}
	flat, err := b.walk(false)
	if err != nil {
		return err
	}
	cube, err := b.walk(true)
	if err != nil {
		return err
	}

	if flat < 1005 || cube < 1005 {
		return fmt.Errorf("passwords %d and %d are below the smallest possible", flat, cube)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", flat, cube); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
