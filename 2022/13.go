// Day 13: Distress Signal
//
// The input is pairs of packets: nested lists of integers like
// [1,[2,[3]],4]. Lists compare element by element (a shorter list that
// runs out first is smaller); integers compare by value; an integer
// compared with a list is first wrapped into a one-element list.
//
// Part 1 sums the (1-based) indices of the pairs already in order.
// Part 2 adds the divider packets [[2]] and [[6]], sorts everything, and
// multiplies the dividers' positions.
//
// Packets are compared as token streams ('[', ']' and numbers) without
// building trees: when a number meets a '[', the number is wrapped in
// place in a scratch copy. Part 2 needs no sort: a divider's position is
// one more than the number of packets smaller than it.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 16
	maxTokens     = 512
	openTok       = -1
	closeTok      = -2
)

// packet is a token stream: numbers, openTok and closeTok.
type packet struct {
	tok [maxTokens]int
	n   int
}

// parse turns text into tokens.
func parse(text string) (packet, error) {
	var p packet
	depth := 0
	for k := 0; k < len(text); k++ {
		if p.n == maxTokens {
			return p, errors.New("packet is too long")
		}
		switch c := text[k]; {
		case c == '[':
			p.tok[p.n], p.n, depth = openTok, p.n+1, depth+1
		case c == ']':
			p.tok[p.n], p.n, depth = closeTok, p.n+1, depth-1
		case c == ',':
		case c >= '0' && c <= '9':
			v, last := 0, k
			for j := k; j < len(text) && text[j] >= '0' && text[j] <= '9'; j++ {
				v, last = v*10+int(text[j]-'0'), j
			}
			k = last
			p.tok[p.n], p.n = v, p.n+1
		default:
			return p, fmt.Errorf("unexpected %q in %q", c, text)
		}
		if depth < 0 {
			return p, fmt.Errorf("unbalanced %q", text)
		}
	}
	if depth != 0 || p.n == 0 || p.tok[0] != openTok {
		return p, fmt.Errorf("%q is not a list", text)
	}
	return p, nil
}

// wrap turns the number at i into the list [number].
func (p *packet) wrap(i int) error {
	if p.n+2 > maxTokens || i < 0 || i >= p.n || p.tok[i] < 0 {
		return errors.New("cannot wrap the number")
	}
	copy(p.tok[i+3:p.n+2], p.tok[i+1:p.n])
	p.tok[i+1], p.tok[i+2] = p.tok[i], closeTok
	p.tok[i] = openTok
	p.n += 2
	return nil
}

// less reports whether a comes before b, using scratch copies.
func less(a, b *packet, x, y *packet) (bool, error) {
	if a == nil || b == nil || x == nil || y == nil {
		return false, errors.New("missing packet or scratch")
	}
	*x, *y = *a, *b
	i, j := 0, 0
	for step := 0; step < 4*maxTokens; step++ {
		if i >= x.n || j >= y.n {
			return false, errors.New("ran off the end of a packet")
		}
		ta, tb := x.tok[i], y.tok[j]
		switch {
		case ta >= 0 && tb >= 0 && ta != tb:
			return ta < tb, nil
		case ta == tb:
			if i == x.n-1 && j == y.n-1 {
				return false, nil // identical
			}
			i, j = i+1, j+1
		case ta == closeTok:
			return true, nil
		case tb == closeTok:
			return false, nil
		case ta >= 0:
			if err := x.wrap(i); err != nil {
				return false, err
			}
		default:
			if err := y.wrap(j); err != nil {
				return false, err
			}
		}
	}
	return false, errors.New("comparison did not finish")
}

func day13(path string) error {
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
	lines := strings.Fields(string(content))
	if len(lines) == 0 || len(lines)%2 != 0 {
		return fmt.Errorf("%s: %d packets is not whole pairs", path, len(lines))
	}
	packets := make([]packet, len(lines))
	for i, l := range lines {
		if packets[i], err = parse(l); err != nil {
			return err
		}
	}
	var x, y packet
	ordered := 0
	for i := 0; i < len(packets); i += 2 {
		lt, err := less(&packets[i], &packets[i+1], &x, &y)
		if err != nil {
			return err
		}
		if lt {
			ordered += i/2 + 1
		}
	}
	key := 1
	for k, text := range [2]string{"[[2]]", "[[6]]"} {
		divider, err := parse(text)
		if err != nil {
			return err
		}
		pos := 1 + k // the first divider comes before the second
		for i := range packets {
			lt, err := less(&packets[i], &divider, &x, &y)
			if err != nil {
				return err
			}
			if lt {
				pos++
			}
		}
		key *= pos
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", ordered, key); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
