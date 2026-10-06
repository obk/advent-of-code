// Day 14: Disk Defragmentation
//
// The disk is a 128x128 grid. Row r is the 128 bits of the Knot Hash (from
// day 10) of the key string (the input), a dash, and r.
//
// Part 1 counts the used squares, the 1 bits.
// Part 2 counts the regions: groups of used squares joined edge to edge.
//
// The Knot Hash is computed into a fixed 16-byte array, and the rows are
// kept as a grid of booleans.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 10
	maxKeyLen     = 64
	side          = 128
	marks         = 256
)

// readKey returns the key string.
func readKey(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	key := strings.TrimSpace(string(content))
	if key == "" || len(key) > maxKeyLen {
		return "", fmt.Errorf("%s: key of %d bytes is outside 1..%d", path, len(key), maxKeyLen)
	}
	return key, nil
}

// knotHash computes the dense Knot Hash of text.
func knotHash(text []byte) ([16]byte, error) {
	var dense [16]byte
	if len(text) == 0 {
		return dense, errors.New("nothing to hash")
	}
	if len(text) > maxKeyLen+8 {
		return dense, fmt.Errorf("text of %d bytes is too long", len(text))
	}
	var list [marks]int
	for i := range marks {
		list[i] = i
	}
	suffix := [5]byte{17, 31, 73, 47, 23}
	pos, skip := 0, 0
	for range 64 {
		for k := range len(text) + len(suffix) {
			n := int(suffix[max(k-len(text), 0)])
			if k < len(text) {
				n = int(text[k])
			}
			for i, j := pos, pos+n-1; i < j; i, j = i+1, j-1 {
				list[i%marks], list[j%marks] = list[j%marks], list[i%marks]
			}
			pos, skip = (pos+n+skip)%marks, skip+1
		}
	}
	for block := range 16 {
		x := 0
		for i := range 16 {
			x ^= list[block*16+i]
		}
		if x < 0 || x > 255 {
			return dense, fmt.Errorf("dense hash byte %d is %d", block, x)
		}
		dense[block] = byte(x)
	}
	return dense, nil
}

// fill computes every row of the grid from the key.
func fill(key string, grid *[side][side]bool) (used int, err error) {
	if grid == nil {
		return 0, errors.New("no grid to fill")
	}
	if key == "" || len(key) > maxKeyLen {
		return 0, fmt.Errorf("key of %d bytes is outside 1..%d", len(key), maxKeyLen)
	}
	var buf [maxKeyLen + 8]byte
	for r := range side {
		text := strconv.AppendInt(append(append(buf[:0], key...), '-'), int64(r), 10)
		hash, err := knotHash(text)
		if err != nil {
			return 0, err
		}
		for c := range side {
			grid[r][c] = hash[c/8]>>(7-c%8)&1 == 1
		}
		for _, b := range hash {
			used += bits.OnesCount8(b)
		}
	}
	return used, nil
}

// countRegions flood-fills each unvisited used square with a fixed queue.
func countRegions(grid *[side][side]bool) (int, error) {
	if grid == nil {
		return 0, errors.New("no grid")
	}
	var seen [side][side]bool
	var queue [side * side]int
	regions := 0
	for start := range side * side {
		r, c := start/side, start%side
		if !grid[r][c] || seen[r][c] {
			continue
		}
		regions++
		seen[r][c], queue[0] = true, start
		tail := 1
		for head := 0; head < tail; head++ {
			y, x := queue[head]/side, queue[head]%side
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				ny, nx := y+d[0], x+d[1]
				if ny < 0 || nx < 0 || ny >= side || nx >= side || !grid[ny][nx] || seen[ny][nx] {
					continue
				}
				seen[ny][nx], queue[tail] = true, ny*side+nx
				tail++
			}
		}
	}
	if regions > side*side {
		return 0, fmt.Errorf("%d regions on a %dx%d disk", regions, side, side)
	}
	return regions, nil
}

func day14(path string) error {
	key, err := readKey(path)
	if err != nil {
		return err
	}
	var grid [side][side]bool
	used, err := fill(key, &grid)
	if err != nil {
		return err
	}

	if used > side*side {
		return fmt.Errorf("%d used squares on a %dx%d disk", used, side, side)
	}
	regions, err := countRegions(&grid)
	if err != nil {
		return err
	}
	if regions > used {
		return fmt.Errorf("%d regions but only %d used squares", regions, used)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", used, regions); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
