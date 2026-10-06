// Day 20: Pulse Propagation
//
// Each line of the input is a module and its outputs: the broadcaster
// (passes pulses on), flip-flops '%' (toggle on a low pulse, sending their
// new state; ignore high), and conjunctions '&' (remember each input's
// last pulse; send low if all are high, else high). Pressing the button
// sends a low pulse to the broadcaster; pulses are handled in order.
//
// Part 1 multiplies the low and high pulse counts over 1000 presses.
// Part 2 is the fewest presses to send a low pulse to rx.
//
// rx is fed by one conjunction, which sends low once all of its inputs
// have just sent it high. In these inputs each of those goes high on its
// own fixed cycle of presses, so the answer is the least common multiple
// of the press counts at which each first does.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 12
	maxModules    = 64
	maxOutputs    = 8
	maxQueue      = 1 << 10
	maxPresses    = 100_000
)

// module is one module's kind, outputs and state.
type module struct {
	kind   byte // 'b' broadcaster, '%' flip-flop, '&' conjunction, 0 none
	outs   [maxOutputs]int
	nOut   int
	on     bool
	inputs uint64 // conjunction: which modules feed it
	high   uint64 // conjunction: which inputs last sent high
}

// network is the modules and the pulse queue.
type network struct {
	m         [maxModules]module
	n         int
	ids       map[string]int
	queue     [maxQueue][3]int // from, to, high (0/1)
	broadcast int
}

// id returns a module's number, numbering it if new.
func (nw *network) id(name string) (int, error) {
	if i, ok := nw.ids[name]; ok {
		return i, nil
	}
	if nw.n == maxModules || name == "" {
		return 0, errors.New("too many modules, or an empty name")
	}
	nw.ids[name], nw.n = nw.n, nw.n+1
	return nw.n - 1, nil
}

// load parses the modules.
func (nw *network) load(text string) error {
	nw.ids = make(map[string]int, maxModules)
	for _, l := range strings.Split(strings.TrimSpace(text), "\n") {
		name, outs, ok := strings.Cut(l, " -> ")
		if !ok || name == "" {
			return fmt.Errorf("malformed module %q", l)
		}
		kind := name[0]
		if kind == '%' || kind == '&' {
			name = name[1:]
		} else if name == "broadcaster" {
			kind = 'b'
		} else {
			return fmt.Errorf("unknown module %q", name)
		}
		i, err := nw.id(name)
		if err != nil {
			return err
		}
		nw.m[i].kind = kind
		for _, o := range strings.Split(outs, ", ") {
			j, err := nw.id(o)
			if err != nil || nw.m[i].nOut == maxOutputs {
				return fmt.Errorf("bad output %q of %q", o, name)
			}
			nw.m[i].outs[nw.m[i].nOut], nw.m[i].nOut = j, nw.m[i].nOut+1
			nw.m[j].inputs |= 1 << i
		}
	}
	b, ok := nw.ids["broadcaster"]
	if !ok {
		return errors.New("no broadcaster")
	}
	nw.broadcast = b
	return nil
}

// press pushes the button once, counting pulses, and returns which
// modules sent a high pulse to watch.
func (nw *network) press(counts *[2]int, watch int) (uint64, error) {
	if counts == nil || nw.broadcast < 0 || nw.broadcast >= nw.n {
		return 0, errors.New("no counts, or no broadcaster")
	}
	nw.queue[0] = [3]int{-1, nw.broadcast, 0}
	tail := 1
	var sentHigh uint64
	for head := 0; head < tail; head++ {
		if head > 64*maxQueue {
			return 0, errors.New("one press sent too many pulses")
		}
		from, to, high := nw.queue[head%maxQueue][0], nw.queue[head%maxQueue][1], nw.queue[head%maxQueue][2]
		counts[high]++
		if to == watch && high == 1 {
			sentHigh |= 1 << from
		}
		m := &nw.m[to]
		out := -1
		switch m.kind {
		case 'b':
			out = high
		case '%':
			if high == 0 {
				m.on = !m.on
				out = 0
				if m.on {
					out = 1
				}
			}
		case '&':
			m.high = m.high&^(1<<from) | uint64(high)<<from
			out = 1
			if m.high == m.inputs {
				out = 0
			}
		}
		for _, o := range m.outs[:m.nOut] {
			if out < 0 {
				break
			}
			if tail-head >= maxQueue {
				return 0, errors.New("pulse queue is full")
			}
			nw.queue[tail%maxQueue], tail = [3]int{to, o, out}, tail+1
		}
	}
	return sentHigh, nil
}

// lcm is the least common multiple of two positive numbers.
func lcm(a, b int) (int, error) {
	if a < 1 || b < 1 {
		return 0, fmt.Errorf("lcm of %d and %d", a, b)
	}
	x, y := a, b
	for k := 0; k < 128 && y != 0; k++ {
		x, y = y, x%y
	}
	if a/x > (1<<62)/b {
		return 0, errors.New("the lcm overflows")
	}
	return a / x * b, nil
}

// firstLow finds the presses until rx gets a low pulse.
func (nw *network) firstLow() (int, error) {
	rx, ok := nw.ids["rx"]
	if !ok {
		return 0, errors.New("no rx module")
	}
	feeder := -1
	for i := range nw.n {
		for _, o := range nw.m[i].outs[:nw.m[i].nOut] {
			if o == rx {
				feeder = i
			}
		}
	}
	if feeder < 0 || nw.m[feeder].kind != '&' || nw.m[rx].inputs != 1<<feeder {
		return 0, errors.New("rx is not fed by a single conjunction")
	}
	var counts [2]int
	need, all := nw.m[feeder].inputs, 1
	for p := 1; p <= maxPresses && need != 0; p++ {
		sent, err := nw.press(&counts, feeder)
		if err != nil {
			return 0, err
		}
		for i := range nw.n {
			if sent&need&(1<<i) != 0 {
				need &^= 1 << i
				if all, err = lcm(all, p); err != nil {
					return 0, err
				}
			}
		}
	}
	if need != 0 {
		return 0, fmt.Errorf("some inputs of the feeder never went high in %d presses", maxPresses)
	}
	return all, nil
}

func day20(path string) error {
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
	nw := new(network)
	if err := nw.load(string(content)); err != nil {
		return err
	}
	var counts [2]int
	for range 1000 {
		if _, err := nw.press(&counts, -1); err != nil {
			return err
		}
	}
	product := counts[0] * counts[1]
	nw = new(network)
	if err := nw.load(string(content)); err != nil {
		return err
	}
	presses, err := nw.firstLow()
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", product, presses); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
