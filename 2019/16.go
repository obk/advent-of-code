// Day 16: Flawed Frequency Transmission
//
// The input is a signal of digits. A phase replaces each digit i (from 1)
// with the last digit of the sum of all digits times the pattern 0, 1, 0,
// -1, each value repeated i times, skipping the very first value.
//
// Part 1 is the first eight digits after 100 phases.
// Part 2 repeats the signal 10000 times; the first seven digits give an
// offset, and the answer is the eight digits there after 100 phases.
//
// The pattern makes output i the sum of some runs of i inputs minus others,
// so with prefix sums each output costs one subtraction per run.
//
// In part 2 the offset falls in the second half of the signal, where the
// pattern for output i is 0 before i and 1 from i to the end. Each output
// is then just the sum of the inputs from it onwards, a running sum from the
// end, and only the digits from the offset on are needed.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 12
	phases        = 100
	repeats       = 10_000
)

// readSignal returns the signal's digits.
func readSignal(path string) ([]int, error) {
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
	text := strings.TrimSpace(string(content))
	if len(text) < 8 || strings.Trim(text, "0123456789") != "" {
		return nil, fmt.Errorf("%s: signal is not at least eight digits", path)
	}
	digits := make([]int, len(text))
	for i := range len(text) {
		digits[i] = int(text[i] - '0')
	}
	return digits, nil
}

// phase runs one FFT phase on signal in place, using prefix as scratch
// (one longer than the signal).
func phase(signal, prefix []int) error {
	n := len(signal)
	if n == 0 || len(prefix) != n+1 {
		return errors.New("scratch space does not match the signal")
	}
	if n > maxInputBytes {
		return fmt.Errorf("signal of %d digits is longer than %d", n, maxInputBytes)
	}
	for i, d := range signal {
		prefix[i+1] = prefix[i] + d
	}
	for i := range n {
		width := i + 1
		total := 0
		// Runs of +1 start at i, then every 4*width; runs of -1 start at
		// i + 2*width.
		for start, sign := i, 1; start < n; start, sign = start+2*width, -sign {
			end := min(start+width, n)
			total += sign * (prefix[end] - prefix[start])
		}
		signal[i] = max(total, -total) % 10
	}
	return nil
}

// message returns the eight digits at the offset of the repeated signal,
// using tail as scratch for the digits from the offset to the end.
func message(digits, tail []int) ([8]byte, error) {
	var msg [8]byte
	if len(digits) < 7 {
		return msg, errors.New("signal is too short for an offset")
	}
	offset := 0
	for _, d := range digits[:7] {
		offset = offset*10 + d
	}
	total := len(digits) * repeats
	if offset < total/2 || offset+8 > total || len(tail) != total-offset {
		return msg, fmt.Errorf("offset %d is not in the second half of %d digits", offset, total)
	}
	for i := range tail {
		tail[i] = digits[(offset+i)%len(digits)]
	}
	for range phases {
		sum := 0
		for i := len(tail) - 1; i >= 0; i-- {
			sum = (sum + tail[i]) % 10
			tail[i] = sum
		}
	}
	for i := range msg {
		msg[i] = byte('0' + tail[i])
	}
	return msg, nil
}

func day16(path string) error {
	signal, err := readSignal(path)
	if err != nil {
		return err
	}
	if len(signal) < 7 {
		return errors.New("signal is too short for an offset")
	}
	offset := 0
	for _, d := range signal[:7] {
		offset = offset*10 + d
	}
	if offset >= len(signal)*repeats {
		return fmt.Errorf("offset %d is past the repeated signal", offset)
	}
	if len(signal)*repeats-offset > maxInputBytes*repeats {
		return errors.New("the message tail is too long")
	}
	tail := make([]int, len(signal)*repeats-offset)
	msg, err := message(signal, tail)
	if err != nil {
		return err
	}
	prefix := make([]int, len(signal)+1)
	for range phases {
		if err := phase(signal, prefix); err != nil {
			return err
		}
	}
	var first [8]byte
	for i := range first {
		first[i] = byte('0' + signal[i])
	}

	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", first[:], msg[:]); err != nil {
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
