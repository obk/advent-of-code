// Day 24: Arithmetic Logic Unit
//
// The input is MONAD, an ALU program (inp, add, mul, div, mod, eql on
// registers w, x, y, z) that checks a 14-digit model number, digits 1-9,
// leaving z = 0 if it is valid.
//
// Part 1 is the largest valid model number.
// Part 2 is the smallest.
//
// MONAD is 14 blocks of 18 instructions, one per digit, differing only in
// three constants: "div z D" (1 or 26), "add x A" and "add y B". z works
// as a stack of base-26 digits: a D=1 block pushes digit+B; a D=26 block
// pops, and must find its digit equal to the popped value plus A, or it
// pushes again and z can never return to 0. So each D=26 block pairs with
// an earlier D=1 block, fixing the difference of their two digits, and the
// largest and smallest numbers follow pair by pair. Each answer is then
// checked by running the whole program on it.
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
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 12
	digits        = 14
	blockLen      = 18
)

// instruction is one ALU instruction; b is a register index or, if
// literal, the value.
type instruction struct {
	op      string
	a, b    int
	literal bool
}

// parse reads the program.
func parse(lines []string) ([]instruction, error) {
	prog := make([]instruction, len(lines))
	for i, l := range lines {
		f := strings.Fields(l)
		if len(f) < 2 || len(f) > 3 || len(f[1]) != 1 || f[1][0] < 'w' || f[1][0] > 'z' {
			return nil, fmt.Errorf("malformed instruction %q", l)
		}
		in := instruction{op: f[0], a: int(f[1][0] - 'w')}
		if len(f) == 3 {
			if f[2] >= "w" && f[2] <= "z" && len(f[2]) == 1 {
				in.b = int(f[2][0] - 'w')
			} else {
				v, err := strconv.Atoi(f[2])
				if err != nil {
					return nil, fmt.Errorf("invalid operand in %q", l)
				}
				in.b, in.literal = v, true
			}
		}
		prog[i] = in
	}
	return prog, nil
}

// run executes the program on a model number and returns z.
func run(prog []instruction, model [digits]int) (int, error) {
	var reg [4]int
	next := 0
	for _, in := range prog {
		b := in.b
		if !in.literal {
			b = reg[in.b]
		}
		switch in.op {
		case "inp":
			if next == digits {
				return 0, errors.New("the program reads more than 14 digits")
			}
			reg[in.a], next = model[next], next+1
		case "add":
			reg[in.a] += b
		case "mul":
			reg[in.a] *= b
		case "div", "mod":
			if b == 0 || (in.op == "mod" && (reg[in.a] < 0 || b < 0)) {
				return 0, fmt.Errorf("invalid %s by %d", in.op, b)
			}
			if in.op == "div" {
				reg[in.a] /= b
			} else {
				reg[in.a] %= b
			}
		case "eql":
			if reg[in.a] == b {
				reg[in.a] = 1
			} else {
				reg[in.a] = 0
			}
		default:
			return 0, fmt.Errorf("unknown instruction %q", in.op)
		}
	}
	if next != digits {
		return 0, fmt.Errorf("the program read %d digits, not %d", next, digits)
	}
	return reg[3], nil
}

// constant reads the literal of the instruction at line i, checking its
// text.
func constant(prog []instruction, lines []string, i int, prefix string) (int, error) {
	if i >= len(lines) || !strings.HasPrefix(lines[i], prefix) || !prog[i].literal {
		return 0, fmt.Errorf("line %d is not %q", i+1, prefix+"N")
	}
	return prog[i].b, nil
}

// solve finds the largest and smallest valid model numbers.
func solve(prog []instruction, lines []string) (hi, lo [digits]int, err error) {
	var stack [digits][2]int // block index and its B
	n := 0
	for d := range digits {
		base := d * blockLen
		div, err := constant(prog, lines, base+4, "div z ")
		if err != nil {
			return hi, lo, err
		}
		a, err := constant(prog, lines, base+5, "add x ")
		if err != nil {
			return hi, lo, err
		}
		b, err := constant(prog, lines, base+15, "add y ")
		if err != nil {
			return hi, lo, err
		}
		switch {
		case div == 1 && n < digits:
			stack[n], n = [2]int{d, b}, n+1
		case div == 26 && n > 0:
			n--
			j, diff := stack[n][0], stack[n][1]+a // digit d = digit j + diff
			hi[j], hi[d] = min(9, 9-diff), min(9, 9+diff)
			lo[j], lo[d] = max(1, 1-diff), max(1, 1+diff)
		default:
			return hi, lo, fmt.Errorf("block %d divides by %d with %d pushed", d, div, n)
		}
	}
	if n != 0 {
		return hi, lo, fmt.Errorf("%d pushes are never popped", n)
	}
	return hi, lo, nil
}

// join turns digits into a number, checking that each is 1-9.
func join(model [digits]int) (int, error) {
	v := 0
	for _, d := range model {
		if d < 1 || d > 9 {
			return 0, fmt.Errorf("digit %d is not 1-9", d)
		}
		v = v*10 + d
	}
	if v < 11_111_111_111_111 {
		return 0, fmt.Errorf("%d is not fourteen digits", v)
	}
	return v, nil
}

func day24(path string) error {
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
	if len(lines) != digits*blockLen {
		return fmt.Errorf("%s: %d instructions, expected %d", path, len(lines), digits*blockLen)
	}
	prog, err := parse(lines)
	if err != nil {
		return err
	}
	hi, lo, err := solve(prog, lines)
	if err != nil {
		return err
	}
	var answers [2]int
	for k, model := range [2][digits]int{hi, lo} {
		z, err := run(prog, model)
		if err != nil {
			return err
		}
		if z != 0 {
			return fmt.Errorf("MONAD rejects %v (z = %d)", model, z)
		}
		if answers[k], err = join(model); err != nil {
			return err
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
