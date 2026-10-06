// Day 21: Monkey Math
//
// Each line of the input is a monkey's job: a number ("dbpl: 5") or an
// operation on two other monkeys ("root: pppw + sjmn").
//
// Part 1 is the number root yells.
// Part 2 makes root check its two numbers for equality and asks what
// humn must yell for them to be equal.
//
// Values are worked out in passes until every monkey is known (there are
// no cycles). For part 2, exactly one side of each operation on the way
// from root to humn depends on humn; walking down that path, each
// operation is undone to find what the dependent side must be. The answer
// is checked by evaluating again with it.
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
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 16
	maxMonkeys    = 4_096
)

// troop is every monkey's job and the evaluation's buffers.
type troop struct {
	op          []byte // '+', '-', '*', '/', or 'n' for a number
	num         []int
	left, right []int
	value       []int
	known, dep  []bool
	root, humn  int
}

// load parses the jobs.
func load(path string) (*troop, error) {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	n := len(lines)
	if n > maxMonkeys {
		return nil, fmt.Errorf("%s: more than %d monkeys", path, maxMonkeys)
	}
	t := &troop{op: make([]byte, n), num: make([]int, n), left: make([]int, n), right: make([]int, n),
		value: make([]int, n), known: make([]bool, n), dep: make([]bool, n)}
	ids := make(map[string]int, n)
	for i, l := range lines {
		ids[l[:min(4, len(l))]] = i
	}
	r, okR := ids["root"]
	h, okH := ids["humn"]
	if !okR || !okH {
		return nil, fmt.Errorf("%s: no root or no humn", path)
	}
	t.root, t.humn = r, h
	for i, l := range lines {
		f := strings.Fields(l)
		switch len(f) {
		case 2:
			if t.num[i], err = strconv.Atoi(f[1]); err != nil {
				return nil, fmt.Errorf("%s: invalid number in %q", path, l)
			}
			t.op[i] = 'n'
		case 4:
			a, okA := ids[f[1]]
			b, okB := ids[f[3]]
			if !okA || !okB || len(f[2]) != 1 || !strings.Contains("+-*/", f[2]) {
				return nil, fmt.Errorf("%s: malformed job %q", path, l)
			}
			if a == i || b == i {
				return nil, fmt.Errorf("%s: monkey %q waits on itself", path, l)
			}
			t.op[i], t.left[i], t.right[i] = f[2][0], a, b
		default:
			return nil, fmt.Errorf("%s: malformed job %q", path, l)
		}
	}
	return t, nil
}

// apply works out a op b.
func apply(op byte, a, b int) (int, error) {
	switch op {
	case '+':
		return a + b, nil
	case '-':
		return a - b, nil
	case '*':
		if a != 0 && (a*b)/a != b {
			return 0, fmt.Errorf("%d * %d overflows", a, b)
		}
		return a * b, nil
	case '/':
		if b == 0 || a%b != 0 {
			return 0, fmt.Errorf("%d / %d is not exact", a, b)
		}
		return a / b, nil
	}
	return 0, fmt.Errorf("unknown operation %q", op)
}

// evaluate works out every value, with humn yelling h, and marks which
// monkeys depend on humn.
func (t *troop) evaluate(h int) error {
	n := len(t.op)
	for i := range n {
		t.known[i], t.dep[i] = t.op[i] == 'n', i == t.humn
		t.value[i] = t.num[i]
	}
	t.value[t.humn] = h
	for pass := 0; pass <= n; pass++ {
		changed := false
		for i := range n {
			a, b := t.left[i], t.right[i]
			if t.known[i] || !t.known[a] || !t.known[b] {
				continue
			}
			v, err := apply(t.op[i], t.value[a], t.value[b])
			if err != nil {
				return err
			}
			t.value[i], t.known[i], t.dep[i], changed = v, true, t.dep[a] || t.dep[b], true
		}
		if !changed {
			break
		}
	}
	if !t.known[t.root] {
		return errors.New("root's value depends on a cycle")
	}
	return nil
}

// solve finds what humn must yell for root's two numbers to be equal.
func (t *troop) solve() (int, error) {
	if err := t.evaluate(t.num[t.humn]); err != nil {
		return 0, err
	}
	at, target := t.root, 0
	for steps := 0; steps <= len(t.op); steps++ {
		if at == t.humn {
			return target, nil
		}
		a, b := t.left[at], t.right[at]
		if t.op[at] == 'n' || t.dep[a] == t.dep[b] {
			return 0, fmt.Errorf("monkey %d does not have exactly one side depending on humn", at)
		}
		va, vb := t.value[a], t.value[b]
		var err error
		switch {
		case at == t.root && t.dep[a]:
			target = vb
		case at == t.root:
			target = va
		case t.dep[a] && t.op[at] == '+':
			target = target - vb
		case t.dep[a] && t.op[at] == '-':
			target = target + vb
		case t.dep[a] && t.op[at] == '*':
			target, err = apply('/', target, vb)
		case t.dep[a]:
			target = target * vb
		case t.op[at] == '+':
			target = target - va
		case t.op[at] == '-':
			target = va - target
		case t.op[at] == '*':
			target, err = apply('/', target, va)
		default:
			target, err = apply('/', va, target)
		}
		if err != nil {
			return 0, err
		}
		at = b
		if t.dep[a] {
			at = a
		}
	}
	return 0, errors.New("the path to humn is too long")
}

func day21(path string) error {
	t, err := load(path)
	if err != nil {
		return err
	}
	if err := t.evaluate(t.num[t.humn]); err != nil {
		return err
	}
	root := t.value[t.root]
	h, err := t.solve()
	if err != nil {
		return err
	}
	if err := t.evaluate(h); err != nil {
		return err
	}
	if t.value[t.left[t.root]] != t.value[t.right[t.root]] {
		return fmt.Errorf("with humn = %d, root still sees %d and %d", h, t.value[t.left[t.root]], t.value[t.right[t.root]])
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", root, h); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
