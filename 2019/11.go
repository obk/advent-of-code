// Day 11: Space Police
//
// The input is an Intcode program driving a hull-painting robot. It reads
// the colour under the robot (0 black, 1 white), outputs a colour to paint
// and a turn (0 left, 1 right), and the robot then steps forward. The robot
// starts facing up on an all-black hull.
//
// Part 1 counts the panels painted at least once.
// Part 2 starts the robot on a white panel, and reads the eight letters it
// paints: 4 panels wide with 1 of spacing, six tall, starting one panel right
// of the robot's start. Each is matched against the known glyphs of the
// Advent of Code font, and one that matches none is an error.
//
// The hull is a fixed grid centred on the start; leaving it is an error.
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
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 12
	maxQueue      = 64
	maxSteps      = 100_000_000
	side          = 201
	maxMoves      = 1_000_000
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

// hull is the painted surface: colour, and whether each panel was painted.
type hull struct {
	white, painted [side][side]bool
}

// paint runs the robot from the centre, starting on a panel of the given
// colour, and returns how many panels it painted.
func (h *hull) paint(m *machine, program []int, startWhite bool) (int, error) {
	if err := m.load(program); err != nil {
		return 0, err
	}
	x, y, dir := side/2, side/2, 0 // dir: 0 up, 1 right, 2 down, 3 left
	h.white[y][x] = startWhite
	for range maxMoves {
		input := 0
		if h.white[y][x] {
			input = 1
		}
		if err := m.push(input); err != nil {
			return 0, err
		}
		var outs [2]int
		for k := range outs {
			status, out, err := m.run()
			if err != nil {
				return 0, err
			}
			if status == halted {
				return h.count(), nil
			}
			if status != gaveOutput || out < 0 || out > 1 {
				return 0, fmt.Errorf("robot gave status %d, output %d", status, out)
			}
			outs[k] = out
		}
		h.white[y][x], h.painted[y][x] = outs[0] == 1, true
		dir = (dir + 3 + 2*outs[1]) % 4 // left is -1, right +1
		x, y = x+[4]int{0, 1, 0, -1}[dir], y+[4]int{-1, 0, 1, 0}[dir]
		if x < 0 || y < 0 || x >= side || y >= side {
			return 0, errors.New("the robot left the hull")
		}
	}
	return 0, fmt.Errorf("the robot moved more than %d times", maxMoves)
}

// count returns how many panels have been painted.
func (h *hull) count() int {
	n := 0
	for y := range side {
		for x := range side {
			if h.painted[y][x] {
				n++
			}
		}
	}
	return n
}

// read returns the letters painted on the hull, into out.
func (h *hull) read(out []byte) ([]byte, error) {
	if cap(out) < 8 {
		return nil, errors.New("no room for eight letters")
	}
	// The font: each letter followed by its six rows of four pixels.
	font := [17]string{
		"A.##.#..##..######..##..#", "B###.#..####.#..##..####.", "C.##.#..##...#...#..#.##.",
		"E#####...###.#...#...####", "F#####...###.#...#...#...", "G.##.#..##...#.###..#.###",
		"H#..##..######..##..##..#", "I.###..#...#...#...#..###", "J..##...#...#...##..#.##.",
		"K#..##.#.##..#.#.#.#.#..#", "L#...#...#...#...#...####", "O.##.#..##..##..##..#.##.",
		"P###.#..##..####.#...#...", "R###.#..##..####.#.#.#..#", "S.####...#....##....####.",
		"U#..##..##..##..##..#.##.", "Z####...#..#..#..#...####",
	}
	out = out[:0]
	var cell [24]byte
	for l := range 8 {
		for r := range 6 {
			for c := range 4 {
				cell[r*4+c] = '.'
				if h.white[side/2+r][side/2+1+l*5+c] {
					cell[r*4+c] = '#'
				}
			}
		}
		letter := byte(0)
		for _, g := range font {
			if g[1:] == string(cell[:]) {
				letter = g[0]
			}
		}
		if letter == 0 {
			return nil, fmt.Errorf("letter %d does not match the font: %s", l, cell[:])
		}
		out = append(out, letter)
	}
	return out, nil
}

func day11(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	m := &machine{mem: make([]int, memory)}
	h := new(hull)
	painted, err := h.paint(m, program, false)
	if err != nil {
		return err
	}

	if painted < 1 {
		return errors.New("the robot painted nothing")
	}
	h = new(hull)
	if _, err := h.paint(m, program, true); err != nil {
		return err
	}
	id, err := h.read(make([]byte, 0, 8))
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", painted, id); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day11(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
