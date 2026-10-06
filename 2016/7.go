// Day 7: Internet Protocol Version 7
//
// Each line of the input is an IPv7 address: runs of letters, some of them in
// square brackets (the "hypernet" sequences). An ABBA is four characters of
// the form xyyx with x and y different.
//
// Part 1 counts the addresses that support TLS: an ABBA somewhere outside
// the brackets and none inside.
// Part 2 counts those that support SSL: an ABA (xyx, x and y different)
// outside the brackets with the matching BAB (yxy) inside some bracket.
//
// One pass over each address tracks whether it is inside brackets and checks
// every window of four (ABBA) and three (ABA/BAB) characters. The ABAs and
// BABs seen are marked in two 26x26 tables, indexed by their outer and inner
// letter.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/7.txt"
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

// supports reports whether the address supports TLS and whether it supports
// SSL.
func supports(addr string) (tls, ssl bool, err error) {
	if addr == "" {
		return false, false, errors.New("empty address")
	}

	var aba, bab [26][26]bool // [outer][inner] letter of xyx seen outside / inside
	inside, outsideABBA, insideABBA := false, false, false
	for i := range len(addr) {
		switch c := addr[i]; {
		case c == '[':
			if inside {
				return false, false, fmt.Errorf("nested brackets in %q", addr)
			}
			inside = true
			continue
		case c == ']':
			if !inside {
				return false, false, fmt.Errorf("unmatched bracket in %q", addr)
			}
			inside = false
			continue
		case c < 'a' || c > 'z':
			return false, false, fmt.Errorf("address %q holds %q", addr, c)
		}
		if i >= 2 && addr[i-2] == addr[i] && addr[i-1] != addr[i] && addr[i-1] >= 'a' && addr[i-1] <= 'z' {
			if inside {
				bab[addr[i-1]-'a'][addr[i]-'a'] = true // stored as the ABA it answers
			} else {
				aba[addr[i]-'a'][addr[i-1]-'a'] = true
			}
		}
		if i < 3 {
			continue
		}
		w := addr[i-3 : i+1]
		// A bracket inside the window fails the letter-equality pattern.
		if w[0] == w[3] && w[1] == w[2] && w[0] != w[1] && w[1] != '[' && w[1] != ']' {
			if inside {
				insideABBA = true
			} else {
				outsideABBA = true
			}
		}
	}

	if inside {
		return false, false, fmt.Errorf("unclosed bracket in %q", addr)
	}
	for x := range 26 {
		for y := range 26 {
			ssl = ssl || (aba[x][y] && bab[x][y])
		}
	}
	return outsideABBA && !insideABBA, ssl, nil
}

func day7(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	tls, ssl, addresses := 0, 0, 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		t, s, err := supports(line)
		if err != nil {
			return err
		}
		if t {
			tls++
		}
		if s {
			ssl++
		}
		addresses++
	}

	if addresses == 0 {
		return fmt.Errorf("%s: no addresses found", path)
	}
	if tls > addresses || ssl > addresses {
		return fmt.Errorf("%d TLS and %d SSL addresses out of %d", tls, ssl, addresses)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", tls, ssl); err != nil {
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
