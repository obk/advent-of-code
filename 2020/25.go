// Day 25: Combo Breaker
//
// The input is the card's and the door's public keys. A public key is 7
// raised to a secret loop size, modulo 20201227; the encryption key is one
// public key raised to the other device's loop size.
//
// Part 1 is the encryption key. (Day 25 has no second puzzle.)
//
// The card's loop size is found by stepping powers of 7 until one equals
// its public key; the modulus bounds the search.
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
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 6
	modulus       = 20201227
	subject       = 7
)

// loopSize finds the exponent giving a public key.
func loopSize(public int) (int, error) {
	if public < 1 || public >= modulus {
		return 0, fmt.Errorf("public key %d is out of range", public)
	}
	v := 1
	for n := 1; n < modulus; n++ {
		v = v * subject % modulus
		if v == public {
			return n, nil
		}
	}
	return 0, fmt.Errorf("no loop size gives %d", public)
}

// transform raises a subject to a loop size, modulo the modulus.
func transform(base, loops int) (int, error) {
	if base < 1 || base >= modulus || loops < 0 {
		return 0, fmt.Errorf("cannot transform %d by %d", base, loops)
	}
	result := 1
	for bit := 0; bit < 64 && loops > 0; bit++ {
		if loops&1 == 1 {
			result = result * base % modulus
		}
		base, loops = base*base%modulus, loops>>1
	}
	return result, nil
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
	f := strings.Fields(string(content))
	if len(f) != 2 {
		return fmt.Errorf("%s: %d keys, expected 2", path, len(f))
	}
	card, errC := strconv.Atoi(f[0])
	door, errD := strconv.Atoi(f[1])
	if errC != nil || errD != nil {
		return fmt.Errorf("%s: keys are not numbers", path)
	}
	if card == door {
		return fmt.Errorf("%s: card and door share a public key", path)
	}
	loops, err := loopSize(card)
	if err != nil {
		return err
	}
	key, err := transform(door, loops)
	if err != nil {
		return err
	}
	// The door, with its own loop size, must reach the same key.
	doorLoops, err := loopSize(door)
	if err != nil {
		return err
	}
	if check, err := transform(card, doorLoops); err != nil || check != key {
		return fmt.Errorf("the door derives key %d, not %d (%v)", check, key, err)
	}

	if _, err := fmt.Printf("Part 1: %d\n", key); err != nil {
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
