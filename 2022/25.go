// Day 25: Full of Hot Air
//
// Each line of the input is a SNAFU number: base 5 with digits 2, 1, 0,
// - (minus one) and = (minus two).
//
// Part 1 is their sum, written in SNAFU. (Day 25 has no second puzzle.)
//
// Writing a number in SNAFU: take base-5 digits from the right; a digit of
// 3 or 4 becomes = or - with a carry of one into the next place.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 14
	maxDigits     = 27 // 5^27 is near the limit of int64
)

// parse reads a SNAFU number.
func parse(s string) (int, error) {
	if s == "" || len(s) > maxDigits {
		return 0, fmt.Errorf("invalid SNAFU number %q", s)
	}
	v := 0
	for k := range len(s) {
		d := strings.IndexByte("=-012", s[k])
		if d < 0 {
			return 0, fmt.Errorf("invalid digit %q in %q", s[k], s)
		}
		v = v*5 + d - 2
	}
	return v, nil
}

// format writes v in SNAFU into out.
func format(v int, out []byte) ([]byte, error) {
	if v < 1 || cap(out) < maxDigits {
		return nil, fmt.Errorf("cannot write %d", v)
	}
	var rev [maxDigits]byte
	n := 0
	for k := 0; k < maxDigits && v > 0; k++ {
		d := v % 5
		v /= 5
		if d > 2 {
			d -= 5
			v++
		}
		rev[n], n = "=-012"[d+2], n+1
	}
	if v != 0 {
		return nil, errors.New("too many digits")
	}
	out = out[:0]
	for i := n - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out, nil
}

func day25(path string) error {
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
	sum, lines := 0, strings.Fields(string(content))
	if len(lines) == 0 {
		return fmt.Errorf("%s: no numbers", path)
	}
	for _, f := range lines {
		v, err := parse(f)
		if err != nil {
			return err
		}
		sum += v
	}
	text, err := format(sum, make([]byte, 0, maxDigits))
	if err != nil {
		return err
	}
	if back, err := parse(string(text)); err != nil || back != sum {
		return fmt.Errorf("%s does not read back as %d", text, sum)
	}

	if _, err := fmt.Printf("Part 1: %s\n", text); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
