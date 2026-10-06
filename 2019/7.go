// Day 7: Amplification Circuit
//
// The input is an Intcode amplifier program. Five amplifiers run copies of
// it in a chain: each first reads its phase setting, then a signal, and
// outputs a signal for the next; the first gets 0.
//
// Part 1 is the highest signal from the last amplifier over every ordering
// of the phase settings 0..4.
// Part 2 wires the last amplifier back into the first and uses phases 5..9:
// the signal goes round until the amplifiers halt, and the last signal out
// of the last amplifier counts.
//
// The orderings are generated with Heap's algorithm in its iterative form;
// the five machines' memories are allocated once and reloaded for each try.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 12
	maxQueue      = 64
	maxSteps      = 10_000_000
	amps          = 5
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

// chain runs the amplifiers once, in order, with the given phases.
func chain(machines *[amps]machine, program []int, phases [amps]int) (int, error) {
	signal := 0
	for i := range amps {
		if phases[i] < 0 || phases[i] > 9 {
			return 0, fmt.Errorf("phase %d is outside 0..9", phases[i])
		}
		m := &machines[i]
		if err := m.load(program); err != nil {
			return 0, err
		}
		if err := m.push(phases[i]); err != nil {
			return 0, err
		}
		if err := m.push(signal); err != nil {
			return 0, err
		}
		status, out, err := m.run()
		if err != nil {
			return 0, err
		}
		if status != gaveOutput {
			return 0, fmt.Errorf("amplifier %d gave no output", i)
		}
		signal = out
	}
	return signal, nil
}

// loop runs the amplifiers as a feedback loop with the given phases.
func loop(machines *[amps]machine, program []int, phases [amps]int) (int, error) {
	for i := range amps {
		if err := machines[i].load(program); err != nil {
			return 0, err
		}
		if err := machines[i].push(phases[i]); err != nil {
			return 0, err
		}
	}
	signal := 0
	for round := 0; round < maxSteps; round++ {
		for i := range amps {
			m := &machines[i]
			if err := m.push(signal); err != nil {
				return 0, err
			}
			status, out, err := m.run()
			if err != nil {
				return 0, err
			}
			// The amplifiers halt in turn once the last round is over, so the
			// first to halt is A, and the signal it was offered is E's last.
			if status == halted && i == 0 && round > 0 {
				return signal, nil
			}
			if status != gaveOutput {
				return 0, fmt.Errorf("amplifier %d stopped with status %d", i, status)
			}
			signal = out
		}
	}
	return 0, errors.New("the feedback loop never ended")
}

// bestSignal tries every ordering of the phases starting from base (0 or 5)
// and returns the highest signal; base 5 runs the feedback loop.
func bestSignal(machines *[amps]machine, program []int, base int) (int, error) {
	if machines == nil || len(program) == 0 {
		return 0, errors.New("no machines or no program")
	}
	if base != 0 && base != 5 {
		return 0, fmt.Errorf("phase base %d is neither 0 nor 5", base)
	}
	var phases [amps]int
	var counters [amps]int
	for i := range phases {
		phases[i] = base + i
	}
	best, i := 0, 0
	for step := 0; step < 1_000; step++ { // 5! = 120 orderings
		signal, err := chain(machines, program, phases)
		if base == 5 {
			signal, err = loop(machines, program, phases)
		}
		if err != nil {
			return 0, err
		}
		best = max(best, signal)
		for k := 0; k < amps && i < amps && counters[i] >= i; k++ {
			counters[i], i = 0, i+1
		}
		if i == amps {
			return best, nil
		}
		if i%2 == 0 {
			phases[0], phases[i] = phases[i], phases[0]
		} else {
			phases[counters[i]], phases[i] = phases[i], phases[counters[i]]
		}
		counters[i]++
		i = 0
	}
	return 0, errors.New("the orderings did not run out")
}

func day7(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	var machines [amps]machine
	for i := range machines {
		machines[i].mem = make([]int, memory)
	}
	best, err := bestSignal(&machines, program, 0)
	if err != nil {
		return err
	}

	feedback, err := bestSignal(&machines, program, 5)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", best, feedback); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
