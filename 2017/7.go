// Day 7: Recursive Circus
//
// Each line of the input is a program in a tower: "fwft (72) -> ktlj, cntj,
// xhth" gives its name, its weight, and the programs standing on its disc.
//
// Part 1 is the name of the bottom program - the only one nobody holds.
// Part 2: every disc should carry sub-towers of equal total weight, and
// exactly one program has the wrong weight. What should its weight be?
//
// Programs are numbered as they are first named; a table of which ones are
// held by someone gives the bottom directly. For part 2, a breadth-first
// order from the bottom is reversed, so every program comes after all those
// above it, and tower weights are summed in that order with no recursion.
// The first unbalanced disc met in that order is the highest one; the odd
// sub-tower on it is the culprit's, and fixing the culprit's own weight by
// the difference balances it.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxWeight     = 1_000_000
)

// program is one tower member.
type program struct {
	name     string
	weight   int
	children []int
	held     bool
}

// tower holds every program, numbered by first mention.
type tower struct {
	programs []program
	ids      map[string]int
}

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

// id returns the number of the named program, adding it if new.
func (t *tower) id(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty program name")
	}
	if strings.Trim(name, "abcdefghijklmnopqrstuvwxyz") != "" {
		return 0, fmt.Errorf("program name %q is not lowercase letters", name)
	}
	if i, ok := t.ids[name]; ok {
		return i, nil
	}
	if len(t.programs) == cap(t.programs) {
		return 0, fmt.Errorf("more than %d programs", cap(t.programs))
	}
	t.ids[name] = len(t.programs)
	t.programs = append(t.programs, program{name: name, weight: -1})
	return t.ids[name], nil
}

// parseLine reads "name (weight) -> a, b, c" into the tower.
func (t *tower) parseLine(line string) error {
	head, tail, _ := strings.Cut(line, " -> ")
	name, weightText, found := strings.Cut(head, " (")
	if !found || !strings.HasSuffix(weightText, ")") {
		return fmt.Errorf("malformed program %q", line)
	}
	weight, err := strconv.Atoi(strings.TrimSuffix(weightText, ")"))
	if err != nil || weight < 0 || weight > maxWeight {
		return fmt.Errorf("invalid weight in %q", line)
	}
	i, err := t.id(name)
	if err != nil {
		return err
	}
	if t.programs[i].weight >= 0 {
		return fmt.Errorf("program %s is listed twice", name)
	}
	t.programs[i].weight = weight
	if tail == "" {
		return nil
	}
	for _, child := range strings.Split(tail, ", ") {
		c, err := t.id(child)
		if err != nil {
			return err
		}
		t.programs[i].children = append(t.programs[i].children, c)
		t.programs[c].held = true
	}
	return nil
}

func parseInput(path string) (*tower, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	t := &tower{programs: make([]program, 0, len(lines)), ids: make(map[string]int, len(lines))}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if err := t.parseLine(strings.TrimSpace(line)); err != nil {
			return nil, err
		}
	}
	for _, p := range t.programs {
		if p.weight < 0 {
			return nil, fmt.Errorf("%s: program %s is held but never described", path, p.name)
		}
	}
	return t, nil
}

// bottom returns the one program that nobody holds.
func (t *tower) bottom() (int, error) {
	if len(t.programs) == 0 {
		return 0, errors.New("empty tower")
	}
	found, count := -1, 0
	for i, p := range t.programs {
		if !p.held {
			found, count = i, count+1
		}
	}
	if count != 1 {
		return 0, fmt.Errorf("%d programs stand on nothing, expected 1", count)
	}
	return found, nil
}

// correctWeight returns the weight the one wrong program should have.
// order and total are scratch space, one entry per program.
func (t *tower) correctWeight(root int, order, total []int) (int, error) {
	n := len(t.programs)
	if root < 0 || root >= n || len(order) != n || len(total) != n {
		return 0, errors.New("bad root or scratch space")
	}
	order[0] = root
	tail := 1
	for head := 0; head < tail; head++ {
		for _, c := range t.programs[order[head]].children {
			if tail == n {
				return 0, errors.New("the tower is not a tree")
			}
			order[tail], tail = c, tail+1
		}
	}
	for k := tail - 1; k >= 0; k-- {
		p := t.programs[order[k]]
		total[order[k]] = p.weight
		for _, c := range p.children {
			total[order[k]] += total[c]
		}
		fix, odd, err := t.oddChild(order[k], total)
		if err != nil {
			return 0, err
		}
		if odd {
			return fix, nil
		}
	}
	return 0, errors.New("the tower is already balanced")
}

// oddChild checks one disc. If its sub-towers differ, it reports the weight
// that would make the odd one match.
func (t *tower) oddChild(i int, total []int) (int, bool, error) {
	if i < 0 || i >= len(t.programs) {
		return 0, false, fmt.Errorf("program %d does not exist", i)
	}
	if len(total) != len(t.programs) {
		return 0, false, errors.New("tower weights do not match the programs")
	}
	kids := t.programs[i].children
	if len(kids) < 2 {
		return 0, false, nil
	}
	for _, c := range kids {
		same := 0
		for _, d := range kids {
			if total[d] == total[c] {
				same++
			}
		}
		if same == 1 {
			// c is the odd one out; any other child shows the right total.
			other := kids[0]
			if other == c {
				other = kids[1]
			}
			if len(kids) == 2 {
				return 0, false, fmt.Errorf("disc of %s has two sub-towers and no majority", t.programs[i].name)
			}
			return t.programs[c].weight + total[other] - total[c], true, nil
		}
	}
	return 0, false, nil
}

func day7(path string) error {
	t, err := parseInput(path)
	if err != nil {
		return err
	}
	root, err := t.bottom()
	if err != nil {
		return err
	}

	n := len(t.programs)
	fix, err := t.correctWeight(root, make([]int, n), make([]int, n))
	if err != nil {
		return err
	}
	if fix < 0 {
		return fmt.Errorf("the corrected weight %d is negative", fix)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %d\n", t.programs[root].name, fix); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
