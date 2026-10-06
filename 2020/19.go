// Day 19: Monster Messages
//
// The input is grammar rules ("4: 1 5", "1: 2 3 | 3 2", "4: \"a\""), a
// blank line, then messages.
//
// Part 1 counts messages matching rule 0 exactly.
// Part 2 replaces rules 8 and 11 with looping versions, "8: 42 | 42 8" and
// "11: 42 31 | 42 11 31", and counts again.
//
// Rule 0 is "8 11", so a message matches when it is some chunks matching
// rule 42 then some matching 31: in part 1 exactly 42 42 31, in part 2 any
// a copies of 42 then b of 31 with a > b >= 1. Every other rule matches
// strings of one fixed length, and the length of 42 and 31 is the same, so
// messages are cut into chunks. Whether a rule matches at a position is
// filled in for each rule, in an order where a rule comes after the rules
// it uses (found by passes, as there are no loops among them).
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
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 16
	maxRules      = 256
	maxAlts       = 2
	maxSeq        = 3
	maxMessage    = 128
)

// rule is a literal letter, or up to two alternatives of up to three rules.
type rule struct {
	letter byte
	alts   [maxAlts][maxSeq]int
	nAlt   int
	nSeq   [maxAlts]int
	known  bool
	length int // -1 until worked out
}

// grammar holds the rules, an order to evaluate them in, and the match
// table for the current message.
type grammar struct {
	rules [maxRules]rule
	order [maxRules]int
	nOrd  int
	match [maxRules][maxMessage]bool
}

// parseRule reads one rule line.
func (g *grammar) parseRule(line string) error {
	num, body, found := strings.Cut(line, ": ")
	n, err := strconv.Atoi(num)
	if !found || err != nil || n < 0 || n >= maxRules || g.rules[n].known {
		return fmt.Errorf("malformed or repeated rule %q", line)
	}
	r := &g.rules[n]
	r.known, r.length = true, -1
	if strings.HasPrefix(body, "\"") {
		if len(body) != 3 || body[2] != '"' {
			return fmt.Errorf("malformed letter rule %q", line)
		}
		r.letter = body[1] // its length is set when it joins the order
		return nil
	}
	alts := strings.Split(body, " | ")
	if len(alts) > maxAlts {
		return fmt.Errorf("rule %q has too many alternatives", line)
	}
	for a, alt := range alts {
		seq := strings.Fields(alt)
		if len(seq) == 0 || len(seq) > maxSeq {
			return fmt.Errorf("rule %q has a bad sequence", line)
		}
		for k, s := range seq {
			if r.alts[a][k], err = strconv.Atoi(s); err != nil || r.alts[a][k] < 0 || r.alts[a][k] >= maxRules {
				return fmt.Errorf("rule %q refers to an invalid rule", line)
			}
		}
		r.nSeq[a] = len(seq)
	}
	r.nAlt = len(alts)
	return nil
}

// seqLength is the length of one alternative, or -1 if not yet known.
func (g *grammar) seqLength(r *rule, a int) int {
	total := 0
	for _, s := range r.alts[a][:r.nSeq[a]] {
		if !g.rules[s].known || g.rules[s].length < 0 {
			return -1
		}
		total += g.rules[s].length
	}
	return total
}

// lengths works out each rule's fixed length (skipping 0, 8 and 11) and
// the order to evaluate rules in.
func (g *grammar) lengths() error {
	for range maxRules {
		progress := false
		for n := range maxRules {
			r := &g.rules[n]
			if !r.known || n == 0 || n == 8 || n == 11 || (r.length >= 0 && r.nAlt > 0) {
				continue
			}
			if r.nAlt == 0 { // a letter: needs nothing before it
				if r.length < 0 {
					r.length, g.order[g.nOrd], g.nOrd, progress = 1, n, g.nOrd+1, true
				}
				continue
			}
			first := g.seqLength(r, 0)
			for a := 1; a < r.nAlt && first >= 0; a++ {
				if l := g.seqLength(r, a); l != first {
					first = -1
					if l >= 0 {
						return fmt.Errorf("rule %d has alternatives of different lengths", n)
					}
				}
			}
			if first >= 0 {
				r.length, g.order[g.nOrd], g.nOrd, progress = first, n, g.nOrd+1, true
			}
		}
		if !progress {
			break
		}
	}
	if g.nOrd == 0 {
		return errors.New("no rule could be ordered")
	}
	if g.rules[42].length < 1 || g.rules[42].length != g.rules[31].length {
		return errors.New("rules 42 and 31 do not match chunks of one length")
	}
	return nil
}

// fill works out, for each rule in order, where in msg it matches.
func (g *grammar) fill(msg string) error {
	if len(msg) > maxMessage {
		return fmt.Errorf("message of %d letters is too long", len(msg))
	}
	for _, n := range g.order[:g.nOrd] {
		r := &g.rules[n]
		for i := range len(msg) {
			ok := false
			if r.nAlt == 0 {
				ok = msg[i] == r.letter
			}
			for a := 0; a < r.nAlt && !ok && i+r.length <= len(msg); a++ {
				at, all := i, true
				for _, s := range r.alts[a][:r.nSeq[a]] {
					all = all && g.match[s][at]
					at += g.rules[s].length
				}
				ok = all
			}
			g.match[n][i] = ok && i+r.length <= len(msg)
		}
	}
	return nil
}

// check reports whether a message matches the original rule 0, and
// whether it matches the looping one.
func (g *grammar) check(msg string) (plain, looping bool, err error) {
	if msg == "" {
		return false, false, errors.New("empty message")
	}
	if err := g.fill(msg); err != nil {
		return false, false, err
	}
	size := g.rules[42].length
	if len(msg)%size != 0 {
		return false, false, nil
	}
	chunks := len(msg) / size
	// The longest run of 42s from the start, and of 31s from the end.
	lead, trail := 0, 0
	for k := 0; k < chunks && g.match[42][k*size]; k++ {
		lead++
	}
	for k := chunks - 1; k >= 0 && g.match[31][k*size]; k-- {
		trail++
	}
	plain = chunks == 3 && lead >= 2 && trail >= 1
	for b := 1; b < chunks; b++ {
		a := chunks - b
		looping = looping || (a <= lead && b <= trail && a > b)
	}
	return plain, looping, nil
}

func day19(path string) error {
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
	rules, messages, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found {
		return fmt.Errorf("%s: no blank line between rules and messages", path)
	}
	g := new(grammar)
	for _, line := range strings.Split(rules, "\n") {
		if err := g.parseRule(line); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	r0, r8, r11 := g.rules[0], g.rules[8], g.rules[11]
	if r0.nAlt != 1 || r0.nSeq[0] != 2 || r0.alts[0][0] != 8 || r0.alts[0][1] != 11 ||
		r8.nAlt != 1 || r8.nSeq[0] != 1 || r8.alts[0][0] != 42 ||
		r11.nAlt != 1 || r11.nSeq[0] != 2 || r11.alts[0][0] != 42 || r11.alts[0][1] != 31 {
		return fmt.Errorf("%s: rules 0, 8 and 11 are not 8 11, 42 and 42 31", path)
	}
	if err := g.lengths(); err != nil {
		return err
	}
	plain, looping := 0, 0
	for _, msg := range strings.Fields(messages) {
		a, b, err := g.check(msg)
		if err != nil {
			return err
		}
		if a {
			plain++
		}
		if b {
			looping++
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", plain, looping); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
