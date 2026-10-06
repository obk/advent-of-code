// Day 1: Inverse Captcha
//
// The input is one long run of digits, read as a circle.
//
// Part 1 sums every digit that matches the digit after it (the last digit's
// successor being the first).
// Part 2 compares each digit with the one halfway round the circle instead.
//
// One pass with the index taken modulo the length handles the wrap-around.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
)

// readDigits returns the run of digits.
func readDigits(path string) (string, error) {
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
	digits := strings.TrimSpace(string(content))
	if digits == "" || strings.Trim(digits, "0123456789") != "" {
		return "", fmt.Errorf("%s: input is not a run of digits", path)
	}
	return digits, nil
}

// matchSum adds every digit equal to the one ahead of it by the given
// distance, going round the circle.
func matchSum(digits string, ahead int) (int, error) {
	n := len(digits)
	if n == 0 || ahead < 0 || ahead >= n {
		return 0, fmt.Errorf("cannot look %d ahead in %d digits", ahead, n)
	}
	sum := 0
	for i := range n {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, fmt.Errorf("%q is not a digit", digits[i])
		}
		if digits[i] == digits[(i+ahead)%n] {
			sum += int(digits[i] - '0')
		}
	}
	if sum > 9*n {
		return 0, fmt.Errorf("sum %d is more than %d digits can make", sum, n)
	}
	return sum, nil
}

func day1(path string) error {
	digits, err := readDigits(path)
	if err != nil {
		return err
	}
	if len(digits)%2 != 0 {
		return fmt.Errorf("%d digits have no halfway point", len(digits))
	}
	next, err := matchSum(digits, 1)
	if err != nil {
		return err
	}
	halfway, err := matchSum(digits, len(digits)/2)
	if err != nil {
		return err
	}

	if next < 0 || halfway < 0 {
		return fmt.Errorf("sums %d and %d should not be negative", next, halfway)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", next, halfway); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
