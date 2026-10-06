// Day 13: Care Package
//
// The input is an Intcode arcade game. It draws tiles with triples of
// outputs (x, y, tile): 0 empty, 1 wall, 2 block, 3 paddle, 4 ball.
//
// Part 1 counts the block tiles on screen when the game exits.
// Part 2 plays the game (address 0 set to 2 for free play) to the end:
// when asked for input, the joystick moves the paddle towards the ball.
// Outputs at x=-1, y=0 are the score; the answer is the final score.
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
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 12
	maxQueue      = 64
	maxSteps      = 1_000_000_000
	screenSide    = 64
	maxOutputs    = 10_000_000
	block         = 2
	paddle        = 3
	ball          = 4
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

// screen is the arcade display.
type screen struct {
	tiles [screenSide][screenSide]int
}

// draw runs the game until it halts, drawing every tile it outputs.
func (s *screen) draw(m *machine) error {
	if m == nil || len(m.mem) == 0 {
		return errors.New("no machine to run")
	}
	var triple [3]int
	for n := 0; n < maxOutputs; n++ {
		status, out, err := m.run()
		if err != nil {
			return err
		}
		switch status {
		case halted:
			if n%3 != 0 {
				return fmt.Errorf("the game halted after %d outputs, not whole tiles", n)
			}
			return nil
		case needInput:
			return errors.New("the game asked for input")
		}
		triple[n%3] = out
		if n%3 != 2 {
			continue
		}
		x, y, tile := triple[0], triple[1], triple[2]
		if x < 0 || y < 0 || x >= screenSide || y >= screenSide || tile < 0 || tile > 4 {
			return fmt.Errorf("tile %d at %d,%d is off the screen", tile, x, y)
		}
		s.tiles[y][x] = tile
	}
	return fmt.Errorf("the game gave more than %d outputs", maxOutputs)
}

// count returns how many tiles of a kind are on screen.
func (s *screen) count(kind int) (int, error) {
	if kind < 0 || kind > 4 {
		return 0, fmt.Errorf("tile kind %d does not exist", kind)
	}
	n := 0
	for y := range screenSide {
		for x := range screenSide {
			if s.tiles[y][x] == kind {
				n++
			}
		}
	}
	return n, nil
}

// play runs the game with the joystick following the ball, and returns the
// final score.
func (s *screen) play(m *machine) (int, error) {
	if m == nil || len(m.mem) == 0 {
		return 0, errors.New("no machine to run")
	}
	var triple [3]int
	score, ballX, paddleX, n := 0, 0, 0, 0
	for range maxOutputs {
		status, out, err := m.run()
		if err != nil {
			return 0, err
		}
		switch status {
		case halted:
			return score, nil
		case needInput:
			if err := m.push(max(-1, min(1, ballX-paddleX))); err != nil {
				return 0, err
			}
			continue
		}
		triple[n%3], n = out, n+1
		if n%3 != 0 {
			continue
		}
		x, y, tile := triple[0], triple[1], triple[2]
		if x == -1 && y == 0 {
			score = tile
			continue
		}
		if x < 0 || y < 0 || x >= screenSide || y >= screenSide || tile < 0 || tile > 4 {
			return 0, fmt.Errorf("tile %d at %d,%d is off the screen", tile, x, y)
		}
		s.tiles[y][x] = tile
		switch tile {
		case ball:
			ballX = x
		case paddle:
			paddleX = x
		}
	}
	return 0, fmt.Errorf("the game gave more than %d outputs", maxOutputs)
}

func day13(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	m := &machine{mem: make([]int, memory)}
	if err := m.load(program); err != nil {
		return err
	}
	s := new(screen)
	if err := s.draw(m); err != nil {
		return err
	}
	blocks, err := s.count(block)
	if err != nil {
		return err
	}

	if err := m.load(program); err != nil {
		return err
	}
	m.mem[0] = 2
	s = new(screen)
	score, err := s.play(m)
	if err != nil {
		return err
	}
	left, err := s.count(block)
	if err != nil {
		return err
	}
	if left != 0 {
		return fmt.Errorf("the game ended with %d blocks left", left)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", blocks, score); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
