// Day 18: Operation Order
//
// Each line of the input is an expression of single digits, + and *, and
// parentheses.
//
// Part 1 evaluates with + and * of equal precedence, left to right.
// Part 2 evaluates + before *.
// Each answer is the sum over all lines.
//
// Expressions are evaluated with the shunting-yard algorithm: a stack of
// values and a stack of pending operators, both fixed-size arrays.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	maxDepth      = 64
)

// machine is the two stacks.
type machine struct {
	vals   [maxDepth]int
	ops    [maxDepth]byte
	nv, no int
}

// apply pops one operator and applies it to the top two values.
func (m *machine) apply() error {
	if m.no == 0 || m.nv < 2 {
		return errors.New("an operator is missing an operand")
	}
	m.no--
	a, b := m.vals[m.nv-2], m.vals[m.nv-1]
	m.nv--
	switch m.ops[m.no] {
	case '+':
		m.vals[m.nv-1] = a + b
	case '*':
		m.vals[m.nv-1] = a * b
	default:
		return fmt.Errorf("cannot apply %q", m.ops[m.no])
	}
	return nil
}

// eval evaluates one expression; with addFirst, + binds tighter than *.
func (m *machine) eval(expr string, addFirst bool) (int, error) {
	if expr == "" {
		return 0, errors.New("empty expression")
	}
	m.nv, m.no = 0, 0
	for k := range len(expr) {
		c := expr[k]
		switch {
		case c == ' ':
		case c >= '0' && c <= '9':
			if m.nv == maxDepth {
				return 0, errors.New("expression is too deep")
			}
			m.vals[m.nv], m.nv = int(c-'0'), m.nv+1
		case c == '(':
			if m.no == maxDepth {
				return 0, errors.New("expression is too deep")
			}
			m.ops[m.no], m.no = c, m.no+1
		case c == ')':
			for i := 0; i < maxDepth && m.no > 0 && m.ops[m.no-1] != '('; i++ {
				if err := m.apply(); err != nil {
					return 0, err
				}
			}
			if m.no == 0 {
				return 0, fmt.Errorf("unmatched ')' in %q", expr)
			}
			m.no--
		case c == '+' || c == '*':
			// Apply what is pending unless it binds less tightly.
			for i := 0; i < maxDepth && m.no > 0 && m.ops[m.no-1] != '(' && !(addFirst && c == '+' && m.ops[m.no-1] == '*'); i++ {
				if err := m.apply(); err != nil {
					return 0, err
				}
			}
			if m.no == maxDepth {
				return 0, errors.New("expression is too deep")
			}
			m.ops[m.no], m.no = c, m.no+1
		default:
			return 0, fmt.Errorf("unexpected %q in %q", c, expr)
		}
	}
	for i := 0; i < maxDepth && m.no > 0; i++ {
		if err := m.apply(); err != nil {
			return 0, err
		}
	}
	if m.nv != 1 || m.no != 0 {
		return 0, fmt.Errorf("malformed expression %q", expr)
	}
	return m.vals[0], nil
}

func day18(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxLines {
		return fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	m := new(machine)
	same, addFirst := 0, 0
	for _, line := range lines {
		a, err := m.eval(line, false)
		if err != nil {
			return err
		}
		b, err := m.eval(line, true)
		if err != nil {
			return err
		}
		same, addFirst = same+a, addFirst+b
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", same, addFirst); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
