// Day 19: Tractor Beam
//
// The input is an Intcode drone program: given X and Y (both >= 0), it
// outputs 1 if the tractor beam pulls at that point and 0 if not.
//
// Part 1 counts the points pulled in the 50x50 area nearest the emitter.
// Part 2 finds the 100x100 square nearest the emitter that fits wholly in
// the beam, and gives 10000*x + y of its top-left corner.
//
// The beam widens away from the emitter, and its left edge only moves
// right row by row. So the left edge is followed down, row by row; a square
// with its bottom-left corner there fits if its top-right corner is pulled.
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
	maxInputBytes = 1 << 20
	memory        = 1 << 10
	maxQueue      = 8
	maxSteps      = 10_000_000
	scan          = 50
	square        = 100
	maxRow        = 100_000
)

// machine is an Intcode computer: memory, instruction pointer, relative
// base, a queue of input values, and a count of the steps it has run.
type machine struct {
	mem      []int
	pc, base int
	in       [maxQueue]int
	inHead   int
	inLen    int
	steps    int
}

// What a run stopped for.
const (
	halted = iota
	needInput
	gaveOutput
)

// load resets the machine and copies a program into its memory, which must
// already be allocated.
func (m *machine) load(program []int) error {
	if len(program) == 0 || len(program) > len(m.mem) {
		return fmt.Errorf("program of %d values does not fit memory of %d", len(program), len(m.mem))
	}
	if len(m.mem) != memory {
		return fmt.Errorf("machine memory is %d values, not %d", len(m.mem), memory)
	}
	clear(m.mem)
	copy(m.mem, program)
	m.pc, m.base, m.inHead, m.inLen, m.steps = 0, 0, 0, 0, 0
	return nil
}

// push queues an input value.
func (m *machine) push(v int) error {
	if len(m.mem) == 0 {
		return errors.New("machine has no memory")
	}
	if m.inLen == maxQueue {
		return fmt.Errorf("input queue of %d values is full", maxQueue)
	}
	m.in[(m.inHead+m.inLen)%maxQueue] = v
	m.inLen++
	return nil
}

// addr returns the memory address of parameter n (1..3) of the current
// instruction, by its mode: 0 position, 1 immediate, 2 relative.
func (m *machine) addr(n int) (int, error) {
	if n < 1 || n > 3 {
		return 0, fmt.Errorf("parameter %d does not exist", n)
	}
	if m.pc+n >= len(m.mem) {
		return 0, fmt.Errorf("parameter %d at %d is past memory", n, m.pc)
	}
	mode := m.mem[m.pc] / [4]int{0, 100, 1000, 10000}[n] % 10
	var a int
	switch mode {
	case 0:
		a = m.mem[m.pc+n]
	case 1:
		a = m.pc + n
	case 2:
		a = m.base + m.mem[m.pc+n]
	default:
		return 0, fmt.Errorf("unknown parameter mode %d at %d", mode, m.pc)
	}
	if a < 0 || a >= len(m.mem) {
		return 0, fmt.Errorf("address %d is outside memory of %d", a, len(m.mem))
	}
	return a, nil
}

// decode reads the current instruction: its opcode, width, and the
// addresses of its parameters.
func (m *machine) decode() (op, width int, p [4]int, err error) {
	if m.pc < 0 || m.pc >= len(m.mem) {
		return 0, 0, p, fmt.Errorf("instruction pointer %d is outside memory", m.pc)
	}
	op = m.mem[m.pc] % 100
	if op == 99 {
		return op, 1, p, nil
	}
	if op < 1 || op > 9 {
		return 0, 0, p, fmt.Errorf("unknown opcode %d at %d", m.mem[m.pc], m.pc)
	}
	width = [10]int{0, 4, 4, 2, 2, 3, 3, 4, 4, 2}[op]
	// Parameters an instruction writes to are never in immediate mode.
	writes := [10]int{0, 3, 3, 1, 0, 0, 0, 3, 3, 0}[op]
	if writes > 0 && m.mem[m.pc]/[4]int{0, 100, 1000, 10000}[writes]%10 == 1 {
		return 0, 0, p, fmt.Errorf("instruction %d at %d writes to an immediate parameter", m.mem[m.pc], m.pc)
	}
	for k := 1; k < width; k++ {
		if p[k], err = m.addr(k); err != nil {
			return 0, 0, p, err
		}
	}
	return op, width, p, nil
}

// run executes until the program halts, needs input it does not have, or
// outputs a value (which it returns).
func (m *machine) run() (status, out int, err error) {
	for n := 0; n < maxSteps && m.steps < maxSteps; n++ {
		m.steps++
		op, width, p, err := m.decode()
		if err != nil {
			return 0, 0, err
		}
		if op == 99 {
			return halted, 0, nil
		}
		next := m.pc + width
		switch op {
		case 1:
			m.mem[p[3]] = m.mem[p[1]] + m.mem[p[2]]
		case 2:
			m.mem[p[3]] = m.mem[p[1]] * m.mem[p[2]]
		case 3:
			if m.inLen == 0 {
				return needInput, 0, nil
			}
			m.mem[p[1]] = m.in[m.inHead]
			m.inHead, m.inLen = (m.inHead+1)%maxQueue, m.inLen-1
		case 4:
			m.pc = next
			return gaveOutput, m.mem[p[1]], nil
		case 5, 6:
			if (m.mem[p[1]] != 0) == (op == 5) {
				next = m.mem[p[2]]
			}
			if next < 0 || next >= len(m.mem) {
				return 0, 0, fmt.Errorf("jump at %d to %d leaves memory", m.pc, next)
			}
		case 7, 8:
			// Compare first: the output cell may also be an input.
			result := 0
			if (op == 7 && m.mem[p[1]] < m.mem[p[2]]) || (op == 8 && m.mem[p[1]] == m.mem[p[2]]) {
				result = 1
			}
			m.mem[p[3]] = result
		case 9:
			m.base += m.mem[p[1]]
			if m.base < 0 || m.base >= len(m.mem) {
				return 0, 0, fmt.Errorf("relative base %d is outside memory", m.base)
			}
		}
		m.pc = next
	}
	return 0, 0, fmt.Errorf("the program ran more than %d steps", maxSteps)
}

// parseProgram reads the comma-separated program.
func parseProgram(path string) ([]int, error) {
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
	fields := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(fields) < 1 || fields[0] == "" {
		return nil, fmt.Errorf("%s: empty program", path)
	}
	if len(fields) > memory {
		return nil, fmt.Errorf("%s: %d values do not fit memory of %d", path, len(fields), memory)
	}
	program := make([]int, len(fields))
	for i, f := range fields {
		if program[i], err = strconv.Atoi(strings.TrimSpace(f)); err != nil {
			return nil, fmt.Errorf("%s: invalid value %q: %w", path, f, err)
		}
	}
	return program, nil
}

// pulled runs the drone program once for one point.
func pulled(m *machine, program []int, x, y int) (bool, error) {
	if x < 0 || y < 0 {
		return false, fmt.Errorf("point %d,%d is negative", x, y)
	}
	if err := m.load(program); err != nil {
		return false, err
	}
	if err := m.push(x); err != nil {
		return false, err
	}
	if err := m.push(y); err != nil {
		return false, err
	}
	status, out, err := m.run()
	if err != nil {
		return false, err
	}
	if status != gaveOutput || out < 0 || out > 1 {
		return false, fmt.Errorf("drone gave status %d, output %d", status, out)
	}
	return out == 1, nil
}

// fit returns the top-left corner of the nearest square that fits.
func fit(m *machine, program []int, size int) (x, y int, err error) {
	if size < 1 || size > maxRow/10 {
		return 0, 0, fmt.Errorf("square size %d is out of range", size)
	}
	left := 0
	for row := size - 1; row < maxRow; row++ {
		found := false
		// Near the emitter some rows miss the beam; the edge stays put then.
		// The first row searched may find the edge far out if the beam is
		// steep, so the search reaches well past the row's own number.
		for col := left; col < left+4*row+size && !found; col++ {
			if found, err = pulled(m, program, col, row); err != nil {
				return 0, 0, err
			}
			if found {
				left = col
			}
		}
		if !found {
			continue
		}
		in, err := pulled(m, program, left+size-1, row-size+1)
		if err != nil {
			return 0, 0, err
		}
		if in {
			return left, row - size + 1, nil
		}
	}
	return 0, 0, fmt.Errorf("no square fits within %d rows", maxRow)
}

func day19(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	m := &machine{mem: make([]int, memory)}
	count := 0
	for y := range scan {
		for x := range scan {
			in, err := pulled(m, program, x, y)
			if err != nil {
				return err
			}
			if in {
				count++
			}
		}
	}
	if count == 0 {
		return errors.New("the beam pulls nowhere")
	}

	x, y, err := fit(m, program, square)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, 10000*x+y); err != nil {
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
