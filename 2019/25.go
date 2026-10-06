// Day 25: Cryostasis
//
// The input is an Intcode text adventure. A droid explores a ship by typed
// commands (north/south/east/west, take/drop <item>) and must stand on a
// pressure-sensitive floor carrying exactly the right weight; the airlock
// then gives its password.
//
// Part 1 is the airlock password. (Day 25 has no second puzzle.)
//
// The droid explores depth-first, keeping a stack of the doors it came
// through so it can back out, and picks up every item except the known
// traps (they kill the droid or stop the game). It remembers the way to the
// security checkpoint and which door leads to the floor. At the checkpoint
// it tries every set of items, in Gray-code order so that each try changes
// one item.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 13
	maxQueue      = 1 << 8
	maxSteps      = 100_000_000
	maxText       = 1 << 12
	maxRooms      = 64
	maxItems      = 16
	maxName       = 40
	maxMoves      = 4 * maxRooms
)

// The four directions; d^1 is the opposite of d.
const directions = "north south west east "

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

// frame is a room on the exploration stack: its doors, the doors tried,
// and the direction back (-1 at the start).
type frame struct {
	doors, tried, back int
}

// game is the droid, its last output, and what it has learned.
type game struct {
	m        machine
	text     [maxText]byte
	n        int
	cmd      [maxName + 8]byte
	items    [maxItems][maxName]byte
	itemLen  [maxItems]int
	held     int
	stack    [maxRooms]frame
	depth    int
	path     [maxRooms]int
	pathLen  int
	floorDir int
}

// dirName returns direction d's command.
func dirName(d int) ([]byte, error) {
	if d < 0 || d > 3 {
		return nil, fmt.Errorf("direction %d does not exist", d)
	}
	f := strings.Fields(directions)
	if len(f) != 4 {
		return nil, fmt.Errorf("%d direction names, expected 4", len(f))
	}
	return []byte(f[d]), nil
}

// send types a command (none at the start) and reads the reply until the
// game asks for the next command or halts.
func (g *game) send(cmd []byte) (done bool, err error) {
	if len(cmd) > maxQueue-1 {
		return false, fmt.Errorf("command %q is too long", cmd)
	}
	for _, c := range cmd {
		if err := g.m.push(int(c)); err != nil {
			return false, err
		}
	}
	if len(cmd) > 0 {
		if err := g.m.push('\n'); err != nil {
			return false, err
		}
	}
	g.n = 0
	for range maxText + 1 {
		status, out, err := g.m.run()
		if err != nil {
			return false, err
		}
		if status != gaveOutput {
			return status == halted, nil
		}
		if g.n == maxText || out < 0 || out > 127 {
			return false, fmt.Errorf("reply too long or not ASCII (%d)", out)
		}
		g.text[g.n], g.n = byte(out), g.n+1
	}
	return false, errors.New("the reply never ended")
}

// room returns the last room description in the reply, from its "== "
// header on.
func (g *game) room() ([]byte, error) {
	at := bytes.LastIndex(g.text[:g.n], []byte("== "))
	if at < 0 {
		return nil, fmt.Errorf("no room in reply %q", g.text[:g.n])
	}
	if !bytes.Contains(g.text[at+3:g.n], []byte(" ==")) {
		return nil, fmt.Errorf("room header is not closed in %q", g.text[at:g.n])
	}
	return g.text[at:g.n], nil
}

// list returns the "- " lines after a heading in a room description, one
// call per index; ok is false past the end of the list.
func list(room []byte, heading string, index int) (line []byte, ok bool) {
	at := bytes.Index(room, []byte(heading))
	if at < 0 || index < 0 {
		return nil, false
	}
	rest := room[at+len(heading):]
	for i := 0; i <= index && i < maxItems; i++ {
		rest = bytes.TrimLeft(rest, "\n")
		end := bytes.IndexByte(rest, '\n')
		if !bytes.HasPrefix(rest, []byte("- ")) || end < 0 {
			return nil, false
		}
		line, rest = rest[2:end], rest[end:]
	}
	return line, true
}

// doors returns the room's doors as a bit mask.
func doors(room []byte) (int, error) {
	mask := 0
	for i := range 4 {
		door, ok := list(room, "Doors here lead:", i)
		if !ok {
			break
		}
		d := strings.Index(directions, string(door)+" ")
		if d < 0 || len(door) < 4 {
			return 0, fmt.Errorf("unknown door %q", door)
		}
		mask |= 1 << strings.Count(directions[:d], " ")
	}
	if mask == 0 {
		return 0, errors.New("the room has no doors")
	}
	return mask, nil
}

// itemCommand writes "<verb> <item i>" into the command buffer.
func (g *game) itemCommand(verb string, i int) ([]byte, error) {
	if verb != "take" && verb != "drop" {
		return nil, fmt.Errorf("unknown verb %q", verb)
	}
	if i < 0 || i >= g.held || len(verb)+1+g.itemLen[i] > len(g.cmd) {
		return nil, fmt.Errorf("cannot %s item %d", verb, i)
	}
	n := copy(g.cmd[:], verb)
	g.cmd[n] = ' '
	n += 1 + copy(g.cmd[n+1:], g.items[i][:g.itemLen[i]])
	return g.cmd[:n], nil
}

// takeAll picks up the safe items in the room just entered. The names are
// copied out first, since each command's reply replaces the room text.
func (g *game) takeAll(room []byte) error {
	traps := [5]string{"infinite loop", "giant electromagnet", "molten lava", "escape pod", "photons"}
	first := g.held
	for i := range maxItems {
		item, ok := list(room, "Items here:", i)
		if !ok {
			break
		}
		trap := false
		for _, t := range traps {
			trap = trap || string(item) == t
		}
		if trap {
			continue
		}
		if g.held == maxItems || len(item) > maxName {
			return fmt.Errorf("no room to carry %q", item)
		}
		for k := range g.held {
			if string(g.items[k][:g.itemLen[k]]) == string(item) {
				return fmt.Errorf("found a second %q", item)
			}
		}
		g.itemLen[g.held] = copy(g.items[g.held][:], item)
		g.held++
	}
	for i := first; i < g.held; i++ {
		cmd, err := g.itemCommand("take", i)
		if err != nil {
			return err
		}
		if _, err := g.send(cmd); err != nil {
			return err
		}
		if !bytes.Contains(g.text[:g.n], []byte("You take")) {
			return fmt.Errorf("could not take %q", cmd)
		}
	}
	return nil
}

// enter records the room just entered (from direction back) and takes its
// items.
func (g *game) enter(back int) error {
	if back < -1 || back > 3 {
		return fmt.Errorf("direction back %d does not exist", back)
	}
	if g.depth == maxRooms {
		return fmt.Errorf("more than %d rooms deep", maxRooms)
	}
	room, err := g.room()
	if err != nil {
		return err
	}
	mask, err := doors(room)
	if err != nil {
		return err
	}
	g.stack[g.depth], g.depth = frame{doors: mask, back: back}, g.depth+1
	if bytes.HasPrefix(room, []byte("== Security Checkpoint ==")) {
		// The way here is the reverse of each room's way back.
		g.pathLen = g.depth - 1
		for i := 1; i < g.depth; i++ {
			g.path[i-1] = g.stack[i].back ^ 1
		}
	}
	return g.takeAll(room)
}

// move walks one way and reports whether the floor threw the droid back.
func (g *game) move(d int) (bounced bool, err error) {
	name, err := dirName(d)
	if err != nil {
		return false, err
	}
	halted, err := g.send(name)
	if err != nil || halted {
		return false, fmt.Errorf("moving %s ended the game (%v): %q", name, err, g.text[:g.n])
	}
	return bytes.Contains(g.text[:g.n], []byte("Pressure-Sensitive Floor")), nil
}

// explore visits every room and comes back to the start.
func (g *game) explore() error {
	if _, err := g.send(nil); err != nil {
		return err
	}
	if err := g.enter(-1); err != nil {
		return err
	}
	g.floorDir = -1
	for range maxMoves {
		if g.depth < 1 {
			return errors.New("the exploration stack is empty")
		}
		f := &g.stack[g.depth-1]
		d := 4
		for k := 3; k >= 0; k-- {
			if (f.doors&^f.tried)>>k&1 == 1 && k != f.back {
				d = k
			}
		}
		if d == 4 { // nothing left here: back out, or finish at the start
			if g.depth == 1 && g.pathLen == 0 {
				return errors.New("explored everywhere without finding the checkpoint")
			}
			if g.depth == 1 {
				return nil
			}
			if _, err := g.move(f.back); err != nil {
				return err
			}
			g.depth--
			continue
		}
		f.tried |= 1 << d
		bounced, err := g.move(d)
		if err != nil {
			return err
		}
		if bounced {
			g.floorDir = d
			continue
		}
		if err := g.enter(d ^ 1); err != nil {
			return err
		}
	}
	return fmt.Errorf("exploring took more than %d moves", maxMoves)
}

// password walks to the checkpoint and tries sets of items on the floor
// until the airlock opens, returning its password.
func (g *game) password() (string, error) {
	if g.pathLen == 0 || g.floorDir < 0 || g.held == 0 {
		return "", errors.New("never found the checkpoint, its floor, or any items")
	}
	for _, d := range g.path[:g.pathLen] {
		bounced, err := g.move(d)
		if err != nil {
			return "", err
		}
		if bounced {
			return "", errors.New("walked onto the floor on the way to the checkpoint")
		}
	}
	if g.held > maxItems {
		return "", fmt.Errorf("%d items is too many to try every set", g.held)
	}
	all := 1<<g.held - 1
	for i := range 1 << g.held {
		if i > 0 { // Gray code: exactly one item changes from the last try
			changed := (i ^ i>>1) ^ ((i - 1) ^ (i-1)>>1)
			item := strings.Count(strconv.FormatInt(int64(changed-1), 2), "1")
			verb := "drop"
			if (all^(i^i>>1))>>item&1 == 1 { // the item is in the next set
				verb = "take"
			}
			cmd, err := g.itemCommand(verb, item)
			if err != nil {
				return "", err
			}
			if _, err := g.send(cmd); err != nil {
				return "", err
			}
		}
		name, err := dirName(g.floorDir)
		if err != nil {
			return "", err
		}
		halted, err := g.send(name)
		if err != nil {
			return "", err
		}
		if halted {
			return g.code()
		}
	}
	return "", errors.New("no set of items opens the airlock")
}

// code finds the number in "typing <number> on the keypad".
func (g *game) code() (string, error) {
	reply := g.text[:g.n]
	at := bytes.Index(reply, []byte("typing "))
	if at < 0 {
		return "", fmt.Errorf("no password in %q", reply)
	}
	digits := reply[at+len("typing "):]
	end := 0
	for k := 0; k < len(digits) && k < 32 && digits[k] >= '0' && digits[k] <= '9'; k++ {
		end++
	}
	if end == 0 {
		return "", fmt.Errorf("no password in %q", reply)
	}
	return string(digits[:end]), nil
}

func day25(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	g := &game{m: machine{mem: make([]int, memory)}}
	if err := g.m.load(program); err != nil {
		return err
	}
	if err := g.explore(); err != nil {
		return err
	}
	pass, err := g.password()
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %s\n", pass); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
