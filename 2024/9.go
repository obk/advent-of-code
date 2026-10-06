// Day 9: Disk Fragmenter
//
// The input is a disk map: digits alternating between a file's length and
// a gap's length. Files are numbered from 0 in order.
//
// Part 1 moves blocks one at a time from the end into the leftmost free
// block, and gives the checksum (sum of position times file number).
// Part 2 moves whole files instead, each once in decreasing file number,
// to the leftmost gap left of it that fits, and gives the checksum.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 15
	maxBlocks     = 9 * maxInputBytes
)

// span is a run of blocks: a file (id >= 0) or a gap (id -1).
type span struct {
	start, length, id int
}

// blocks compacts block by block and returns the checksum.
func blocks(disk []int) (int, error) {
	if len(disk) == 0 {
		return 0, errors.New("empty disk")
	}
	lo, hi := 0, len(disk)-1
	// Each step moves an end or fills a gap (then the end moves next).
	for k := 0; k <= 2*len(disk) && lo < hi; k++ {
		switch {
		case disk[lo] >= 0:
			lo++
		case disk[hi] < 0:
			hi--
		default:
			disk[lo], disk[hi] = disk[hi], -1
		}
	}
	sum, seenGap := 0, false
	for i, id := range disk {
		if id >= 0 && seenGap {
			return 0, fmt.Errorf("block %d is still after a gap", i)
		}
		seenGap = seenGap || id < 0
		if id >= 0 {
			sum += i * id
		}
	}
	return sum, nil
}

// files compacts whole files and returns the checksum.
func files(spans []span, gaps []span) (int, error) {
	if len(spans) == 0 {
		return 0, errors.New("no files")
	}
	if len(gaps) > len(spans) {
		return 0, fmt.Errorf("%d gaps around %d files", len(gaps), len(spans))
	}
	for f := len(spans) - 1; f >= 0; f-- {
		file := &spans[f]
		for g := range gaps {
			if gaps[g].start >= file.start {
				break
			}
			if gaps[g].length >= file.length {
				file.start = gaps[g].start
				gaps[g].start += file.length
				gaps[g].length -= file.length
				break
			}
		}
	}
	sum := 0
	for _, s := range spans {
		for k := range s.length {
			sum += (s.start + k) * s.id
		}
	}
	return sum, nil
}

func day9(path string) error {
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
	text := strings.TrimSpace(string(content))
	if strings.Trim(text, "0123456789") != "" {
		return fmt.Errorf("%s: the disk map is not digits", path)
	}
	if len(text) == 0 {
		return fmt.Errorf("%s: empty disk map", path)
	}
	disk := make([]int, 0, maxBlocks)
	spans, gaps := make([]span, 0, len(text)/2+1), make([]span, 0, len(text)/2+1)
	for i := range len(text) {
		n, id := int(text[i]-'0'), -1
		if i%2 == 0 {
			id = i / 2
			spans = append(spans, span{len(disk), n, id})
		} else {
			gaps = append(gaps, span{len(disk), n, -1})
		}
		for range n {
			disk = append(disk, id)
		}
	}
	first, err := blocks(disk)
	if err != nil {
		return err
	}
	second, err := files(spans, gaps)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
