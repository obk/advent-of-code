// Day 17: Chronospatial Computer
//
// The input is a 3-bit computer's registers A, B, C and its program. The
// eight instructions divide A by powers of two (into A, B or C), xor B
// with a literal or with C, set B from a value mod 8, jump if A is not
// zero, and output a value mod 8.
//
// Part 1 is the program's output, comma-separated.
// Part 2 is the lowest A that makes the program output itself.
//
// These programs loop: each pass outputs one value worked out from A's low
// bits, shifts A right by three, and jumps back while A is not zero. So
// the last output depends only on A's top three bits, the one before on
// the top six, and so on. Candidates for A are built three bits at a time
// from the end, keeping those whose output matches the program's tail.
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
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 8
	maxProgram    = 32
	maxSteps      = 100_000
	maxCandidates = 1 << 12
)

// run executes the program and writes its output into out.
func run(prog []int, a, b, c int, out []int) ([]int, error) {
	if len(prog)%2 != 0 || a < 0 {
		return nil, errors.New("program of odd length, or negative A")
	}
	out = out[:0]
	for pc, step := 0, 0; pc+1 < len(prog); step++ {
		if step == maxSteps {
			return nil, fmt.Errorf("the program ran more than %d steps", maxSteps)
		}
		op, lit := prog[pc], prog[pc+1]
		combo := [8]int{0, 1, 2, 3, a, b, c, -1}[lit]
		if combo < 0 && (op == 0 || op == 2 || op == 5 || op == 6 || op == 7) {
			return nil, errors.New("combo operand 7 is reserved")
		}
		pc += 2
		switch op {
		case 0:
			a >>= min(combo, 63)
		case 1:
			b ^= lit
		case 2:
			b = combo % 8
		case 3:
			if a != 0 {
				pc = lit
			}
		case 4:
			b ^= c
		case 5:
			if len(out) == cap(out) {
				return nil, errors.New("too much output")
			}
			out = append(out, combo%8)
		case 6:
			b = a >> min(combo, 63)
		case 7:
			c = a >> min(combo, 63)
		}
	}
	return out, nil
}

// equal reports whether two outputs match.
func equal(x, y []int) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// quine finds the lowest A for which the program outputs itself.
func quine(prog []int, cand, next []int, out []int) (int, error) {
	n, shifts := len(prog), false
	for k := 0; k+1 < n; k += 2 {
		shifts = shifts || (prog[k] == 0 && prog[k+1] == 3) // adv 3
	}
	if n < 4 || prog[n-2] != 3 || prog[n-1] != 0 || !shifts {
		return 0, errors.New("the program is not a loop shifting A by three")
	}
	cand = append(cand[:0], 0)
	for i := n - 1; i >= 0; i-- {
		next = next[:0]
		for _, base := range cand {
			for d := range 8 {
				a := base*8 + d
				o, err := run(prog, a, 0, 0, out)
				if err != nil {
					return 0, err
				}
				if equal(o, prog[i:]) && len(next) < cap(next) {
					next = append(next, a)
				}
			}
		}
		cand, next = next, cand
	}
	best := -1
	for _, a := range cand {
		if a > 0 && (best < 0 || a < best) {
			best = a
		}
	}
	if best < 0 {
		return 0, errors.New("no A reproduces the program")
	}
	return best, nil
}

func day17(path string) error {
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
	var a, b, c int
	var text string
	n, err := fmt.Sscanf(string(content), "Register A: %d\nRegister B: %d\nRegister C: %d\n\nProgram: %s", &a, &b, &c, &text)
	if err != nil || n != 4 {
		return fmt.Errorf("%s: malformed computer", path)
	}
	f := strings.Split(text, ",")
	if len(f) > maxProgram {
		return fmt.Errorf("%s: program is too long", path)
	}
	prog := make([]int, len(f))
	for i, s := range f {
		if prog[i], err = strconv.Atoi(s); err != nil || prog[i] < 0 || prog[i] > 7 {
			return fmt.Errorf("%s: invalid program value %q", path, s)
		}
	}
	out := make([]int, 0, maxProgram*4)
	o, err := run(prog, a, b, c, out)
	if err != nil {
		return err
	}
	line := make([]byte, 0, 2*cap(out))
	for i, v := range o {
		if i > 0 {
			line = append(line, ',')
		}
		line = append(line, byte('0'+v))
	}
	self, err := quine(prog, make([]int, 0, maxCandidates), make([]int, 0, maxCandidates), make([]int, 0, maxProgram*4))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", line, self); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
