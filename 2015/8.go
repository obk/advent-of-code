// Day 8: Matchsticks
//
// Each line of the input is a double-quoted string literal using three
// escapes: \\, \" and \x followed by two hex digits.
//
// Part 1 is the total characters of code minus the total characters the
// strings hold in memory.
// Part 2 encodes each literal again - wrapping it in new quotes and escaping
// every quote and backslash - and wants the encoded length minus the code.
//
// Neither part needs the strings built: decoding saves 1 character per \\ or
// \" and 3 per \xHH, plus the 2 quotes; encoding adds 1 per quote or backslash
// plus 2 new quotes.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
)

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// decodeSaving returns how many fewer characters the literal holds in memory
// than it takes to write.
func decodeSaving(literal string) (int, error) {
	if len(literal) < 2 || literal[0] != '"' || literal[len(literal)-1] != '"' {
		return 0, fmt.Errorf("%q is not a quoted literal", literal)
	}

	saving := 2 // the surrounding quotes
	body := literal[1 : len(literal)-1]
	for i := 0; i < len(body); i++ {
		if body[i] != '\\' {
			continue
		}
		if i+1 >= len(body) {
			return 0, fmt.Errorf("%q ends in a lone backslash", literal)
		}
		switch body[i+1] {
		case '\\', '"':
			saving++
			i++
		case 'x':
			const hexDigits = "0123456789abcdefABCDEF"
			if i+3 >= len(body) || strings.IndexByte(hexDigits, body[i+2]) < 0 || strings.IndexByte(hexDigits, body[i+3]) < 0 {
				return 0, fmt.Errorf("%q has a malformed \\x escape", literal)
			}
			saving += 3
			i += 3
		default:
			return 0, fmt.Errorf("%q has an unknown escape \\%c", literal, body[i+1])
		}
	}
	return saving, nil
}

// encodeCost returns how many more characters the literal takes once it is
// quoted and escaped again.
func encodeCost(literal string) (int, error) {
	if literal == "" {
		return 0, errors.New("empty literal")
	}
	cost := 2 + strings.Count(literal, `"`) + strings.Count(literal, `\`)
	if cost > 2+len(literal) {
		return 0, fmt.Errorf("%q cannot need %d extra characters", literal, cost)
	}
	return cost, nil
}

func day8(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	saved, added, literals := 0, 0, 0
	for _, line := range lines {
		literal := strings.TrimSpace(line)
		if literal == "" {
			continue
		}
		saving, err := decodeSaving(literal)
		if err != nil {
			return err
		}
		cost, err := encodeCost(literal)
		if err != nil {
			return err
		}
		saved += saving
		added += cost
		literals++
	}

	if literals == 0 {
		return fmt.Errorf("%s: no literals found", path)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", saved, added); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
