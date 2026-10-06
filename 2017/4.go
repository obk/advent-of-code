// Day 4: High-Entropy Passphrases
//
// Each line of the input is a passphrase: words of lowercase letters.
//
// Part 1 counts the passphrases with no word repeated.
// Part 2 counts those in which no two words are anagrams of each other.
//
// Passphrases are short, so every pair of words is compared directly; for
// anagrams, by their letter counts.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxWords      = 100
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

// letterCounts counts each letter in a lowercase word.
func letterCounts(word string) ([26]int, error) {
	var counts [26]int
	for i := range len(word) {
		if word[i] < 'a' || word[i] > 'z' {
			return counts, fmt.Errorf("word %q holds %q", word, word[i])
		}
		counts[word[i]-'a']++
	}
	return counts, nil
}

// check reports whether no word appears twice, and whether no two words are
// anagrams.
func check(words []string) (noRepeat, noAnagram bool, err error) {
	if len(words) == 0 || len(words) > maxWords {
		return false, false, fmt.Errorf("passphrase of %d words is outside 1..%d", len(words), maxWords)
	}
	var counts [maxWords][26]int
	for i, w := range words {
		if counts[i], err = letterCounts(w); err != nil {
			return false, false, err
		}
	}
	noRepeat, noAnagram = true, true
	for i := range words {
		for j := i + 1; j < len(words); j++ {
			noRepeat = noRepeat && words[i] != words[j]
			noAnagram = noAnagram && counts[i] != counts[j]
		}
	}
	if noAnagram && !noRepeat {
		return false, false, errors.New("a repeated word is always an anagram")
	}
	return noRepeat, noAnagram, nil
}

func day4(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	valid, strict, phrases := 0, 0, 0
	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		if strings.Trim(strings.Join(words, ""), "abcdefghijklmnopqrstuvwxyz") != "" {
			return fmt.Errorf("passphrase %q holds something other than lowercase words", line)
		}
		noRepeat, noAnagram, err := check(words)
		if err != nil {
			return err
		}
		if noRepeat {
			valid++
		}
		if noAnagram {
			strict++
		}
		phrases++
	}

	if phrases == 0 {
		return fmt.Errorf("%s: no passphrases", path)
	}
	if valid > phrases {
		return fmt.Errorf("%d valid phrases out of %d", valid, phrases)
	}
	if strict > valid {
		return fmt.Errorf("%d anagram-free phrases but only %d repeat-free", strict, valid)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", valid, strict); err != nil {
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
