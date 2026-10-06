// Day 9: Explosives in Cyberspace
//
// The input is a compressed file. A marker "(AxB)" means: take the next A
// characters and repeat them B times. Characters inside a repeated section
// are plain data, even if they look like markers. Whitespace is ignored.
//
// Part 1 is the length of the decompressed file.
// Part 2 uses version two of the format, where markers inside a repeated
// section are expanded too.
//
// Nothing is decompressed. In version one a marker adds A*B to the length
// and skips over its A characters. In version two every character carries a
// weight - how many copies of it the output will hold - starting at 1; a
// marker multiplies the weights of the next A characters by B, contributes
// nothing itself, and the length is the sum of the weights of the plain
// characters.
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
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 20
	maxRepeat     = 1_000_000
	maxLength     = 1 << 60
)

// readFile returns the compressed file with all whitespace removed.
func readFile(path string) (string, error) {
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
	return strings.Join(strings.Fields(string(content)), ""), nil
}

// readMarker parses the marker starting at text[at], returning its span and
// repeat count and the index just past it.
func readMarker(text string, at int) (span, times, next int, err error) {
	if at < 0 || at >= len(text) || text[at] != '(' {
		return 0, 0, 0, fmt.Errorf("no marker at %d", at)
	}
	end := strings.IndexByte(text[at:], ')')
	if end < 0 {
		return 0, 0, 0, fmt.Errorf("unclosed marker at %d", at)
	}
	spanText, timesText, found := strings.Cut(text[at+1:at+end], "x")
	if !found {
		return 0, 0, 0, fmt.Errorf("malformed marker at %d", at)
	}
	if span, err = strconv.Atoi(spanText); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid marker at %d: %w", at, err)
	}
	if times, err = strconv.Atoi(timesText); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid marker at %d: %w", at, err)
	}
	next = at + end + 1
	if span < 0 || times < 0 || times > maxRepeat || next+span > len(text) {
		return 0, 0, 0, fmt.Errorf("marker at %d reaches outside the file", at)
	}
	return span, times, next, nil
}

// decompressedLength measures the file as decompressed by part 1's rules.
func decompressedLength(text string) (int, error) {
	if text == "" {
		return 0, errors.New("empty file")
	}

	length := 0
	// Each step consumes at least one character, so len(text) steps suffice.
	for i, steps := 0, 0; i < len(text); steps++ {
		if steps > len(text) {
			return 0, errors.New("decompression made no progress")
		}
		if text[i] != '(' {
			length++
			i++
			continue
		}
		span, times, next, err := readMarker(text, i)
		if err != nil {
			return 0, err
		}
		length += span * times
		i = next + span
	}
	if length < 0 || length > maxLength {
		return 0, fmt.Errorf("decompressed length %d is out of range", length)
	}
	return length, nil
}

// nestedLength measures the file by version two's rules. weights is scratch
// space with one entry per character.
func nestedLength(text string, weights []int) (int, error) {
	if text == "" || len(weights) != len(text) {
		return 0, errors.New("empty file or mismatched scratch space")
	}

	for i := range weights {
		weights[i] = 1
	}
	length := 0
	for i := 0; i < len(text); i++ {
		if text[i] != '(' {
			length += weights[i]
			continue
		}
		span, times, next, err := readMarker(text, i)
		if err != nil {
			return 0, err
		}
		for j := next; j < next+span; j++ {
			if weights[j] > maxLength/max(times, 1) {
				return 0, fmt.Errorf("decompressed length passes %d", maxLength)
			}
			weights[j] *= times
		}
		i = next - 1 // continue just past the marker
	}
	if length < 0 || length > maxLength {
		return 0, fmt.Errorf("decompressed length %d is out of range", length)
	}
	return length, nil
}

func day9(path string) error {
	text, err := readFile(path)
	if err != nil {
		return err
	}
	length, err := decompressedLength(text)
	if err != nil {
		return err
	}
	nested, err := nestedLength(text, make([]int, len(text)))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", length, nested); err != nil {
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
