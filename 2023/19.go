// Day 19: Aplenty
//
// The input is workflows, "px{a<2006:qkq,m>2090:A,rfg}", then parts with
// four ratings, "{x=787,m=2655,a=1222,s=2876}". A part starts at workflow
// "in" and follows the first rule it matches (or the last, catch-all
// target) until it is accepted (A) or rejected (R).
//
// Part 1 sums the ratings of the accepted parts.
// Part 2 counts how many rating combinations (each 1-4000) are accepted.
//
// Part 2 pushes ranges of ratings through the workflows on an explicit
// stack: each rule splits a range into the piece it sends on and the piece
// that carries on to the next rule.
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
	maxFlows      = 1_024
	maxRules      = 8
	accept        = -1
	reject        = -2
	maxStack      = 4 * maxFlows
)

// rule sends parts with rating cat op value to target; cat < 0 matches all.
type rule struct {
	cat, value, target int
	less               bool
}

// system is the workflows by number.
type system struct {
	ids   map[string]int
	rules [maxFlows][maxRules]rule
	count [maxFlows]int
	n     int
	start int
}

// id returns a workflow's number (or accept/reject), numbering it if new.
func (s *system) id(name string) (int, error) {
	switch name {
	case "A":
		return accept, nil
	case "R":
		return reject, nil
	}
	if i, ok := s.ids[name]; ok {
		return i, nil
	}
	if s.n == maxFlows || name == "" {
		return 0, fmt.Errorf("too many workflows, or an empty name")
	}
	s.ids[name], s.n = s.n, s.n+1
	return s.n - 1, nil
}

// addFlow parses one workflow.
func (s *system) addFlow(line string) error {
	name, body, found := strings.Cut(strings.TrimSuffix(line, "}"), "{")
	w, err := s.id(name)
	if !found || err != nil || w < 0 {
		return fmt.Errorf("malformed workflow %q", line)
	}
	parts := strings.Split(body, ",")
	if len(parts) > maxRules {
		return fmt.Errorf("workflow %q has too many rules", name)
	}
	for k, p := range parts {
		r := rule{cat: -1}
		cond, target, hasCond := strings.Cut(p, ":")
		if !hasCond {
			target = cond
		} else {
			r.cat = strings.IndexByte("xmas", cond[0])
			r.less = cond[1] == '<'
			if r.value, err = strconv.Atoi(cond[2:]); err != nil || r.cat < 0 || (cond[1] != '<' && cond[1] != '>') {
				return fmt.Errorf("malformed rule %q", p)
			}
		}
		if r.target, err = s.id(target); err != nil {
			return err
		}
		s.rules[w][k] = r
	}
	s.count[w] = len(parts)
	return nil
}

// route follows one part and reports whether it is accepted.
func (s *system) route(part [4]int) (bool, error) {
	for _, r := range part {
		if r < 1 || r > 4000 {
			return false, fmt.Errorf("rating %d is outside 1..4000", r)
		}
	}
	w := s.start
	for steps := 0; steps <= s.n; steps++ {
		if w == accept || w == reject {
			return w == accept, nil
		}
		for _, r := range s.rules[w][:s.count[w]] {
			if r.cat < 0 || (r.less && part[r.cat] < r.value) || (!r.less && part[r.cat] > r.value) {
				w = r.target
				break
			}
		}
	}
	return false, errors.New("a part went round in circles")
}

// frame is a range of parts waiting at workflow w.
type frame struct {
	w      int
	lo, hi [4]int // inclusive
}

// combinations counts accepted rating combinations.
func (s *system) combinations() (int, error) {
	var stack [maxStack]frame
	stack[0] = frame{w: s.start, lo: [4]int{1, 1, 1, 1}, hi: [4]int{4000, 4000, 4000, 4000}}
	n, total := 1, 0
	for steps := 0; n > 0 && steps < maxStack*maxRules; steps++ {
		n--
		f := stack[n]
		if f.w == reject {
			continue
		}
		if f.w == accept {
			prod := 1
			for c := range 4 {
				prod *= f.hi[c] - f.lo[c] + 1
			}
			total += prod
			continue
		}
		for _, r := range s.rules[f.w][:s.count[f.w]] {
			sent := f // the piece this rule sends on
			if r.cat >= 0 && r.less {
				sent.hi[r.cat], f.lo[r.cat] = min(f.hi[r.cat], r.value-1), max(f.lo[r.cat], r.value)
			} else if r.cat >= 0 {
				sent.lo[r.cat], f.hi[r.cat] = max(f.lo[r.cat], r.value+1), min(f.hi[r.cat], r.value)
			}
			if r.cat < 0 || sent.lo[r.cat] <= sent.hi[r.cat] {
				if n == maxStack {
					return 0, errors.New("range stack is full")
				}
				sent.w = r.target
				stack[n], n = sent, n+1
			}
			if r.cat < 0 || f.lo[r.cat] > f.hi[r.cat] {
				break // nothing carries on
			}
		}
	}
	if n > 0 {
		return 0, errors.New("the range search ran too long")
	}
	return total, nil
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
	flows, parts, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found {
		return fmt.Errorf("%s: no blank line between workflows and parts", path)
	}
	s := &system{ids: make(map[string]int, maxFlows)}
	if s.start, err = s.id("in"); err != nil {
		return err
	}
	for _, l := range strings.Split(flows, "\n") {
		if err := s.addFlow(l); err != nil {
			return err
		}
	}
	sum := 0
	for _, l := range strings.Split(parts, "\n") {
		var part [4]int
		n, err := fmt.Sscanf(l, "{x=%d,m=%d,a=%d,s=%d}", &part[0], &part[1], &part[2], &part[3])
		if err != nil || n != 4 {
			return fmt.Errorf("%s: malformed part %q", path, l)
		}
		ok, err := s.route(part)
		if err != nil {
			return err
		}
		if ok {
			sum += part[0] + part[1] + part[2] + part[3]
		}
	}
	combos, err := s.combinations()
	if err != nil {
		return err
	}

	if combos < 0 || combos > 4000*4000*4000*4000 {
		return fmt.Errorf("%d combinations is out of range", combos)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, combos); err != nil {
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
