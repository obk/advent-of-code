// Day 5: How About a Nice Game of Chess?
//
// The input is a door ID. Hashing the ID followed by an increasing index
// (0, 1, 2, ...) with MD5, every hash whose hex form starts with five zeroes
// is "interesting".
//
// Part 1 is the eight-character password made of the sixth hex digit of the
// first eight interesting hashes.
// Part 2 reads the sixth hex digit as a position (0-7) and the seventh as the
// character to put there, keeping only the first character for each position
// and ignoring positions past 7.
//
// Five hex zeroes are the first two bytes of the hash plus the high half of
// the third, and the sixth hex digit is the low half of that third byte. The
// ID and index are written into one reused buffer, so the search allocates
// nothing.
package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 10
	maxIDLen      = 64
	maxIndex      = 1_000_000_000
	maxDigits     = 20 // room for any int in decimal
	passwordLen   = 8
)

// readDoorID returns the door ID.
func readDoorID(path string) (string, error) {
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
	id := strings.TrimSpace(string(content))
	if id == "" || len(id) > maxIDLen {
		return "", fmt.Errorf("%s: door ID of %d bytes is outside 1..%d", path, len(id), maxIDLen)
	}
	if strings.Trim(id, "abcdefghijklmnopqrstuvwxyz") != "" {
		return "", fmt.Errorf("%s: door ID %q holds something other than lowercase letters", path, id)
	}
	return id, nil
}

// passwords finds both parts' passwords in one pass over the hashes. buf
// holds the door ID and has room for the index after it.
func passwords(buf []byte, idLen int) (first, second string, err error) {
	if idLen < 1 || cap(buf) < idLen+maxDigits {
		return "", "", fmt.Errorf("buffer of %d bytes cannot hold a %d byte ID and an index", cap(buf), idLen)
	}

	const hexDigits = "0123456789abcdef"
	var inOrder, placed [passwordLen]byte
	found, filled := 0, 0
	for index := 0; index < maxIndex && (found < passwordLen || filled < passwordLen); index++ {
		sum := md5.Sum(strconv.AppendInt(buf[:idLen], int64(index), 10))
		if sum[0] != 0 || sum[1] != 0 || sum[2]&0xF0 != 0 {
			continue
		}
		if found < passwordLen {
			inOrder[found] = hexDigits[sum[2]&0x0F]
			found++
		}
		if pos := sum[2] & 0x0F; pos < passwordLen && placed[pos] == 0 {
			placed[pos] = hexDigits[sum[3]>>4]
			filled++
		}
	}

	if found < passwordLen || filled < passwordLen {
		return "", "", fmt.Errorf("passwords incomplete below index %d", maxIndex)
	}
	return string(inOrder[:]), string(placed[:]), nil
}

func day5(path string) error {
	id, err := readDoorID(path)
	if err != nil {
		return err
	}

	buf := make([]byte, len(id), len(id)+maxDigits)
	copy(buf, id)
	first, second, err := passwords(buf, len(id))
	if err != nil {
		return err
	}

	if len(first) != passwordLen || len(second) != passwordLen {
		return fmt.Errorf("passwords %q and %q are not %d characters", first, second, passwordLen)
	}
	if strings.ContainsRune(second, 0) {
		return fmt.Errorf("password %q has an unfilled position", second)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
