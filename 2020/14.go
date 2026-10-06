// Day 14: Docking Data
//
// Each line of the input sets a 36-bit mask ("mask = 1X0...") or writes
// memory ("mem[8] = 11").
//
// Part 1 applies the mask to values: 0 and 1 force bits, X leaves them.
// Part 2 applies it to addresses: 1 forces bits, 0 leaves them, and each X
// takes both values, so one write goes to many addresses.
// Each answer is the sum of all values left in memory.
//
// Memory is a map, sized before running from the number of addresses each
// write can touch.
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
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	maskBits      = 36
	maxFloating   = 10
)

// line is a mask (ones, zeros and floating bits) or a write.
type line struct {
	isMask         bool
	ones, floating uint64
	address, value uint64
}

// readProgram parses the lines and returns how many addresses part 2 can
// write at most.
func readProgram(path string) ([]line, int, error) {
	if path == "" {
		return nil, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	texts := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(texts) > maxLines {
		return nil, 0, fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	lines := make([]line, len(texts))
	writes, spread := 0, 1
	for i, t := range texts {
		left, right, found := strings.Cut(t, " = ")
		if !found {
			return nil, 0, fmt.Errorf("%s: malformed line %q", path, t)
		}
		if left == "mask" {
			if len(right) != maskBits || strings.Trim(right, "01X") != "" {
				return nil, 0, fmt.Errorf("%s: malformed mask %q", path, right)
			}
			l := line{isMask: true}
			for k := range maskBits {
				l.ones, l.floating = l.ones<<1, l.floating<<1
				switch right[k] {
				case '1':
					l.ones |= 1
				case 'X':
					l.floating |= 1
				}
			}
			if bits.OnesCount64(l.floating) > maxFloating {
				return nil, 0, fmt.Errorf("%s: mask %q floats too many bits", path, right)
			}
			lines[i], spread = l, 1<<bits.OnesCount64(l.floating)
			continue
		}
		addr, errA := strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(left, "mem["), "]"), 10, 36)
		value, errV := strconv.ParseUint(right, 10, 36)
		if errA != nil || errV != nil || !strings.HasPrefix(left, "mem[") {
			return nil, 0, fmt.Errorf("%s: malformed write %q", path, t)
		}
		lines[i] = line{address: addr, value: value}
		writes += spread
	}
	return lines, writes, nil
}

// run executes the program, masking values (version 1) or addresses
// (version 2), and sums memory.
func run(lines []line, mem map[uint64]uint64, v2 bool) (uint64, error) {
	if mem == nil {
		return 0, errors.New("no memory")
	}
	clear(mem)
	var mask line
	for _, l := range lines {
		if l.isMask {
			mask = l
			continue
		}
		if !v2 {
			fixed := ^mask.floating & (1<<maskBits - 1)
			mem[l.address] = l.value&mask.floating | mask.ones&fixed
			continue
		}
		base := (l.address | mask.ones) &^ mask.floating
		// Walk every subset of the floating bits.
		sub := uint64(0)
		for k := 0; k <= 1<<maxFloating; k++ {
			mem[base|sub] = l.value
			sub = (sub - mask.floating) & mask.floating
			if sub == 0 {
				break
			}
		}
	}
	var sum uint64
	for _, v := range mem {
		sum += v
	}
	return sum, nil
}

func day14(path string) error {
	lines, writes, err := readProgram(path)
	if err != nil {
		return err
	}
	mem := make(map[uint64]uint64, writes)
	v1, err := run(lines, mem, false)
	if err != nil {
		return err
	}
	v2, err := run(lines, mem, true)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", v1, v2); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
