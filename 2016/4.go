// Day 4: Security Through Obscurity
//
// Each line of the input is a room: an encrypted name of lowercase letters and
// dashes, a sector ID, and a checksum in brackets, as
// "aaaaa-bbb-z-y-x-123[abxyz]". A room is real when its checksum is the five
// most common letters of the name, most common first, ties alphabetical.
//
// Part 1 is the sum of the sector IDs of the real rooms.
// Part 2 decrypts the names - each letter shifted forward through the
// alphabet by the sector ID, dashes read as spaces - and wants the sector of
// the room holding the North Pole objects.
//
// The expected checksum is built by picking, five times over, the letter with
// the highest remaining count (the first in the alphabet on a tie) and
// zeroing its count.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxSector     = 1_000_000
	checksumLen   = 5
	target        = "northpole object storage"
)

// room is one parsed line.
type room struct {
	name     string
	sector   int
	checksum string
}

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

// parseRoom reads a line such as "aaaaa-bbb-z-y-x-123[abxyz]".
func parseRoom(line string) (room, error) {
	open := strings.IndexByte(line, '[')
	dash := strings.LastIndexByte(line, '-')
	if open < 0 || dash < 0 || dash > open || !strings.HasSuffix(line, "]") {
		return room{}, fmt.Errorf("malformed room %q", line)
	}
	r := room{name: line[:dash], checksum: line[open+1 : len(line)-1]}
	sector, err := strconv.Atoi(line[dash+1 : open])
	if err != nil {
		return room{}, fmt.Errorf("invalid sector in %q: %w", line, err)
	}
	if sector < 0 || sector > maxSector {
		return room{}, fmt.Errorf("sector in %q is outside 0..%d", line, maxSector)
	}
	if strings.Trim(r.name, "abcdefghijklmnopqrstuvwxyz-") != "" || len(r.checksum) != checksumLen {
		return room{}, fmt.Errorf("malformed name or checksum in %q", line)
	}
	r.sector = sector
	return r, nil
}

// isReal reports whether the room's checksum matches its name.
func isReal(r room) (bool, error) {
	if len(r.checksum) != checksumLen || r.name == "" {
		return false, fmt.Errorf("room %+v is malformed", r)
	}

	var counts [26]int
	for i := range len(r.name) {
		if r.name[i] == '-' {
			continue
		}
		if r.name[i] < 'a' || r.name[i] > 'z' {
			return false, fmt.Errorf("name %q holds %q", r.name, r.name[i])
		}
		counts[r.name[i]-'a']++
	}

	for i := range checksumLen {
		best := 0
		for letter := 1; letter < 26; letter++ {
			if counts[letter] > counts[best] {
				best = letter
			}
		}
		if r.checksum[i] != byte('a'+best) {
			return false, nil
		}
		counts[best] = -1 // used; never picked again
	}
	return true, nil
}

// decryptsTo reports whether the room's name decrypts to want, comparing
// letter by letter without building the decrypted name.
func decryptsTo(r room, want string) (bool, error) {
	if r.sector < 0 || want == "" {
		return false, fmt.Errorf("cannot decrypt room %+v to %q", r, want)
	}
	if len(r.name) != len(want) {
		return false, nil
	}
	for i := range len(r.name) {
		c := r.name[i]
		switch {
		case c == '-':
			c = ' '
		case c >= 'a' && c <= 'z':
			c = byte('a' + (int(c-'a')+r.sector)%26)
		default:
			return false, fmt.Errorf("name %q holds %q", r.name, c)
		}
		if c != want[i] {
			return false, nil
		}
	}
	return true, nil
}

func day4(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	sum, rooms, storage := 0, 0, -1
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r, err := parseRoom(line)
		if err != nil {
			return err
		}
		real, err := isReal(r)
		if err != nil {
			return err
		}
		if real {
			sum += r.sector
		}
		found, err := decryptsTo(r, target)
		if err != nil {
			return err
		}
		if found && real {
			storage = r.sector
		}
		rooms++
	}

	if rooms == 0 {
		return fmt.Errorf("%s: no rooms found", path)
	}
	if storage < 0 {
		return fmt.Errorf("part 2: no real room decrypts to %q", target)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, storage); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
