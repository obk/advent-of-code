// Day 10: Factory
//
// Each line of the manual describes one machine: an indicator light diagram,
// the buttons and what each of them is wired to, and the joltages the machine
// has to end up at.
//
//	[.##.] (0,2) (1,3) (2) {3,5,4,7}
//
// Part 1 wants the fewest presses that turn the lights from all-off into the
// diagram's pattern. A button flips every light it lists, so pressing it twice
// undoes itself and only which buttons were pressed an odd number of times
// matters: the answer is the smallest subset of buttons whose flips add up to
// the pattern.
//
// Part 2 reinterprets the same buttons as counters: a press adds 1 to every
// position the button lists, and each counter has to land on its joltage
// exactly. Counting presses one at a time is hopeless - the targets run into
// the hundreds - so it is solved one binary digit at a time from the lowest
// bit up, which is what minJoltagePresses explains in detail.
//
// Both parts work from one table, built per machine: the effect of pressing
// every subset of the buttons once, bucketed by the parity pattern it leaves
// on the counters. Part 1 reads its answer straight out of the bucket for the
// light pattern. The table, its buckets and part 2's search stack are fixed
// size and allocated once, before the first machine is solved.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxButtons    = 14
	maxCounters   = 12
	maxJoltage    = 1<<14 - 1
	// Each level of part 2's search halves the targets, so no path through
	// it is deeper than the number of bits in the largest target.
	maxDepth = 15
	// A generous cap on part 2's search per machine; the real inputs need a
	// few thousand steps.
	maxSearchSteps = 50_000_000
)

// machine holds one parsed line of the manual: an indicator light diagram,
// button wiring schematics, and joltage requirements.
type machine struct {
	lights      uint             // target pattern for part 1, '#' sets the bit at its position
	buttons     [maxButtons]uint // one bitmask per button, listing the counters it affects
	numButtons  int              // how many entries of buttons are in use
	joltages    [maxCounters]int // target counter values for part 2
	numCounters int              // how many lights, and so how many counters
}

// subsetOption describes the effect of pressing some subset of buttons once
// each: how many presses that is, how much each counter increases, and the
// parity pattern those increases leave.
type subsetOption struct {
	presses int
	parity  uint
	counts  [maxCounters]int
}

// frame is one level of part 2's search: the targets still to reach, divided
// down to this level, and which options from their parity bucket are left.
type frame struct {
	remaining [maxCounters]int
	cost      int // presses spent on the levels above, already scaled
	cursor    int // next index into workspace.order to try
	end       int // one past the last index of this frame's bucket
}

// workspace holds every buffer the solver needs, sized for the largest
// machine the limits allow.
type workspace struct {
	options [1 << maxButtons]subsetOption
	order   [1 << maxButtons]int    // subset IDs, grouped by parity
	start   [1<<maxCounters + 1]int // order[start[p]:start[p+1]] has parity p
	stack   [maxDepth + 1]frame
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

// parseLights converts a diagram like "[.##.]" into a bitmask, also returning
// how many lights it has.
func parseLights(s string) (mask uint, n int, err error) {
	if len(s) < 3 || s[0] != '[' || s[len(s)-1] != ']' {
		return 0, 0, fmt.Errorf("malformed light diagram %q", s)
	}
	diagram := s[1 : len(s)-1]
	if len(diagram) > maxCounters {
		return 0, 0, fmt.Errorf("diagram %q has more than %d lights", s, maxCounters)
	}
	if strings.Trim(diagram, ".#") != "" {
		return 0, 0, fmt.Errorf("diagram %q holds something other than '.' and '#'", s)
	}

	for i := range len(diagram) {
		if diagram[i] == '#' {
			mask |= 1 << i
		}
	}
	return mask, len(diagram), nil
}

// parseButton converts a schematic like "(1,3)" into a bitmask with each
// listed position set. Every position must be below n.
func parseButton(s string, n int) (uint, error) {
	if len(s) < 3 || s[0] != '(' || s[len(s)-1] != ')' {
		return 0, fmt.Errorf("malformed button %q", s)
	}

	var mask uint
	for _, field := range strings.Split(s[1:len(s)-1], ",") {
		pos, err := strconv.Atoi(field)
		if err != nil {
			return 0, fmt.Errorf("invalid button spec %q: %w", s, err)
		}
		if pos < 0 || pos >= n {
			return 0, fmt.Errorf("button %q reaches past the %d counters", s, n)
		}
		mask |= 1 << pos
	}
	return mask, nil
}

// parseJoltages converts requirements like "{3,5,4,7}" into the machine's
// targets. There must be exactly one per counter.
func parseJoltages(s string, m *machine) error {
	if len(s) < 3 || s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("malformed joltages %q", s)
	}
	fields := strings.Split(s[1:len(s)-1], ",")
	if len(fields) != m.numCounters {
		return fmt.Errorf("%d joltages for %d counters in %q", len(fields), m.numCounters, s)
	}

	for i, field := range fields {
		v, err := strconv.Atoi(field)
		if err != nil {
			return fmt.Errorf("invalid joltage spec %q: %w", s, err)
		}
		if v < 0 || v > maxJoltage {
			return fmt.Errorf("joltage %d in %q is outside 0..%d", v, s, maxJoltage)
		}
		m.joltages[i] = v
	}
	return nil
}

// parseMachine parses one manual line of the form
// "[lights] (a,b) (c,d) ... {j1,j2,...}".
func parseMachine(line string) (machine, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return machine{}, fmt.Errorf("malformed line: %q", line)
	}
	buttons := parts[1 : len(parts)-1]
	if len(buttons) > maxButtons {
		return machine{}, fmt.Errorf("%d buttons, more than the %d allowed: %q", len(buttons), maxButtons, line)
	}

	var m machine
	var err error
	if m.lights, m.numCounters, err = parseLights(parts[0]); err != nil {
		return machine{}, err
	}
	for i, part := range buttons {
		if m.buttons[i], err = parseButton(part, m.numCounters); err != nil {
			return machine{}, err
		}
	}
	m.numButtons = len(buttons)
	if err := parseJoltages(parts[len(parts)-1], &m); err != nil {
		return machine{}, err
	}
	return m, nil
}

func parseInput(path string) ([]machine, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	machines := make([]machine, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		m, err := parseMachine(line)
		if err != nil {
			return nil, err
		}
		machines = append(machines, m)
	}
	if len(machines) == 0 {
		return nil, fmt.Errorf("%s: no machines found", path)
	}
	return machines, nil
}

// fillOptions works out, for every subset of the machine's buttons, how many
// presses it takes and what it adds to each counter.
func (w *workspace) fillOptions(m machine) error {
	if m.numButtons < 1 || m.numButtons > maxButtons {
		return fmt.Errorf("%d buttons is outside 1..%d", m.numButtons, maxButtons)
	}
	if m.numCounters < 1 || m.numCounters > maxCounters {
		return fmt.Errorf("%d counters is outside 1..%d", m.numCounters, maxCounters)
	}

	for subset := range 1 << m.numButtons {
		opt := &w.options[subset]
		*opt = subsetOption{presses: bits.OnesCount(uint(subset))}
		for i := range m.numButtons {
			if subset>>i&1 == 0 {
				continue
			}
			for j := range m.numCounters {
				opt.counts[j] += int(m.buttons[i] >> j & 1)
			}
		}
		for j := range m.numCounters {
			opt.parity |= uint(opt.counts[j]&1) << j
		}
	}
	return nil
}

// bucketOptions groups the subsets by parity with a counting sort, so that
// order[start[p]:start[p+1]] lists every subset leaving parity pattern p.
func (w *workspace) bucketOptions(m machine) error {
	subsets := 1 << m.numButtons
	if subsets > len(w.order) || m.numCounters > maxCounters {
		return fmt.Errorf("machine is larger than the workspace")
	}

	clear(w.start[:])
	for subset := range subsets {
		w.start[w.options[subset].parity+1]++
	}
	for p := 1; p < len(w.start); p++ {
		w.start[p] += w.start[p-1]
	}

	// Fill each bucket front to back, using start[p] as its write cursor and
	// shifting it back to the bucket's start afterwards.
	for subset := range subsets {
		p := w.options[subset].parity
		w.order[w.start[p]] = subset
		w.start[p]++
	}
	for p := len(w.start) - 1; p > 0; p-- {
		w.start[p] = w.start[p-1]
	}
	w.start[0] = 0

	if w.start[len(w.start)-1] != subsets {
		return fmt.Errorf("buckets hold %d subsets, expected %d", w.start[len(w.start)-1], subsets)
	}
	return nil
}

// minLightPresses returns the fewest button presses that turn the lights from
// all-off into the target pattern. A button pressed twice cancels out, so this
// is the smallest subset whose flips leave exactly the target's parity.
func (w *workspace) minLightPresses(m machine) (int, error) {
	if m.lights >= 1<<m.numCounters {
		return 0, fmt.Errorf("light pattern %b has more than %d lights", m.lights, m.numCounters)
	}

	best := -1
	for _, subset := range w.order[w.start[m.lights]:w.start[m.lights+1]] {
		if presses := w.options[subset].presses; best < 0 || presses < best {
			best = presses
		}
	}
	if best < 0 || best > m.numButtons {
		return 0, fmt.Errorf("part 1: no button presses reach lights %b", m.lights)
	}
	return best, nil
}

// pushFrame starts a new search level for the given remaining targets.
func (w *workspace) pushFrame(level int, remaining [maxCounters]int, cost int) error {
	if level < 0 || level >= len(w.stack) {
		return fmt.Errorf("search level %d is outside 0..%d", level, len(w.stack)-1)
	}

	var parity uint
	for j, r := range remaining {
		if r < 0 {
			return fmt.Errorf("counter %d was overshot", j)
		}
		parity |= uint(r&1) << j
	}
	if parity+1 >= uint(len(w.start)) {
		return fmt.Errorf("parity %b is outside the buckets", parity)
	}

	w.stack[level] = frame{
		remaining: remaining,
		cost:      cost,
		cursor:    w.start[parity],
		end:       w.start[parity+1],
	}
	return nil
}

// minJoltagePresses returns the fewest button presses needed to raise all
// counters from zero to exactly their targets, where each press of a button
// adds 1 to every counter it lists.
//
// It works one binary digit at a time. The buttons pressed an odd number of
// times are the only ones that affect the lowest bit of each counter, so we
// try every subset of buttons whose combined effect matches the targets'
// parities, subtract that subset's contribution, halve what remains, and
// carry on one level down: pressing a button at level k costs 2^k presses.
// The levels are an explicit stack, and any branch that already costs as much
// as the best complete answer found is cut off.
func (w *workspace) minJoltagePresses(m machine) (int, error) {
	if err := w.pushFrame(0, m.joltages, 0); err != nil {
		return 0, err
	}

	best, top := -1, 0
	for range maxSearchSteps {
		if top < 0 {
			break
		}
		f := &w.stack[top]
		if f.cursor >= f.end {
			top-- // every option at this level has been tried
			continue
		}
		opt := &w.options[w.order[f.cursor]]
		f.cursor++

		cost := f.cost + opt.presses<<top
		if best >= 0 && cost >= best {
			continue
		}
		var next [maxCounters]int
		feasible, done := true, true
		for j := range m.numCounters {
			left := f.remaining[j] - opt.counts[j]
			feasible = feasible && left >= 0
			next[j] = left / 2
			done = done && next[j] == 0
		}
		switch {
		case !feasible: // the subset would overshoot a counter
		case done:
			best = cost
		default:
			if err := w.pushFrame(top+1, next, cost); err != nil {
				return 0, err
			}
			top++
		}
	}

	if top >= 0 {
		return 0, fmt.Errorf("part 2: search exceeded %d steps", maxSearchSteps)
	}
	if best < 0 {
		return 0, fmt.Errorf("part 2: no solution for joltages %v", m.joltages[:m.numCounters])
	}
	// A press adds at most 1 to any counter, so no counter can need more
	// presses than were made.
	highest := 0
	for _, j := range m.joltages {
		highest = max(highest, j)
	}
	if best < highest {
		return 0, fmt.Errorf("part 2: %d presses cannot reach joltages %v", best, m.joltages[:m.numCounters])
	}
	return best, nil
}

func day10(path string) error {
	machines, err := parseInput(path)
	if err != nil {
		return err
	}
	w := new(workspace)

	lightSum, joltageSum := 0, 0
	for _, m := range machines {
		if err := w.fillOptions(m); err != nil {
			return err
		}
		if err := w.bucketOptions(m); err != nil {
			return err
		}
		lightPresses, err := w.minLightPresses(m)
		if err != nil {
			return err
		}
		joltagePresses, err := w.minJoltagePresses(m)
		if err != nil {
			return err
		}
		lightSum += lightPresses
		joltageSum += joltagePresses
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", lightSum, joltageSum); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day10(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
