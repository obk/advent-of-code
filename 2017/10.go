// Day 10: Knot Hash
//
// A circle of marks 0..255 is "knotted" by a list of lengths: for each
// length, the span of that many marks from the current position is
// reversed, then the position moves on by the length plus a skip size that
// grows by one each time.
//
// Part 1 reads the input as comma-separated lengths, does one round, and
// multiplies the first two marks.
// Part 2 is the full Knot Hash: the input's bytes, followed by 17, 31, 73,
// 47, 23, are the lengths; 64 rounds are run keeping position and skip; and
// each block of 16 marks is XORed into one byte of a 16-byte hash, written
// in hex.
//
// The circle is a fixed array of 256 bytes; a span that wraps round is
// reversed by swapping from both ends with indices taken modulo 256.
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
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 10
	marks         = 256
	maxLengths    = 256
)

// knot is the circle and where the knotting has got to.
type knot struct {
	list      [marks]int
	pos, skip int
}

// readInput returns the input line.
func readInput(path string) (string, error) {
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
	text := strings.TrimSpace(string(content))
	if strings.ContainsAny(text, "\r\n") {
		return "", fmt.Errorf("%s: input should be a single line", path)
	}
	return text, nil
}

// parseLengths reads the comma-separated lengths for part 1.
func parseLengths(text string, out []int) ([]int, error) {
	if text == "" {
		return nil, errors.New("no lengths given")
	}
	fields := strings.Split(text, ",")
	if len(fields) > cap(out) {
		return nil, fmt.Errorf("%d lengths, more than %d", len(fields), cap(out))
	}
	out = out[:0]
	for _, f := range fields {
		n, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			return nil, fmt.Errorf("invalid length %q: %w", f, err)
		}
		if n < 0 || n > marks {
			return nil, fmt.Errorf("length %d is outside 0..%d", n, marks)
		}
		out = append(out, n)
	}
	return out, nil
}

// reset puts the circle back to 0..255 with position and skip at 0.
func (k *knot) reset() error {
	if k == nil {
		return errors.New("no knot to reset")
	}
	for i := range marks {
		k.list[i] = i
	}
	k.pos, k.skip = 0, 0
	return nil
}

// round applies every length once.
func (k *knot) round(lengths []int) error {
	if len(lengths) == 0 {
		return errors.New("no lengths")
	}
	if k.pos < 0 || k.pos >= marks || k.skip < 0 {
		return fmt.Errorf("position %d or skip %d is out of range", k.pos, k.skip)
	}
	for _, n := range lengths {
		if n < 0 || n > marks {
			return fmt.Errorf("length %d is outside 0..%d", n, marks)
		}
		for i, j := k.pos, k.pos+n-1; i < j; i, j = i+1, j-1 {
			k.list[i%marks], k.list[j%marks] = k.list[j%marks], k.list[i%marks]
		}
		k.pos = (k.pos + n + k.skip) % marks
		k.skip++
	}
	return nil
}

// knotHash returns the full Knot Hash of text as 32 hex digits.
func knotHash(text string, lengths []int) (string, error) {
	suffix := [5]int{17, 31, 73, 47, 23}
	if len(text)+len(suffix) > cap(lengths) {
		return "", fmt.Errorf("text of %d bytes is too long to hash", len(text))
	}
	lengths = lengths[:0]
	for i := range len(text) {
		lengths = append(lengths, int(text[i]))
	}
	lengths = append(lengths, suffix[:]...)

	var k knot
	if err := k.reset(); err != nil {
		return "", err
	}
	for range 64 {
		if err := k.round(lengths); err != nil {
			return "", err
		}
	}
	total := 0
	for _, v := range k.list {
		total += v
	}
	if total != marks*(marks-1)/2 {
		return "", errors.New("the marks are no longer a permutation of 0..255")
	}
	const hexDigits = "0123456789abcdef"
	var out [32]byte
	for block := range 16 {
		x := 0
		for i := range 16 {
			x ^= k.list[block*16+i]
		}
		if x < 0 || x > 255 {
			return "", fmt.Errorf("dense hash byte %d is %d", block, x)
		}
		out[2*block], out[2*block+1] = hexDigits[x>>4], hexDigits[x&15]
	}
	return string(out[:]), nil
}

func day10(path string) error {
	text, err := readInput(path)
	if err != nil {
		return err
	}
	lengths, err := parseLengths(text, make([]int, 0, maxLengths))
	if err != nil {
		return err
	}

	var k knot
	if err := k.reset(); err != nil {
		return err
	}
	if err := k.round(lengths); err != nil {
		return err
	}
	hash, err := knotHash(text, make([]int, 0, maxInputBytes+5))
	if err != nil {
		return err
	}
	if len(hash) != 32 {
		return fmt.Errorf("knot hash %q is not 32 hex digits", hash)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", k.list[0]*k.list[1], hash); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
