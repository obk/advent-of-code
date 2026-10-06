// Day 7: No Space Left On Device
//
// The input is a terminal session: "$ cd x", "$ cd ..", "$ cd /", "$ ls",
// and ls output ("dir x" or "size name").
//
// Part 1 sums the sizes of directories of at most 100000 (a directory's
// size includes everything below it).
// Part 2 finds the smallest directory to delete so that 30000000 of the
// 70000000 disk is free.
//
// Directories are numbered as they are entered; the path from the root is
// a stack of their numbers, and a file's size is added to every directory
// on it. Each directory's contents are listed once.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 16
	maxDirs       = 1_000
	maxDepth      = 64
	disk          = 70_000_000
	needed        = 30_000_000
)

// sizes walks the session and returns each directory's total size.
func sizes(lines []string, size []int) (int, error) {
	if len(lines) == 0 || lines[0] != "$ cd /" {
		return 0, errors.New("the session must start at the root")
	}
	var path [maxDepth]int
	depth, dirs := 1, 1
	path[0] = 0
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		switch {
		case l == "$ cd /":
			depth = 1
		case l == "$ cd ..":
			if depth == 1 {
				return 0, errors.New("cd .. above the root")
			}
			depth--
		case len(f) == 3 && f[1] == "cd":
			if depth == maxDepth || dirs == len(size) {
				return 0, errors.New("too many or too deep directories")
			}
			path[depth], depth, dirs = dirs, depth+1, dirs+1
		case l == "$ ls" || (len(f) == 2 && f[0] == "dir"):
		case len(f) == 2:
			n, err := strconv.Atoi(f[0])
			if err != nil || n < 0 {
				return 0, fmt.Errorf("invalid file %q", l)
			}
			for _, d := range path[:depth] {
				size[d] += n
			}
		default:
			return 0, fmt.Errorf("unexpected line %q", l)
		}
	}
	return dirs, nil
}

func day7(path string) error {
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
	size := make([]int, maxDirs)
	dirs, err := sizes(strings.Split(strings.TrimSpace(string(content)), "\n"), size)
	if err != nil {
		return err
	}
	small, smallest := 0, math.MaxInt
	free := disk - size[0]
	for _, s := range size[:dirs] {
		if s <= 100_000 {
			small += s
		}
		if free+s >= needed {
			smallest = min(smallest, s)
		}
	}
	if smallest == math.MaxInt {
		return errors.New("deleting any one directory frees too little")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", small, smallest); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
