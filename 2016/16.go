// Day 16: Dragon Checksum
//
// The input is a string of 0s and 1s. It is grown to fill a disk with a
// modified dragon curve - the data a becomes a, then 0, then a reversed with
// every bit flipped - and cut to the disk's length. Its checksum replaces each
// pair of bits with 1 if they match and 0 if not, repeating until the length
// is odd.
//
// Part 1 is the checksum for a disk of length 272.
// Part 2 is the checksum for a disk of length 35651584.
//
// The data is grown in one buffer allocated for the largest disk, and the
// checksum is computed in place, each round overwriting the front half.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 10
	maxDisk       = 1 << 26
	firstDisk     = 272
	secondDisk    = 35651584
)

// readSeed returns the initial state.
func readSeed(path string) (string, error) {
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
	seed := strings.TrimSpace(string(content))
	if seed == "" || strings.Trim(seed, "01") != "" {
		return "", fmt.Errorf("%s: %q is not a string of 0s and 1s", path, seed)
	}
	return seed, nil
}

// checksum fills a disk of the given length from the seed, using buf (which
// must have room for twice the disk) for the data, and returns the checksum.
func checksum(seed string, disk int, buf []byte) (string, error) {
	if disk < 1 || disk > maxDisk || seed == "" {
		return "", fmt.Errorf("disk length %d is outside 1..%d", disk, maxDisk)
	}
	if cap(buf) < 2*disk+len(seed)+1 {
		return "", fmt.Errorf("buffer of %d bytes is too small for a disk of %d", cap(buf), disk)
	}

	data := append(buf[:0], seed...)
	// Each round more than doubles the data, so 64 rounds always suffice.
	for round := 0; round < 64 && len(data) < disk; round++ {
		n := len(data)
		data = append(data, '0')
		for i := n - 1; i >= 0; i-- {
			data = append(data, '0'+'1'-data[i])
		}
	}
	if len(data) < disk {
		return "", errors.New("the data did not grow to fill the disk")
	}

	data = data[:disk]
	if len(bytes.Trim(data[:min(disk, 64)], "01")) != 0 {
		return "", errors.New("the data holds something other than 0s and 1s")
	}
	for round := 0; round < 64 && len(data)%2 == 0; round++ {
		for i := range len(data) / 2 {
			if data[2*i] == data[2*i+1] {
				data[i] = '1'
			} else {
				data[i] = '0'
			}
		}
		data = data[:len(data)/2]
	}
	return string(data), nil
}

func day16(path string) error {
	seed, err := readSeed(path)
	if err != nil {
		return err
	}

	buf := make([]byte, 0, 2*secondDisk+len(seed)+1)
	first, err := checksum(seed, firstDisk, buf)
	if err != nil {
		return err
	}
	second, err := checksum(seed, secondDisk, buf)
	if err != nil {
		return err
	}

	if len(first)%2 != 1 || len(second)%2 != 1 {
		return fmt.Errorf("checksums %s and %s must have odd length", first, second)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
