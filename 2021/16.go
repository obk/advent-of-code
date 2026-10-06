// Day 16: Packet Decoder
//
// The input is a hexadecimal transmission holding one BITS packet. A
// packet has a 3-bit version and a 3-bit type. Type 4 is a literal number
// in 5-bit groups; any other type is an operator holding sub-packets,
// given by their total length in bits (15-bit field) or by their count
// (11-bit field).
//
// Part 1 sums the versions of every packet.
// Part 2 evaluates the outer packet: types 0-3 are sum, product, minimum
// and maximum; 5, 6 and 7 compare two values (greater, less, equal) giving
// 1 or 0.
//
// Packets nest, so open operators are kept on a stack. When a packet
// finishes, its value is handed to the operator below it, which may then
// finish too.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 12
	maxDepth      = 64
	maxPackets    = 4 * maxInputBytes
)

// operator is an open operator packet.
type operator struct {
	kind   int
	end    int // bit where its sub-packets end, or -1 if counted
	left   int // sub-packets still to come, if counted
	value  int
	values int
}

// decoder reads bits and keeps the open operators.
type decoder struct {
	bits     []byte
	pos      int
	stack    [maxDepth]operator
	depth    int
	versions int
}

// read takes the next n bits as a number.
func (d *decoder) read(n int) (int, error) {
	if n < 1 || n > 62 || d.pos+n > len(d.bits) {
		return 0, fmt.Errorf("cannot read %d bits at %d of %d", n, d.pos, len(d.bits))
	}
	v := 0
	for range n {
		v, d.pos = v<<1|int(d.bits[d.pos]), d.pos+1
	}
	return v, nil
}

// literal reads a literal's 5-bit groups.
func (d *decoder) literal() (int, error) {
	v := 0
	for group := 0; group < 16; group++ {
		g, err := d.read(5)
		if err != nil {
			return 0, err
		}
		v = v<<4 | g&15
		if v < 0 {
			return 0, errors.New("literal overflowed")
		}
		if g < 16 {
			return v, nil
		}
	}
	return 0, errors.New("literal is longer than 16 groups")
}

// combine folds a finished sub-packet's value into operator o.
func combine(o *operator, v int) error {
	if o == nil || o.kind == 4 || o.kind < 0 || o.kind > 7 {
		return errors.New("not an operator")
	}
	if o.kind == 1 && o.values > 0 && v != 0 && (o.value > math.MaxInt64/v || o.value < 0 || v < 0) {
		return errors.New("product overflows")
	}
	switch {
	case o.values == 0:
		o.value = v
	case o.kind == 0:
		o.value += v
	case o.kind == 1:
		o.value *= v
	case o.kind == 2:
		o.value = min(o.value, v)
	case o.kind == 3:
		o.value = max(o.value, v)
	case o.values > 1:
		return fmt.Errorf("comparison %d has more than two operands", o.kind)
	default:
		r := (o.kind == 5 && o.value > v) || (o.kind == 6 && o.value < v) || (o.kind == 7 && o.value == v)
		o.value = 0
		if r {
			o.value = 1
		}
	}
	o.values++
	o.left--
	return nil
}

// packet reads one packet header and either its literal value (done is
// true) or opens an operator.
func (d *decoder) packet() (value int, done bool, err error) {
	if d.pos+6 > len(d.bits) {
		return 0, false, fmt.Errorf("no room for a packet at bit %d", d.pos)
	}
	version, err := d.read(3)
	if err != nil {
		return 0, false, err
	}
	kind, err := d.read(3)
	if err != nil {
		return 0, false, err
	}
	d.versions += version
	if kind == 4 {
		v, err := d.literal()
		return v, true, err
	}
	if d.depth == maxDepth {
		return 0, false, fmt.Errorf("packets nest deeper than %d", maxDepth)
	}
	o := operator{kind: kind, end: -1}
	lengthType, err := d.read(1)
	if err != nil {
		return 0, false, err
	}
	if lengthType == 0 {
		n, err := d.read(15)
		o.end = d.pos + n
		if err != nil {
			return 0, false, err
		}
	} else if o.left, err = d.read(11); err != nil {
		return 0, false, err
	}
	d.stack[d.depth], d.depth = o, d.depth+1
	return 0, false, nil
}

// run decodes the outer packet and returns its value.
func (d *decoder) run() (int, error) {
	if len(d.bits) == 0 {
		return 0, errors.New("empty transmission")
	}
	for range maxPackets {
		v, done, err := d.packet()
		if err != nil {
			return 0, err
		}
		// Hand finished values down until an operator still needs more.
		for k := 0; k <= maxDepth; k++ {
			if done && d.depth == 0 {
				return v, nil
			}
			if done {
				if err := combine(&d.stack[d.depth-1], v); err != nil {
					return 0, err
				}
			}
			top := &d.stack[d.depth-1]
			finished := (top.end >= 0 && d.pos >= top.end) || (top.end < 0 && top.left == 0)
			if !finished {
				break
			}
			if top.values == 0 || (top.kind >= 5 && top.values != 2) {
				return 0, fmt.Errorf("operator %d finished with %d values", top.kind, top.values)
			}
			v, done, d.depth = top.value, true, d.depth-1
		}
	}
	return 0, fmt.Errorf("more than %d packets", maxPackets)
}

func day16(path string) error {
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
	hex := strings.TrimSpace(string(content))
	if hex == "" {
		return fmt.Errorf("%s: empty transmission", path)
	}
	d := &decoder{bits: make([]byte, 4*len(hex))}
	for i := range len(hex) {
		v, err := strconv.ParseUint(hex[i:i+1], 16, 8)
		if err != nil {
			return fmt.Errorf("%s: invalid hex digit %q", path, hex[i])
		}
		for b := range 4 {
			d.bits[4*i+b] = byte(v >> (3 - b) & 1)
		}
	}
	value, err := d.run()
	if err != nil {
		return err
	}

	if d.versions < 0 {
		return errors.New("version sum overflowed")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", d.versions, value); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
