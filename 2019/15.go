// Day 15: Oxygen System
//
// The input is an Intcode program controlling a repair droid. Given a move
// (1 north, 2 south, 3 west, 4 east) it replies 0 (hit a wall), 1 (moved)
// or 2 (moved, and found the oxygen system).
//
// Part 1 is the fewest moves from the start to the oxygen system.
// Part 2 is the minutes for oxygen to fill the area, spreading one cell a
// minute from the oxygen system: the farthest any open cell is from it.
//
// The droid maps the whole area first by depth-first search, keeping the
// moves that led to its position on a stack so it can back out of dead
// ends. A breadth-first search over the map then gives the distance.
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
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 12
	maxQueue      = 64
	maxSteps      = 100_000_000
	side          = 101
	maxMoves      = 4 * 4 * side * side
)

// Map cells.
const (
	unknown = iota
	open
	wall
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

// area is the map the droid builds, with the oxygen system's position.
type area struct {
	cell       [side][side]int
	dist       [side][side]int
	stack      [side * side]int
	queue      [side * side][2]int
	oxyX, oxyY int
	found      bool
}

// step offsets for moves 1..4; reverse[d] undoes move d.
func step(d int) (dx, dy, back int) {
	return [5]int{0, 0, 0, -1, 1}[d], [5]int{0, -1, 1, 0, 0}[d], [5]int{0, 2, 1, 4, 3}[d]
}

// move sends one move to the droid and returns its reply.
func move(m *machine, d int) (int, error) {
	if d < 1 || d > 4 {
		return 0, fmt.Errorf("move %d does not exist", d)
	}
	if err := m.push(d); err != nil {
		return 0, err
	}
	status, out, err := m.run()
	if err != nil {
		return 0, err
	}
	if status != gaveOutput || out < 0 || out > 2 {
		return 0, fmt.Errorf("droid gave status %d, reply %d", status, out)
	}
	return out, nil
}

// explore maps every reachable cell, starting from the centre.
func (a *area) explore(m *machine) error {
	x, y, depth := side/2, side/2, 0
	a.cell[y][x] = open
	for range maxMoves {
		d := 5
		for k := 4; k >= 1; k-- {
			dx, dy, _ := step(k)
			if a.cell[y+dy][x+dx] == unknown {
				d = k
			}
		}
		if d > 4 { // nothing new here: back out, or finish at the start
			if depth == 0 {
				return nil
			}
			depth--
			dx, dy, back := step(a.stack[depth])
			if reply, err := move(m, back); err != nil || reply == 0 {
				return fmt.Errorf("could not back out (%v)", err)
			}
			x, y = x-dx, y-dy
			continue
		}
		dx, dy, _ := step(d)
		reply, err := move(m, d)
		if err != nil {
			return err
		}
		if reply == 0 {
			a.cell[y+dy][x+dx] = wall
			continue
		}
		x, y = x+dx, y+dy
		if x < 1 || y < 1 || x >= side-1 || y >= side-1 || depth == len(a.stack) {
			return errors.New("the droid left the map")
		}
		a.cell[y][x], a.stack[depth], depth = open, d, depth+1
		if reply == 2 {
			a.oxyX, a.oxyY, a.found = x, y, true
		}
	}
	return fmt.Errorf("exploring took more than %d moves", maxMoves)
}

// distances fills dist with the moves from (x, y) to every open cell
// (-1 where unreachable) and returns the largest.
func (a *area) distances(x, y int) (int, error) {
	if x < 0 || y < 0 || x >= side || y >= side || a.cell[y][x] != open {
		return 0, fmt.Errorf("%d,%d is not an open cell", x, y)
	}
	for r := range side {
		for c := range side {
			a.dist[r][c] = -1
		}
	}
	a.dist[y][x] = 0
	a.queue[0] = [2]int{x, y}
	farthest := 0
	for head, tail := 0, 1; head < tail; head++ {
		p := a.queue[head]
		for d := 1; d <= 4; d++ {
			dx, dy, _ := step(d)
			nx, ny := p[0]+dx, p[1]+dy
			if a.cell[ny][nx] != open || a.dist[ny][nx] >= 0 {
				continue
			}
			if tail == len(a.queue) {
				return 0, errors.New("search queue is full")
			}
			a.dist[ny][nx] = a.dist[p[1]][p[0]] + 1
			farthest = max(farthest, a.dist[ny][nx])
			a.queue[tail], tail = [2]int{nx, ny}, tail+1
		}
	}
	return farthest, nil
}

func day15(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	m := &machine{mem: make([]int, memory)}
	if err := m.load(program); err != nil {
		return err
	}
	a := new(area)
	if err := a.explore(m); err != nil {
		return err
	}
	if !a.found {
		return errors.New("the droid never found the oxygen system")
	}
	if _, err := a.distances(side/2, side/2); err != nil {
		return err
	}

	moves := a.dist[a.oxyY][a.oxyX]
	if moves < 1 {
		return errors.New("the oxygen system is not reachable from the start")
	}
	minutes, err := a.distances(a.oxyX, a.oxyY)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", moves, minutes); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
