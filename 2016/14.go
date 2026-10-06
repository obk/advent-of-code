// Day 14: One-Time Pad
//
// The input is a salt. The MD5 hash of the salt followed by an index (0, 1,
// 2, ...), as lowercase hex, is a key when it contains a run of three equal
// characters and one of the next 1000 hashes contains five of that same
// character in a row. Only the first run of three in a hash counts.
//
// Part 1 is the index that produces the 64th key.
// Part 2 stretches every hash: its hex form is hashed again, 2016 more times.
//
// Each hash is reduced to what the test needs - the character of its first
// triple (or -1) and a 16-bit set of the characters it has five in a row -
// and kept in a ring buffer covering the 1001 most recent indices, so every
// hash is computed exactly once.
package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 10
	maxSaltLen    = 64
	maxIndex      = 10_000_000
	window        = 1000
	keysWanted    = 64
	stretching    = 2016
)

// summary is what the key test needs from one hash.
type summary struct {
	triple int    // hex digit of the first run of three, or -1
	fives  uint16 // bit d set if digit d appears five times in a row
}

// readSalt returns the salt.
func readSalt(path string) (string, error) {
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
	salt := strings.TrimSpace(string(content))
	if salt == "" || len(salt) > maxSaltLen {
		return "", fmt.Errorf("%s: salt of %d bytes is outside 1..%d", path, len(salt), maxSaltLen)
	}
	if strings.ContainsAny(salt, " \t\r\n") {
		return "", fmt.Errorf("%s: salt %q holds whitespace", path, salt)
	}
	return salt, nil
}

// summarize hashes the salt (in buf) with an index, then rehashes the hex
// form extra times, and records the runs in the result.
func summarize(buf []byte, saltLen, index, extra int) (summary, error) {
	if saltLen < 1 || cap(buf) < saltLen+20 {
		return summary{}, fmt.Errorf("buffer of %d bytes cannot hold a %d byte salt and an index", cap(buf), saltLen)
	}
	if extra < 0 || extra > stretching {
		return summary{}, fmt.Errorf("%d extra hashings is outside 0..%d", extra, stretching)
	}
	sum := md5.Sum(strconv.AppendInt(buf[:saltLen], int64(index), 10))
	var hexForm [2 * md5.Size]byte
	for range extra {
		_ = hex.Encode(hexForm[:], sum[:]) // always writes 2*md5.Size bytes
		sum = md5.Sum(hexForm[:])
	}

	var digits [32]int
	for i, b := range sum {
		digits[2*i], digits[2*i+1] = int(b>>4), int(b&0x0F)
	}
	s := summary{triple: -1}
	run := 1
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[i-1] {
			run = 1
			continue
		}
		run++
		if run == 3 && s.triple < 0 {
			s.triple = digits[i]
		}
		if run == 5 {
			s.fives |= 1 << digits[i]
		}
	}
	if s.triple < -1 || s.triple > 15 {
		return summary{}, fmt.Errorf("triple digit %d is not hex", s.triple)
	}
	return s, nil
}

// sixtyFourthKey returns the index of the 64th key. ring is scratch space for
// window+1 summaries.
func sixtyFourthKey(buf []byte, saltLen, extra int, ring []summary) (int, error) {
	if len(ring) != window+1 {
		return 0, fmt.Errorf("ring holds %d summaries, need %d", len(ring), window+1)
	}
	if extra < 0 || extra > stretching {
		return 0, fmt.Errorf("%d extra hashings is outside 0..%d", extra, stretching)
	}

	// Fill the ring with indices 0..window first.
	for i := range ring {
		s, err := summarize(buf, saltLen, i, extra)
		if err != nil {
			return 0, err
		}
		ring[i] = s
	}
	keys := 0
	for index := 0; index < maxIndex; index++ {
		s := ring[index%len(ring)]
		isKey := false
		for ahead := 1; ahead <= window && s.triple >= 0 && !isKey; ahead++ {
			isKey = ring[(index+ahead)%len(ring)].fives>>s.triple&1 == 1
		}
		if isKey {
			if keys++; keys == keysWanted {
				return index, nil
			}
		}
		// Index is done with; its slot takes the hash window+1 ahead.
		next, err := summarize(buf, saltLen, index+window+1, extra)
		if err != nil {
			return 0, err
		}
		ring[index%len(ring)] = next
	}
	return 0, fmt.Errorf("only %d keys below index %d", keys, maxIndex)
}

func day14(path string) error {
	salt, err := readSalt(path)
	if err != nil {
		return err
	}

	buf := make([]byte, len(salt), len(salt)+20)
	copy(buf, salt)
	ring := make([]summary, window+1)
	index, err := sixtyFourthKey(buf, len(salt), 0, ring)
	if err != nil {
		return err
	}
	stretched, err := sixtyFourthKey(buf, len(salt), stretching, ring)
	if err != nil {
		return err
	}

	if index < keysWanted-1 || stretched < keysWanted-1 {
		return fmt.Errorf("64 keys cannot come before index 63 (%d, %d)", index, stretched)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", index, stretched); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
