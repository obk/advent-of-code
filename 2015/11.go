// Day 11: Corporate Policy
//
// The input is Santa's current password, eight lowercase letters. The next
// password is found by incrementing it like a base-26 number (a..z, with z
// carrying into the letter to its left) until it is valid: it must contain a
// straight of three increasing letters such as "abc", must not contain i, o
// or l, and must contain two different non-overlapping pairs such as "aa" and
// "bb".
//
// Part 1 is the next valid password, part 2 the one after that.
//
// Incrementing straight past a forbidden letter would test every password
// that still contains it, so on reaching one the search jumps: the forbidden
// letter is stepped on and every letter to its right reset to 'a'.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 10
	passwordLen   = 8
	maxSteps      = 100_000_000 // valid passwords are far closer together than this
)

type password [passwordLen]byte

// readPassword returns the current password.
func readPassword(path string) (password, error) {
	if path == "" {
		return password{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return password{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return password{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	text := strings.TrimSpace(string(content))
	if len(text) != passwordLen || strings.Trim(text, "abcdefghijklmnopqrstuvwxyz") != "" {
		return password{}, fmt.Errorf("%s: %q is not %d lowercase letters", path, text, passwordLen)
	}

	var p password
	copy(p[:], text)
	return p, nil
}

// increment adds one to the password, carrying from z to a. It reports false
// if the password wrapped all the way round.
func (p *password) increment() (bool, error) {
	if *p == (password{}) {
		return false, errors.New("password is not set")
	}
	for i := passwordLen - 1; i >= 0; i-- {
		if p[i] < 'a' || p[i] > 'z' {
			return false, fmt.Errorf("password holds %q", p[i])
		}
		if p[i] < 'z' {
			p[i]++
			return true, nil
		}
		p[i] = 'a'
	}
	return false, nil
}

// skipForbidden steps the leftmost i, o or l on to the next letter and resets
// everything after it to 'a', reporting whether there was one.
func (p *password) skipForbidden() (bool, error) {
	if *p == (password{}) {
		return false, errors.New("password is not set")
	}
	for i, c := range p {
		if c < 'a' || c > 'z' {
			return false, fmt.Errorf("password holds %q", c)
		}
		if c != 'i' && c != 'o' && c != 'l' {
			continue
		}
		p[i]++ // none of the three is z, so this never carries
		for j := i + 1; j < passwordLen; j++ {
			p[j] = 'a'
		}
		return true, nil
	}
	return false, nil
}

// isValid checks the straight and pair rules; forbidden letters have already
// been skipped.
func (p *password) isValid() (bool, error) {
	if strings.ContainsAny(string(p[:]), "iol") {
		return false, fmt.Errorf("password %s still holds a forbidden letter", p[:])
	}
	straight, pairs, lastPair := false, 0, -2
	for i := 1; i < passwordLen; i++ {
		if p[i] < 'a' || p[i] > 'z' {
			return false, fmt.Errorf("password holds %q", p[i])
		}
		straight = straight || (i >= 2 && p[i-2]+1 == p[i-1] && p[i-1]+1 == p[i])
		// A pair must not share a letter with the previous one.
		if p[i] == p[i-1] && i-1 > lastPair+1 {
			pairs++
			lastPair = i - 1
		}
	}
	if pairs > passwordLen/2 {
		return false, fmt.Errorf("found %d pairs in %d letters", pairs, passwordLen)
	}
	return straight && pairs >= 2, nil
}

// next returns the first valid password after p.
func next(p password) (password, error) {
	for range maxSteps {
		ok, err := p.increment()
		if err != nil {
			return password{}, err
		}
		if !ok {
			return password{}, errors.New("ran out of passwords")
		}
		// Every password skipped here holds a forbidden letter, and the one
		// jumped to holds none, so it is the next candidate to test.
		if _, err := p.skipForbidden(); err != nil {
			return password{}, err
		}
		valid, err := p.isValid()
		if err != nil {
			return password{}, err
		}
		if valid && strings.ContainsAny(string(p[:]), "iol") {
			return password{}, fmt.Errorf("accepted %s despite a forbidden letter", p[:])
		}
		if valid {
			return p, nil
		}
	}
	return password{}, fmt.Errorf("no valid password within %d steps", maxSteps)
}

func day11(path string) error {
	current, err := readPassword(path)
	if err != nil {
		return err
	}
	first, err := next(current)
	if err != nil {
		return err
	}
	second, err := next(first)
	if err != nil {
		return err
	}

	if string(first[:]) <= string(current[:]) {
		return fmt.Errorf("next password %s does not come after %s", first[:], current[:])
	}
	if string(second[:]) <= string(first[:]) {
		return fmt.Errorf("next password %s does not come after %s", second[:], first[:])
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", first[:], second[:]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day11(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
