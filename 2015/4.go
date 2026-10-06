// Day 4: The Ideal Stocking Stuffer
//
// The input is a secret key. AdventCoins are mined by finding the lowest
// positive number that, written in decimal after the key, gives an MD5 hash
// whose hexadecimal form starts with enough zeroes.
//
// Part 1 wants five leading zeroes, part 2 six.
//
// Five hex zeroes are the first two bytes of the hash plus the high half of
// the third, so the test is on bytes rather than on a hex string. The key and
// number are written into one reused buffer, so the search allocates nothing.
// Six zeroes imply five, so part 2 starts its search where part 1 stopped.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 10
	maxKeyLen     = 64
	maxNumber     = 100_000_000 // no answer is expected past this
	maxDigits     = 20          // room for any int in decimal
)

// readKey returns the secret key.
func readKey(path string) (string, error) {
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
	key := strings.TrimSpace(string(content))
	if key == "" || len(key) > maxKeyLen {
		return "", fmt.Errorf("%s: key of %d bytes is outside 1..%d", path, len(key), maxKeyLen)
	}
	if strings.Trim(key, "abcdefghijklmnopqrstuvwxyz") != "" {
		return "", fmt.Errorf("%s: key %q holds something other than lowercase letters", path, key)
	}
	return key, nil
}

// mine returns the lowest number from start upwards whose hash, after the key
// in buf, starts with the given number of hex zeroes (5 or 6).
func mine(buf []byte, keyLen, start, zeroes int) (int, error) {
	if zeroes != 5 && zeroes != 6 {
		return 0, fmt.Errorf("can only mine for 5 or 6 zeroes, not %d", zeroes)
	}
	if keyLen < 1 || cap(buf) < keyLen+maxDigits {
		return 0, fmt.Errorf("buffer of %d bytes cannot hold a %d byte key and a number", cap(buf), keyLen)
	}
	if start < 1 || start > maxNumber {
		return 0, fmt.Errorf("search start %d is outside 1..%d", start, maxNumber)
	}

	// The third byte must be zero in full for six zeroes, in its high half
	// for five.
	thirdMask := byte(0xF0)
	if zeroes == 6 {
		thirdMask = 0xFF
	}
	for n := start; n <= maxNumber; n++ {
		data := strconv.AppendInt(buf[:keyLen], int64(n), 10)
		sum := md5.Sum(data)
		if sum[0] == 0 && sum[1] == 0 && sum[2]&thirdMask == 0 {
			return n, nil
		}
	}
	return 0, fmt.Errorf("no hash with %d zeroes below %d", zeroes, maxNumber)
}

func day4(path string) error {
	key, err := readKey(path)
	if err != nil {
		return err
	}

	buf := make([]byte, len(key), len(key)+maxDigits)
	copy(buf, key)
	five, err := mine(buf, len(key), 1, 5)
	if err != nil {
		return err
	}
	six, err := mine(buf, len(key), five, 6)
	if err != nil {
		return err
	}

	if six < five {
		return fmt.Errorf("six zeroes at %d came before five at %d", six, five)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", five, six); err != nil {
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
