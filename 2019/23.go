// Day 23: Category Six
//
// The input is an Intcode network interface program. Fifty computers run
// it, each first given its address 0..49. A computer sends a packet with
// three outputs (address, X, Y); it reads packets by taking X then Y as
// input, and gets -1 when nothing is waiting.
//
// Part 1 is the Y value of the first packet sent to address 255.
// Part 2 makes address 255 a NAT: it keeps the last packet sent to it, and
// when the network is idle (no computer sent anything for two whole
// rounds, so every queue has drained) it sends that packet to computer 0. The
// answer is the first Y it sends to computer 0 twice in a row.
//
// The computers take turns: each runs until it wants input it does not
// have (it then gets -1) or has sent a packet.
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
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 12
	maxQueue      = 1 << 10
	maxSteps      = 100_000_000
	computers     = 50
	nat           = 255
	maxRounds     = 1_000_000
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

// network is the fifty computers and each one's partial outgoing packet.
type network struct {
	nodes [computers]machine
	out   [computers][3]int
	nout  [computers]int
}

// boot loads every computer and gives it its address.
func (n *network) boot(program []int) error {
	if len(program) == 0 {
		return errors.New("no program")
	}
	for i := range computers {
		if len(n.nodes[i].mem) == 0 {
			n.nodes[i].mem = make([]int, memory)
		}
		n.nout[i] = 0
		if err := n.nodes[i].load(program); err != nil {
			return err
		}
		if err := n.nodes[i].push(i); err != nil {
			return err
		}
	}
	return nil
}

// turn runs computer i until it wants input it lacks or finishes a packet.
// It returns the finished packet's address (-1 if none) and values.
func (n *network) turn(i int) (to, x, y int, err error) {
	if i < 0 || i >= computers {
		return 0, 0, 0, fmt.Errorf("computer %d does not exist", i)
	}
	m := &n.nodes[i]
	if m.inLen == 0 { // nothing waiting: its next read gets -1
		if err := m.push(-1); err != nil {
			return 0, 0, 0, err
		}
	}
	for range 8 {
		status, out, err := m.run()
		if err != nil {
			return 0, 0, 0, err
		}
		switch status {
		case halted:
			return 0, 0, 0, fmt.Errorf("computer %d halted", i)
		case needInput:
			if m.inLen > 0 {
				return 0, 0, 0, errors.New("asked for input it already has")
			}
			return -1, 0, 0, nil
		}
		if n.nout[i] < 0 || n.nout[i] > 2 {
			return 0, 0, 0, fmt.Errorf("computer %d has %d packet values", i, n.nout[i])
		}
		n.out[i][n.nout[i]] = out
		n.nout[i]++
		if n.nout[i] == 3 {
			n.nout[i] = 0
			return n.out[i][0], n.out[i][1], n.out[i][2], nil
		}
	}
	return -1, 0, 0, nil
}

// firstToNAT runs the network until a packet is sent to address 255 and
// returns its Y.
func (n *network) firstToNAT() (int, error) {
	for range maxRounds {
		for i := range computers {
			to, x, y, err := n.turn(i)
			if err != nil {
				return 0, err
			}
			if to == nat {
				return y, nil
			}
			if to < -1 || to >= computers {
				return 0, fmt.Errorf("packet to unknown address %d", to)
			}
			if to < 0 {
				continue
			}
			if err := n.nodes[to].push(x); err != nil {
				return 0, err
			}
			if err := n.nodes[to].push(y); err != nil {
				return 0, err
			}
		}
	}
	return 0, fmt.Errorf("no packet reached %d in %d rounds", nat, maxRounds)
}

// natLoop runs the network with the NAT until it repeats a Y.
func (n *network) natLoop() (int, error) {
	natX, natY, lastY, have, sentY := 0, 0, 0, false, false
	quiet := 0
	for range maxRounds {
		busy := false
		for i := range computers {
			to, x, y, err := n.turn(i)
			if err != nil {
				return 0, err
			}
			switch {
			case to == nat:
				natX, natY, have, busy = x, y, true, true
			case to >= 0 && to < computers:
				busy = true
				if err := n.nodes[to].push(x); err != nil {
					return 0, err
				}
				if err := n.nodes[to].push(y); err != nil {
					return 0, err
				}
			case to != -1:
				return 0, fmt.Errorf("packet to unknown address %d", to)
			}
		}
		if quiet = quiet + 1; busy {
			quiet = 0
		}
		if quiet < 2 || !have {
			continue
		}
		if sentY && natY == lastY {
			return natY, nil
		}
		if err := n.nodes[0].push(natX); err != nil {
			return 0, err
		}
		if err := n.nodes[0].push(natY); err != nil {
			return 0, err
		}
		lastY, sentY, quiet = natY, true, 0
	}
	return 0, fmt.Errorf("the NAT did not repeat in %d rounds", maxRounds)
}

func day23(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	n := new(network)
	if err := n.boot(program); err != nil {
		return err
	}
	y, err := n.firstToNAT()
	if err != nil {
		return err
	}

	if err := n.boot(program); err != nil {
		return err
	}
	repeat, err := n.natLoop()
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", y, repeat); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day23(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
