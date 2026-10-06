// Day 16: Permutation Promenade
//
// Sixteen programs a..p stand in a line and dance: "sX" moves the last X to
// the front, "xA/B" swaps the programs at positions A and B, and "pA/B" swaps
// the programs named A and B. The input is the dance, comma-separated.
//
// Part 1 is the order after one dance.
// Part 2 is the order after a billion dances.
//
// A dance maps every order to exactly one order, and no two orders to the
// same one, so repeating it from the starting order must come back to the
// start. Dancing until it does gives the period p, and a billion dances end
// where (a billion mod p) dances do.
//
// The moves are parsed once into a slice, and the line is a fixed array.
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
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 20
	programs      = 16
	dances        = 1_000_000_000
	maxPeriod     = 1_000_000
)

// move is one dance move: kind 's', 'x' or 'p', and its two operands.
type move struct {
	kind byte
	a, b int
}

// line is the order of the programs.
type line [programs]byte

// readMoves parses the dance.
func readMoves(path string) ([]move, error) {
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
	fields := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(fields) == 0 || fields[0] == "" {
		return nil, fmt.Errorf("%s: no dance moves", path)
	}
	moves := make([]move, len(fields))
	for i, f := range fields {
		if moves[i], err = parseMove(f); err != nil {
			return nil, err
		}
	}
	return moves, nil
}

// parseMove reads one move such as "s1", "x3/4" or "pe/b".
func parseMove(f string) (move, error) {
	if len(f) < 2 {
		return move{}, fmt.Errorf("malformed move %q", f)
	}
	m := move{kind: f[0]}
	switch m.kind {
	case 's':
		n, err := strconv.Atoi(f[1:])
		if err != nil || n < 0 || n > programs {
			return move{}, fmt.Errorf("invalid spin %q", f)
		}
		m.a = n
	case 'x':
		a, b, found := strings.Cut(f[1:], "/")
		x, errA := strconv.Atoi(a)
		y, errB := strconv.Atoi(b)
		if !found || errA != nil || errB != nil || x < 0 || y < 0 || x >= programs || y >= programs {
			return move{}, fmt.Errorf("invalid exchange %q", f)
		}
		m.a, m.b = x, y
	case 'p':
		if len(f) != 4 || f[2] != '/' || f[1] < 'a' || f[3] < 'a' || f[1] >= 'a'+programs || f[3] >= 'a'+programs {
			return move{}, fmt.Errorf("invalid partner %q", f)
		}
		m.a, m.b = int(f[1]), int(f[3])
	default:
		return move{}, fmt.Errorf("unknown move %q", f)
	}
	return m, nil
}

// dance performs every move once.
func (l *line) dance(moves []move) error {
	for _, m := range moves {
		switch m.kind {
		case 's':
			old := *l
			for i := range programs {
				l[(i+m.a)%programs] = old[i]
			}
		case 'x':
			l[m.a], l[m.b] = l[m.b], l[m.a]
		case 'p':
			i, j := -1, -1
			for k := range programs {
				switch int(l[k]) {
				case m.a:
					i = k
				case m.b:
					j = k
				}
			}
			if i < 0 || j < 0 {
				return fmt.Errorf("partners %c and %c are not both dancing", m.a, m.b)
			}
			l[i], l[j] = l[j], l[i]
		default:
			return fmt.Errorf("unknown move %c", m.kind)
		}
	}
	return nil
}

// afterMany returns the order after the given number of dances from start.
func afterMany(start line, moves []move, count int) (line, error) {
	if count < 0 || len(moves) == 0 {
		return line{}, fmt.Errorf("cannot dance %d times with %d moves", count, len(moves))
	}
	l := start
	period := 0
	for i := 1; i <= maxPeriod; i++ {
		if err := l.dance(moves); err != nil {
			return line{}, err
		}
		if l == start {
			period = i
			break
		}
	}
	if period == 0 {
		return line{}, fmt.Errorf("the dance did not repeat within %d rounds", maxPeriod)
	}
	l = start
	for range count % period {
		if err := l.dance(moves); err != nil {
			return line{}, err
		}
	}
	return l, nil
}

func day16(path string) error {
	moves, err := readMoves(path)
	if err != nil {
		return err
	}
	var start line
	for i := range programs {
		start[i] = byte('a' + i)
	}
	once := start
	if err := once.dance(moves); err != nil {
		return err
	}
	many, err := afterMany(start, moves, dances)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", once[:], many[:]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
