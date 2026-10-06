// Day 17: Reservoir Research
//
// Each line of the input is a vein of clay, "x=495, y=2..7" or "y=7,
// x=495..501". Water pours from a spring at x=500, y=0: it falls through
// sand, spreads left and right over clay or settled water, settles where it
// is walled in on both sides, and spills over any edge it reaches.
//
// Part 1 counts the tiles water reaches (settled or flowing) between the
// smallest and largest y in the scan.
// Part 2 counts only the settled water, which is what remains once the
// spring stops.
//
// The flow is simulated without recursion, with a stack of points water
// falls from. Water falls until something is beneath it, then spreads along
// that row. Walled in on both sides, the row settles and the point above is
// pushed to spread again one level up; otherwise the row is flowing water
// and each edge it spills over becomes a new falling point.
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
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 20
	maxLines      = 5_000
	maxX          = 2_000
	maxY          = 3_000
	maxStack      = 100_000
	maxSteps      = 10_000_000
	springX       = 500
)

// Tile kinds.
const (
	sand    = '.'
	clay    = '#'
	flowing = '|'
	settled = '~'
)

// ground is the scan as a grid, with its y range.
type ground struct {
	tiles      [maxY + 2][maxX + 2]byte
	minY, maxY int
	stack      [maxStack][2]int
}

// load parses the veins of clay into the grid.
func (g *ground) load(path string) error {
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
	if len(lines) > maxLines {
		return fmt.Errorf("%s: %d lines, more than %d", path, len(lines), maxLines)
	}
	for y := range g.tiles {
		for x := range g.tiles[y] {
			g.tiles[y][x] = sand
		}
	}
	g.minY, g.maxY = maxY, 0
	for _, line := range lines {
		f := strings.Fields(strings.NewReplacer("=", " ", ",", " ", "..", " ").Replace(line))
		if len(f) != 5 || (f[0] != "x" && f[0] != "y") {
			return fmt.Errorf("%s: malformed vein %q", path, line)
		}
		var n [3]int
		for i, s := range [3]string{f[1], f[3], f[4]} {
			if n[i], err = strconv.Atoi(s); err != nil {
				return fmt.Errorf("%s: invalid number in %q: %w", path, line, err)
			}
		}
		for k := n[1]; k <= n[2]; k++ {
			x, y := n[0], k
			if f[0] == "y" {
				x, y = k, n[0]
			}
			if x < 1 || x > maxX || y < 1 || y > maxY {
				return fmt.Errorf("%s: vein %q leaves the supported area", path, line)
			}
			g.tiles[y][x] = clay
			g.minY, g.maxY = min(g.minY, y), max(g.maxY, y)
		}
	}
	return nil
}

// spread scans row y from x towards dir (-1 or +1) over ground that holds
// water up. It returns the last tile reached and whether water spills there.
func (g *ground) spread(x, y, dir int) (int, bool, error) {
	if y < 0 || y >= g.maxY || x < 1 || x > maxX {
		return 0, false, fmt.Errorf("cannot spread from %d,%d", x, y)
	}
	for range maxX {
		below := g.tiles[y+1][x]
		if below == sand || below == flowing {
			return x, true, nil
		}
		next := g.tiles[y][x+dir]
		if next == clay {
			return x, false, nil
		}
		x += dir
		if x < 1 || x > maxX {
			return 0, false, errors.New("water spread past the edge of the supported area")
		}
	}
	return 0, false, errors.New("spread did not end")
}

// push adds a falling point to the stack.
func (g *ground) push(top, x, y int) (int, error) {
	if top >= maxStack {
		return 0, fmt.Errorf("more than %d falling points", maxStack)
	}
	g.stack[top] = [2]int{x, y}
	return top + 1, nil
}

// flow runs the water until nothing more changes.
func (g *ground) flow() error {
	if g.maxY < g.minY || g.maxY < 1 {
		return errors.New("the scan holds no clay")
	}
	top, err := g.push(0, springX, 0)
	if err != nil {
		return err
	}
	for steps := 0; top > 0; steps++ {
		if steps > maxSteps {
			return fmt.Errorf("water still moving after %d steps", maxSteps)
		}
		top--
		x, y := g.stack[top][0], g.stack[top][1]
		if x < 1 || x > maxX || y < 0 || y > g.maxY {
			return fmt.Errorf("falling point %d,%d is outside the scan", x, y)
		}
		for k := 0; k < maxY && y < g.maxY && g.tiles[y+1][x] == sand; k++ {
			y++
			g.tiles[y][x] = flowing
		}
		if y >= g.maxY || g.tiles[y+1][x] == flowing {
			continue // fell out of the scan, or joined water already flowing
		}
		l, spillL, err := g.spread(x, y, -1)
		if err != nil {
			return err
		}
		r, spillR, err := g.spread(x, y, 1)
		if err != nil {
			return err
		}
		fill := byte(flowing)
		if !spillL && !spillR {
			fill = settled
		}
		for k := l; k <= r; k++ {
			g.tiles[y][k] = fill
		}
		switch {
		case fill == settled && y > 0:
			top, err = g.push(top, x, y-1) // spread again one level up
		case fill == flowing && spillL && spillR:
			if top, err = g.push(top, l, y); err == nil {
				top, err = g.push(top, r, y)
			}
		case spillL:
			top, err = g.push(top, l, y)
		case spillR:
			top, err = g.push(top, r, y)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// count returns how many tiles in the scan's y range hold water at all, and
// how many hold settled water.
func (g *ground) count() (wet, still int, err error) {
	if g.minY < 0 || g.maxY > maxY || g.minY > g.maxY {
		return 0, 0, fmt.Errorf("y range %d..%d is invalid", g.minY, g.maxY)
	}
	for y := g.minY; y <= g.maxY; y++ {
		for _, t := range g.tiles[y] {
			if t == flowing || t == settled {
				wet++
			}
			if t == settled {
				still++
			}
		}
	}
	return wet, still, nil
}

func day17(path string) error {
	g := new(ground)
	if err := g.load(path); err != nil {
		return err
	}
	if err := g.flow(); err != nil {
		return err
	}
	wet, still, err := g.count()
	if err != nil {
		return err
	}
	if wet == 0 {
		return errors.New("the water reached nothing")
	}

	if still > wet {
		return fmt.Errorf("%d settled tiles but only %d wet ones", still, wet)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", wet, still); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
