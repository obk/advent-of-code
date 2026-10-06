// Day 25: The Halting Problem
//
// The input is a Turing machine blueprint: the start state, the number of
// steps before a diagnostic checksum, and for each state and each value
// under the cursor, what to write, which way to move, and the next state.
//
// Part 1 is the checksum: the number of 1s on the tape after the given
// number of steps. (Day 25 has no second puzzle.)
//
// The rules are parsed into a table indexed by state and value. The tape is
// a fixed array as long as the cursor could possibly travel either way.
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
	maxInputBytes = 1 << 16
	maxStates     = 26
	maxSteps      = 100_000_000
)

// action is what to do for one state and value.
type action struct {
	write int
	move  int // +1 right, -1 left
	next  int
	set   bool
}

// machine is the parsed blueprint.
type machine struct {
	start, steps int
	table        [maxStates][2]action
}

// lastWord returns the last word of a line, without its trailing punctuation.
func lastWord(line string) string {
	f := strings.Fields(line)
	if len(f) == 0 {
		return ""
	}
	return strings.TrimRight(f[len(f)-1], ".:")
}

// stateID turns "A".."Z" into 0..25.
func stateID(name string) (int, error) {
	if len(name) != 1 || name[0] < 'A' || name[0] > 'Z' {
		return 0, fmt.Errorf("invalid state %q", name)
	}
	return int(name[0] - 'A'), nil
}

// parse reads the blueprint.
func parse(text string) (machine, error) {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return machine{}, errors.New("blueprint is too short")
	}
	if !strings.HasPrefix(lines[0], "Begin in state") {
		return machine{}, fmt.Errorf("blueprint starts with %q", lines[0])
	}
	var m machine
	var err error
	if m.start, err = stateID(lastWord(lines[0])); err != nil {
		return machine{}, err
	}
	stepWords := strings.Fields(lines[1])
	if len(stepWords) < 6 {
		return machine{}, fmt.Errorf("malformed step count line %q", lines[1])
	}
	if m.steps, err = strconv.Atoi(stepWords[5]); err != nil || m.steps < 0 || m.steps > maxSteps {
		return machine{}, fmt.Errorf("invalid step count in %q", lines[1])
	}
	state, value := -1, -1
	for _, line := range lines[2:] {
		w := lastWord(line)
		if w == "" {
			continue
		}
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "In state"):
			if state, err = stateID(w); err != nil {
				return machine{}, err
			}
		case strings.Contains(line, "If the current value is"):
			value = int(w[0] - '0')
		case strings.Contains(line, "Write the value"):
			if state < 0 || value < 0 || value > 1 {
				return machine{}, fmt.Errorf("rule %q comes before its state and value", line)
			}
			m.table[state][value].write, m.table[state][value].set = int(w[0]-'0'), true
		case strings.Contains(line, "Move one slot"):
			m.table[state][value].move = map[string]int{"right": 1, "left": -1}[w]
		case strings.Contains(line, "Continue with state"):
			if m.table[state][value].next, err = stateID(w); err != nil {
				return machine{}, err
			}
		}
	}
	return m, nil
}

// checksum runs the machine on tape (all zeros) and counts the 1s.
func (m *machine) checksum(tape []byte) (int, error) {
	if len(tape) < 2*m.steps+1 {
		return 0, fmt.Errorf("tape of %d cells is too short for %d steps", len(tape), m.steps)
	}
	pos, state := m.steps, m.start
	for range m.steps {
		a := m.table[state][tape[pos]]
		if !a.set || a.move == 0 {
			return 0, fmt.Errorf("state %c has no rule for value %d", 'A'+state, tape[pos])
		}
		tape[pos] = byte(a.write)
		pos += a.move
		state = a.next
		if pos < 0 || pos >= len(tape) {
			return 0, errors.New("the cursor ran off the tape")
		}
	}
	ones := 0
	for _, v := range tape {
		ones += int(v)
	}
	if ones > m.steps {
		return 0, fmt.Errorf("%d ones after only %d steps", ones, m.steps)
	}
	return ones, nil
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
	m, err := parse(string(content))
	if err != nil {
		return err
	}
	ones, err := m.checksum(make([]byte, 2*m.steps+1))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\n", ones); err != nil {
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
