// Day 11: Monkey in the Middle
//
// The input describes monkeys: their items' worry levels, an operation
// ("new = old * 19", "+ 6", "* old"), a divisibility test, and the monkeys
// to throw to when the test passes or fails.
//
// Part 1 runs 20 rounds, dividing worry by 3 after each inspection, and
// multiplies the two highest inspection counts.
// Part 2 runs 10000 rounds without dividing.
//
// Without the division by 3, only divisibility by the test numbers matters,
// so part 2 keeps worry modulo their product; that keeps it small. Each monkey's items are a fixed
// queue sized to hold every item.
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
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 12
	maxMonkeys    = 10
	maxItems      = 64
)

// monkey is one monkey's rules and items.
type monkey struct {
	items     [maxItems]int
	n         int
	op        byte // '+', '*', or 's' for squaring
	arg       int
	test      int
	onTrue    int
	onFalse   int
	inspected int
}

// field returns the text after the colon of line i, checking its label.
func field(lines []string, i int, label string) (string, error) {
	if i >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[i]), label) {
		return "", fmt.Errorf("expected %q on line %d", label, i+1)
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), label)), nil
}

// parseMonkey reads one monkey's six lines.
func parseMonkey(block string, count int) (monkey, error) {
	var m monkey
	lines := strings.Split(block, "\n")
	if len(lines) != 6 {
		return m, fmt.Errorf("monkey block has %d lines", len(lines))
	}
	items, err := field(lines, 1, "Starting items:")
	if err != nil {
		return m, err
	}
	for _, it := range strings.Split(items, ", ") {
		v, err := strconv.Atoi(it)
		if err != nil || v < 0 || m.n == maxItems {
			return m, fmt.Errorf("invalid item %q", it)
		}
		m.items[m.n], m.n = v, m.n+1
	}
	op, err := field(lines, 2, "Operation: new = old ")
	if err != nil {
		return m, err
	}
	m.op = op[0]
	if op == "* old" {
		m.op = 's'
	} else if m.arg, err = strconv.Atoi(op[2:]); err != nil || (m.op != '+' && m.op != '*') {
		return m, fmt.Errorf("invalid operation %q", op)
	}
	var texts [3]string
	for k, label := range [3]string{"Test: divisible by", "If true: throw to monkey", "If false: throw to monkey"} {
		if texts[k], err = field(lines, 3+k, label); err != nil {
			return m, err
		}
	}
	m.test, err = strconv.Atoi(texts[0])
	t, errT := strconv.Atoi(texts[1])
	f, errF := strconv.Atoi(texts[2])
	if err != nil || errT != nil || errF != nil || m.test < 1 || t < 0 || f < 0 || t >= count || f >= count {
		return m, errors.New("invalid test or targets")
	}
	m.onTrue, m.onFalse = t, f
	return m, nil
}

// business runs the rounds and multiplies the two busiest monkeys' counts.
func business(start []monkey, rounds int, relief bool) (int, error) {
	var ms [maxMonkeys]monkey
	copy(ms[:], start)
	mod := 1
	for _, m := range start {
		mod *= m.test
	}
	if mod > 1<<30 || rounds < 1 {
		return 0, errors.New("tests multiply too large, or no rounds")
	}
	for range rounds {
		for i := range start {
			m := &ms[i]
			for _, w := range m.items[:m.n] {
				switch m.op {
				case '+':
					w += m.arg
				case '*':
					w *= m.arg
				default:
					w *= w
				}
				if relief {
					w /= 3 // small enough as it is; reducing would break the division
				} else {
					w %= mod
				}
				if w < 0 {
					return 0, errors.New("worry overflowed")
				}
				to := &ms[m.onFalse]
				if w%m.test == 0 {
					to = &ms[m.onTrue]
				}
				if to.n == maxItems {
					return 0, errors.New("a monkey holds too many items")
				}
				to.items[to.n], to.n = w, to.n+1
			}
			m.inspected, m.n = m.inspected+m.n, 0
		}
	}
	first, second := 0, 0
	for _, m := range ms[:len(start)] {
		if m.inspected > first {
			first, second = m.inspected, first
		} else if m.inspected > second {
			second = m.inspected
		}
	}
	return first * second, nil
}

func day11(path string) error {
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
	blocks := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(blocks) < 2 || len(blocks) > maxMonkeys {
		return fmt.Errorf("%s: %d monkeys, expected 2..%d", path, len(blocks), maxMonkeys)
	}
	monkeys := make([]monkey, len(blocks))
	for i, b := range blocks {
		if monkeys[i], err = parseMonkey(b, len(blocks)); err != nil {
			return fmt.Errorf("%s: monkey %d: %w", path, i, err)
		}
	}
	short, err := business(monkeys, 20, true)
	if err != nil {
		return err
	}
	long, err := business(monkeys, 10_000, false)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", short, long); err != nil {
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
