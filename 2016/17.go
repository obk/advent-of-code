// Day 17: Two Steps Forward
//
// You move through a 4x4 grid of rooms from the top left to the vault at the
// bottom right. Which doors of the current room are open depends on the MD5
// hash of the passcode (the input) followed by the path so far as letters U,
// D, L, R: the first four hex digits stand for up, down, left and right, and
// b-f mean open.
//
// Part 1 is the shortest path to the vault.
// Part 2 is the length of the longest path that reaches it; a path ends as
// soon as it arrives, so the search never continues from the vault.
//
// The search is a breadth-first search over paths, since the same room with a
// different path is a different state. Paths are not stored as strings: each
// queued node keeps its parent and its last move, and the path is rebuilt
// into one reused buffer when it is needed for hashing.
package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 10
	maxPasscode   = 64
	maxNodes      = 1 << 22
	maxPath       = 1 << 12
	side          = 4
)

// node is one path in the search tree, kept small because the exhaustive
// search for part 2 stores a great many of them.
type node struct {
	parent int32 // index of the node this one extends, -1 for the start
	depth  int16
	move   byte
	x, y   int8
}

// vault holds the search tree and the path buffer.
type vault struct {
	nodes []node
	path  []byte // passcode followed by the path being hashed
}

// readPasscode returns the passcode.
func readPasscode(path string) (string, error) {
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
	code := strings.TrimSpace(string(content))
	if code == "" || len(code) > maxPasscode {
		return "", fmt.Errorf("%s: passcode of %d bytes is outside 1..%d", path, len(code), maxPasscode)
	}
	return code, nil
}

// spell writes node i's path after the passcode in v.path and returns it.
func (v *vault) spell(i, codeLen int) ([]byte, error) {
	if i < 0 || i >= len(v.nodes) {
		return nil, fmt.Errorf("node %d does not exist", i)
	}
	depth := int(v.nodes[i].depth)
	if codeLen+depth > cap(v.path) {
		return nil, fmt.Errorf("path of %d moves does not fit the buffer", depth)
	}
	p := v.path[:codeLen+depth]
	for k := codeLen + depth - 1; k >= codeLen; k-- {
		p[k] = v.nodes[i].move
		i = int(v.nodes[i].parent)
	}
	if i != 0 {
		return nil, errors.New("path did not lead back to the start")
	}
	return p, nil
}

// search explores every path from the start, returning the shortest path to
// the vault and the length of the longest.
func (v *vault) search(code string) (shortest string, longest int, err error) {
	if code == "" || len(code) > maxPasscode {
		return "", 0, errors.New("invalid passcode")
	}
	v.path = append(v.path[:0], code...)
	v.nodes = append(v.nodes[:0], node{parent: -1})

	moves := [4]struct {
		letter byte
		dx, dy int
	}{{'U', 0, -1}, {'D', 0, 1}, {'L', -1, 0}, {'R', 1, 0}}
	for head := 0; head < len(v.nodes); head++ {
		n := v.nodes[head]
		p, err := v.spell(head, len(code))
		if err != nil {
			return "", 0, err
		}
		if n.x == side-1 && n.y == side-1 {
			if shortest == "" {
				shortest = string(p[len(code):]) // BFS: the first arrival is shortest
			}
			longest = int(n.depth) // and every later one is at least as long
			continue
		}
		sum := md5.Sum(p)
		for d, m := range moves {
			digit := sum[d/2] >> (4 * (1 - d%2)) & 0x0F
			nx, ny := int(n.x)+m.dx, int(n.y)+m.dy
			if digit < 0xB || nx < 0 || ny < 0 || nx >= side || ny >= side {
				continue
			}
			if len(v.nodes) == cap(v.nodes) || int(n.depth)+1 >= maxPath {
				return "", 0, fmt.Errorf("more than %d paths or %d moves to explore", cap(v.nodes), maxPath)
			}
			v.nodes = append(v.nodes, node{parent: int32(head), move: m.letter, x: int8(nx), y: int8(ny), depth: n.depth + 1})
		}
	}
	if shortest == "" {
		return "", 0, errors.New("the vault cannot be reached")
	}
	return shortest, longest, nil
}

func day17(path string) error {
	code, err := readPasscode(path)
	if err != nil {
		return err
	}

	v := &vault{nodes: make([]node, 0, maxNodes), path: make([]byte, 0, maxPasscode+maxPath)}
	shortest, longest, err := v.search(code)
	if err != nil {
		return err
	}

	if strings.Trim(shortest, "UDLR") != "" {
		return fmt.Errorf("path %q holds something other than moves", shortest)
	}
	if longest < len(shortest) {
		return fmt.Errorf("longest path %d is shorter than the shortest %d", longest, len(shortest))
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", shortest, longest); err != nil {
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
