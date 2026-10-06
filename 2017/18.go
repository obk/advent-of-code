// Day 18: Duet
//
// The input is a program for registers a-z: snd X plays a sound, set/add/
// mul/mod change a register, rcv X recovers the last sound played if X is
// not zero, and jgz X Y jumps by Y if X is greater than zero.
//
// Part 1 is the frequency recovered the first time rcv runs with a non-zero
// operand.
// Part 2 reads snd and rcv as send and receive: two copies of the program
// (register p set to 0 and 1) send values to each other through queues, and
// rcv waits for a value. Once both are waiting (or done), how many values
// did program 1 send?
//
// Operands are parsed once into a register number or a constant. In part 2
// each program runs until it has to wait, then the other takes over; the
// queues are fixed-size rings.
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
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxSteps      = 100_000_000
	queueSize     = 1 << 20
)

// operand is a register (0-25) or a constant.
type operand struct {
	isReg bool
	value int
}

// instruction is one parsed line.
type instruction struct {
	op   string
	x, y operand
}

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// parseOperand reads a register letter or an integer.
func parseOperand(text string) (operand, error) {
	if text == "" {
		return operand{}, errors.New("empty operand")
	}
	if len(text) == 1 && text[0] >= 'a' && text[0] <= 'z' {
		return operand{isReg: true, value: int(text[0] - 'a')}, nil
	}
	v, err := strconv.Atoi(text)
	if err != nil {
		return operand{}, fmt.Errorf("invalid operand %q: %w", text, err)
	}
	return operand{value: v}, nil
}

func parseInput(path string) ([]instruction, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	program := make([]instruction, 0, len(lines))
	for _, line := range lines {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		in := instruction{op: f[0]}
		two := strings.Contains(" set add mul mod jgz ", " "+in.op+" ")
		one := in.op == "snd" || in.op == "rcv"
		if (!two && !one) || (two && len(f) != 3) || (one && len(f) != 2) {
			return nil, fmt.Errorf("invalid instruction %q", line)
		}
		if in.x, err = parseOperand(f[1]); err != nil {
			return nil, err
		}
		if two {
			if in.y, err = parseOperand(f[2]); err != nil {
				return nil, err
			}
		}
		if in.op != "jgz" && in.op != "snd" && !in.x.isReg {
			return nil, fmt.Errorf("instruction %q must name a register", line)
		}
		program = append(program, in)
	}
	if len(program) == 0 {
		return nil, fmt.Errorf("%s: empty program", path)
	}
	if len(program) > maxLines {
		return nil, fmt.Errorf("%s: %d instructions, more than %d", path, len(program), maxLines)
	}
	return program, nil
}

// machine is one running copy of the program for part 2.
type machine struct {
	regs       [26]int
	pc         int
	queue      []int // values waiting to be received, as a ring
	head, size int
	sent       int
}

// recoverSound runs the program until the first rcv with a non-zero operand
// and returns the last sound played.
func recoverSound(program []instruction) (int, error) {
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	var regs [26]int
	pc, sound, played := 0, 0, false
	for range maxSteps {
		if pc < 0 || pc >= len(program) {
			return 0, errors.New("the program ended without recovering a sound")
		}
		in := program[pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = regs[in.x.value]
		}
		if in.y.isReg {
			y = regs[in.y.value]
		}
		switch in.op {
		case "snd":
			sound, played = x, true
		case "set":
			regs[in.x.value] = y
		case "add":
			regs[in.x.value] += y
		case "mul":
			regs[in.x.value] *= y
		case "mod":
			if y <= 0 {
				return 0, fmt.Errorf("mod by %d at %d", y, pc)
			}
			regs[in.x.value] = ((x % y) + y) % y
		case "rcv":
			if x != 0 && played {
				return sound, nil
			}
		case "jgz":
			if x > 0 {
				pc += y - 1
			}
		}
		pc++
	}
	return 0, fmt.Errorf("no recovery within %d steps", maxSteps)
}

// run executes m until it must wait for a value or leaves the program,
// sending into other's queue. It reports how many instructions it ran.
func (m *machine) run(program []instruction, other *machine) (int, error) {
	if m == other || len(m.queue) == 0 || len(other.queue) == 0 {
		return 0, errors.New("the machines are not set up")
	}
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	for steps := 0; steps < maxSteps; steps++ {
		if m.pc < 0 || m.pc >= len(program) {
			return steps, nil
		}
		in := program[m.pc]
		x, y := in.x.value, in.y.value
		if in.x.isReg {
			x = m.regs[in.x.value]
		}
		if in.y.isReg {
			y = m.regs[in.y.value]
		}
		switch in.op {
		case "snd":
			if other.size == len(other.queue) {
				return 0, fmt.Errorf("queue of %d values overflowed", len(other.queue))
			}
			other.queue[(other.head+other.size)%len(other.queue)] = x
			other.size++
			m.sent++
		case "rcv":
			if m.size == 0 {
				return steps, nil // wait here for the other program
			}
			m.regs[in.x.value] = m.queue[m.head]
			m.head, m.size = (m.head+1)%len(m.queue), m.size-1
		case "set":
			m.regs[in.x.value] = y
		case "add":
			m.regs[in.x.value] += y
		case "mul":
			m.regs[in.x.value] *= y
		case "mod":
			if y <= 0 {
				return 0, fmt.Errorf("mod by %d at %d", y, m.pc)
			}
			m.regs[in.x.value] = ((x % y) + y) % y
		case "jgz":
			if x > 0 {
				m.pc += y - 1
			}
		}
		m.pc++
	}
	return 0, fmt.Errorf("a program ran %d steps without waiting", maxSteps)
}

// duet runs both copies until neither can make progress and returns how
// many values program 1 sent.
func duet(program []instruction) (int, error) {
	if len(program) == 0 {
		return 0, errors.New("empty program")
	}
	p0 := &machine{queue: make([]int, queueSize)}
	p1 := &machine{queue: make([]int, queueSize)}
	p1.regs['p'-'a'] = 1
	for round := 0; round < maxSteps; round++ {
		ran0, err := p0.run(program, p1)
		if err != nil {
			return 0, err
		}
		ran1, err := p1.run(program, p0)
		if err != nil {
			return 0, err
		}
		if ran0 == 0 && ran1 == 0 {
			return p1.sent, nil // deadlock, or both finished
		}
	}
	return 0, errors.New("the duet did not end")
}

func day18(path string) error {
	program, err := parseInput(path)
	if err != nil {
		return err
	}
	freq, err := recoverSound(program)
	if err != nil {
		return err
	}
	sent, err := duet(program)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", freq, sent); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
