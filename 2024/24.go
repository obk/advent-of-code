// Day 24: Crossed Wires
//
// The input is initial wire values "x00: 1", then gates "a AND b -> c"
// (AND, OR, XOR). The z wires, read as a binary number (z00 lowest), are
// the output.
//
// Part 1 is the number the z wires output.
// Part 2: the gates are meant to add the x and y numbers, but the outputs
// of four pairs of gates were swapped. The answer lists the eight swapped
// wires, sorted and joined with commas.
//
// Part 1 evaluates the gates in passes until every one has fired. For
// part 2, a ripple-carry adder has a fixed shape, and a gate that breaks
// it has a swapped output: a z wire must come from a XOR (the top one from
// an OR); a XOR not fed by x and y must output a z; an x/y XOR must feed a
// XOR; an AND must feed an OR. (The lowest bit's half adder is exempt.)
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 14
	maxWires      = 512
	swapped       = 8
)

// gate is a, op, b -> out (wire numbers).
type gate struct {
	a, b, out int
	op        string
}

// circuit is the wires and gates.
type circuit struct {
	names [maxWires]string
	ids   map[string]int
	n     int
	value [maxWires]int // -1 unknown
	gates []gate
}

// id returns a wire's number, numbering it if new.
func (c *circuit) id(name string) (int, error) {
	if i, ok := c.ids[name]; ok {
		return i, nil
	}
	if c.n == maxWires || len(name) != 3 {
		return 0, fmt.Errorf("too many wires, or a bad name %q", name)
	}
	c.ids[name], c.names[c.n], c.value[c.n] = c.n, name, -1
	c.n++
	return c.n - 1, nil
}

// load parses the wires and gates.
func (c *circuit) load(text string) error {
	values, gates, found := strings.Cut(strings.TrimSpace(text), "\n\n")
	if !found {
		return errors.New("no blank line between wires and gates")
	}
	c.ids = make(map[string]int, maxWires)
	for _, l := range strings.Split(values, "\n") {
		name, v, ok := strings.Cut(l, ": ")
		w, err := c.id(name)
		if !ok || err != nil || (v != "0" && v != "1") {
			return fmt.Errorf("malformed wire %q", l)
		}
		c.value[w] = int(v[0] - '0')
	}
	lines := strings.Split(gates, "\n")
	if len(lines) == 0 || len(lines) > maxWires {
		return fmt.Errorf("%d gates is out of range", len(lines))
	}
	c.gates = make([]gate, len(lines))
	var driven [maxWires]bool
	for i, l := range lines {
		f := strings.Fields(l)
		if len(f) != 5 || f[3] != "->" || (f[1] != "AND" && f[1] != "OR" && f[1] != "XOR") {
			return fmt.Errorf("malformed gate %q", l)
		}
		var ids [3]int
		for k, name := range [3]string{f[0], f[2], f[4]} {
			var err error
			if ids[k], err = c.id(name); err != nil {
				return err
			}
		}
		if driven[ids[2]] || c.value[ids[2]] >= 0 {
			return fmt.Errorf("wire %s is driven twice", f[4])
		}
		driven[ids[2]] = true
		c.gates[i] = gate{ids[0], ids[1], ids[2], f[1]}
	}
	return nil
}

// output fires every gate and reads the z wires.
func (c *circuit) output() (int, error) {
	if len(c.gates) == 0 {
		return 0, errors.New("no gates")
	}
	for pass := 0; pass <= len(c.gates); pass++ {
		changed := false
		for _, g := range c.gates {
			if c.value[g.out] >= 0 || c.value[g.a] < 0 || c.value[g.b] < 0 {
				continue
			}
			switch g.op {
			case "AND":
				c.value[g.out] = c.value[g.a] & c.value[g.b]
			case "OR":
				c.value[g.out] = c.value[g.a] | c.value[g.b]
			default:
				c.value[g.out] = c.value[g.a] ^ c.value[g.b]
			}
			changed = true
		}
		if !changed {
			break
		}
	}
	z := 0
	for w := c.n - 1; w >= 0; w-- {
		if c.names[w][0] != 'z' {
			continue
		}
		if c.value[w] < 0 {
			return 0, fmt.Errorf("wire %s never got a value", c.names[w])
		}
		bit := int(c.names[w][1]-'0')*10 + int(c.names[w][2]-'0')
		if bit < 0 || bit > 62 {
			return 0, fmt.Errorf("wire %s is past bit 62", c.names[w])
		}
		z |= c.value[w] << bit
	}
	return z, nil
}

// feeds reports whether wire w is an input of some gate with operation op.
func (c *circuit) feeds(w int, op string) bool {
	for _, g := range c.gates {
		if (g.a == w || g.b == w) && g.op == op {
			return true
		}
	}
	return false
}

// suspects lists the wires whose gates break the adder's shape, sorted,
// into out.
func (c *circuit) suspects(out []string) ([]string, error) {
	top := ""
	for _, name := range c.names[:c.n] {
		if name[0] == 'z' && name > top {
			top = name
		}
	}
	if top == "" {
		return nil, errors.New("no z wires")
	}
	out = out[:0]
	for _, g := range c.gates {
		name := c.names[g.out]
		xy := strings.IndexByte("xy", c.names[g.a][0]) >= 0
		lowest := c.names[g.a][1:] == "00"
		bad := (name[0] == 'z' && name != top && g.op != "XOR") ||
			(name == top && g.op != "OR") ||
			(g.op == "XOR" && !xy && name[0] != 'z') ||
			(g.op == "XOR" && xy && !lowest && !c.feeds(g.out, "XOR")) ||
			(g.op == "AND" && !lowest && !c.feeds(g.out, "OR"))
		if !bad {
			continue
		}
		if len(out) == cap(out) {
			return nil, fmt.Errorf("more than %d suspect wires", cap(out))
		}
		out = append(out, name)
		for k := len(out) - 1; k > 0 && out[k-1] > out[k]; k-- {
			out[k-1], out[k] = out[k], out[k-1]
		}
	}
	if len(out) != swapped {
		return nil, fmt.Errorf("found %d suspect wires, expected %d", len(out), swapped)
	}
	return out, nil
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
	c := new(circuit)
	if err := c.load(string(content)); err != nil {
		return err
	}
	z, err := c.output()
	if err != nil {
		return err
	}
	wires, err := c.suspects(make([]string, 0, 2*swapped))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", z, strings.Join(wires, ",")); err != nil {
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
