// Day 17: Set and Forget
//
// The input is an Intcode program that outputs a camera view of scaffolding
// as ASCII: '#' scaffold, '.' space, and the robot as ^ v < >.
//
// Part 1 sums the alignment parameters (x * y) of the scaffold
// intersections: scaffold cells with scaffold on all four sides.
// Part 2 wakes the robot (address 0 set to 2) and drives it over the whole
// scaffold. The route must be given as a main routine calling functions A,
// B and C, each line at most 20 characters; the answer is the dust the
// robot reports at the end (its last output, beyond ASCII).
//
// The route is found by walking straight until the scaffold ends, then
// turning whichever way it continues. It is split by trying every length
// for A (which must start the route), then B (starting at the first move A
// does not cover), then C, and checking that A, B and C cover the route.
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
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 20
	memory        = 1 << 13
	maxQueue      = 1 << 10
	maxSteps      = 100_000_000
	maxView       = 1 << 14
	maxMoves      = 64
	maxLine       = 20
	maxCalls      = 10
)

// move is one turn and the steps after it.
type move struct {
	turn  byte
	steps int
}

// route is the robot's moves, and a split of them into functions.
type route struct {
	moves  [maxMoves]move
	n      int
	length [3]int // moves in A, B and C
	start  [3]int // where each function's moves begin in the route
	calls  [maxCalls]int
	ncalls int
}

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

// view is the camera picture: rows of width cells, each row followed by a
// newline in buf.
type view struct {
	buf           [maxView]byte
	width, height int
}

// capture runs the program until it halts, recording its picture.
func (v *view) capture(m *machine) error {
	n := 0
	for i := 0; i <= maxView; i++ {
		status, out, err := m.run()
		if err != nil {
			return err
		}
		if status == halted {
			break
		}
		if status != gaveOutput || out < 0 || out > 127 || i == maxView {
			return fmt.Errorf("camera gave status %d, output %d, at %d", status, out, i)
		}
		v.buf[i], n = byte(out), i+1
	}
	for i := 0; i < maxView && n > 0 && v.buf[n-1] == '\n'; i++ { // trailing blank lines
		n--
	}
	for i := 0; i < n && v.buf[i] != '\n'; i++ {
		v.width++
	}
	if v.width == 0 || (n+1)%(v.width+1) != 0 {
		return fmt.Errorf("the picture of %d bytes is not rows of %d", n, v.width)
	}
	v.height = (n + 1) / (v.width + 1)
	if v.height < 3 || v.width < 3 {
		return fmt.Errorf("the picture is only %dx%d", v.width, v.height)
	}
	return nil
}

// at returns the cell at x, y, or '.' off the picture.
func (v *view) at(x, y int) byte {
	if x < 0 || y < 0 || x >= v.width || y >= v.height {
		return '.'
	}
	return v.buf[y*(v.width+1)+x]
}

// alignment sums x*y over the intersections.
func (v *view) alignment() (int, error) {
	if v.width < 1 || v.height < 1 {
		return 0, errors.New("empty picture")
	}
	total, found := 0, 0
	for y := range v.height {
		for x := range v.width {
			if v.at(x, y) == '#' && v.at(x-1, y) == '#' && v.at(x+1, y) == '#' && v.at(x, y-1) == '#' && v.at(x, y+1) == '#' {
				total, found = total+x*y, found+1
			}
		}
	}
	if found == 0 {
		return 0, errors.New("the scaffold never crosses itself")
	}
	return total, nil
}

// trace walks the scaffold from the robot and records the moves.
func (v *view) trace(r *route) error {
	if v.width == 0 || r == nil {
		return errors.New("no picture or no route to trace")
	}
	x, y, dir := -1, -1, 0 // dir: 0 up, 1 right, 2 down, 3 left
	for yy := range v.height {
		for xx := range v.width {
			d := strings.IndexByte("^>v<", v.at(xx, yy))
			if d >= 0 && x >= 0 {
				return errors.New("the picture shows two robots")
			}
			if d >= 0 {
				x, y, dir = xx, yy, d
			}
		}
	}
	if x < 0 {
		return errors.New("the robot is not in the picture")
	}
	dx, dy := [4]int{0, 1, 0, -1}, [4]int{-1, 0, 1, 0}
	for r.n = 0; r.n < maxMoves; r.n++ {
		left, right := (dir+3)%4, (dir+1)%4
		switch {
		case v.at(x+dx[left], y+dy[left]) == '#':
			r.moves[r.n].turn, dir = 'L', left
		case v.at(x+dx[right], y+dy[right]) == '#':
			r.moves[r.n].turn, dir = 'R', right
		default: // the end of the scaffold
			if r.n == 0 {
				return errors.New("the robot cannot move")
			}
			return nil
		}
		steps := 0
		for k := 0; k < v.width+v.height && v.at(x+dx[dir], y+dy[dir]) == '#'; k++ {
			x, y, steps = x+dx[dir], y+dy[dir], steps+1
		}
		r.moves[r.n].steps = steps
	}
	return fmt.Errorf("the route has more than %d moves", maxMoves)
}

// matches reports whether function f's moves appear at position at.
func (r *route) matches(f, at int) (bool, error) {
	if f < 0 || f > 2 || at < 0 || r.start[f] < 0 || r.start[f]+r.length[f] > r.n {
		return false, fmt.Errorf("function %d at %d is outside the route", f, at)
	}
	if r.length[f] == 0 || at+r.length[f] > r.n {
		return false, nil
	}
	for i := range r.length[f] {
		if r.moves[r.start[f]+i] != r.moves[at+i] {
			return false, nil
		}
	}
	return true, nil
}

// cover calls the first `funcs` functions greedily from the start, and
// returns where it got stuck (r.n if it covered the whole route).
func (r *route) cover(funcs int) (int, error) {
	if funcs < 1 || funcs > 3 {
		return 0, fmt.Errorf("%d functions; there are 1 to 3", funcs)
	}
	at := 0
	for r.ncalls = 0; r.ncalls < maxCalls && at < r.n; r.ncalls++ {
		f := -1
		for k := 0; k < funcs && f < 0; k++ {
			ok, err := r.matches(k, at)
			if err != nil {
				return 0, err
			}
			if ok {
				f = k
			}
		}
		if f < 0 {
			return at, nil
		}
		r.calls[r.ncalls], at = f, at+r.length[f]
	}
	if at > r.n {
		return 0, fmt.Errorf("calls ran to %d, past the route's %d moves", at, r.n)
	}
	return at, nil
}

// lineLength is how long function f is when written out.
func (r *route) lineLength(f int) (int, error) {
	if f < 0 || f > 2 || r.length[f] < 1 {
		return 0, fmt.Errorf("function %d is empty or does not exist", f)
	}
	n := r.length[f] - 1 // commas between moves
	for i := range r.length[f] {
		n += 2 + len(strconv.Itoa(r.moves[r.start[f]+i].steps)) // "L,8"
	}
	return n, nil
}

func day17(path string) error {
	program, err := parseProgram(path)
	if err != nil {
		return err
	}
	m := &machine{mem: make([]int, memory)}
	if err := m.load(program); err != nil {
		return err
	}
	v := new(view)
	if err := v.capture(m); err != nil {
		return err
	}
	sum, err := v.alignment()
	if err != nil {
		return err
	}

	r := new(route)
	if err := v.trace(r); err != nil {
		return err
	}
	if err := r.split(); err != nil {
		return err
	}
	if err := m.load(program); err != nil {
		return err
	}
	m.mem[0] = 2
	if err := r.program(m, make([]byte, 0, 4*maxLine)); err != nil {
		return err
	}
	collected, err := dust(m)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, collected); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// split finds lengths for A, B and C that cover the route.
func (r *route) split() error {
	if r.n == 0 {
		return errors.New("the route is empty")
	}
	var err error
	for a := 1; a <= 5; a++ {
		r.length, r.start = [3]int{a, 0, 0}, [3]int{}
		if r.start[1], err = r.cover(1); err != nil {
			return err
		}
		for b := 1; b <= 5 && r.start[1]+b <= r.n; b++ {
			r.length[1], r.length[2] = b, 0
			if r.start[2], err = r.cover(2); err != nil {
				return err
			}
			for c := 1; c <= 5 && r.start[2]+c <= r.n; c++ {
				r.length[2] = c
				fits, err := r.fits()
				if err != nil {
					return err
				}
				if fits {
					return nil
				}
			}
		}
	}
	return errors.New("no split into three functions fits")
}

// fits reports whether A, B and C cover the route and fit on their lines.
func (r *route) fits() (bool, error) {
	if r.start[0] != 0 || r.start[1] < r.start[0] || r.start[2] < r.start[1] {
		return false, fmt.Errorf("functions start out of order at %v", r.start)
	}
	end, err := r.cover(3)
	if err != nil || end != r.n {
		return false, err
	}
	for f := range 3 {
		n, err := r.lineLength(f)
		if err != nil {
			return false, err
		}
		if n > maxLine {
			return false, nil
		}
	}
	return true, nil
}

// pushLine sends text to the robot followed by a newline.
func pushLine(m *machine, text []byte) error {
	if len(text) > maxLine {
		return fmt.Errorf("line %q is longer than %d", text, maxLine)
	}
	for _, c := range text {
		if c < ' ' || c > '~' {
			return fmt.Errorf("line %q has a non-printing character", text)
		}
		if err := m.push(int(c)); err != nil {
			return err
		}
	}
	return m.push('\n')
}

// program sends the main routine, the three functions and "n" (no video).
func (r *route) program(m *machine, line []byte) error {
	if r.ncalls == 0 || cap(line) < 4*maxLine {
		return errors.New("no routine, or no room to write it")
	}
	if 2*r.ncalls-1 > maxLine {
		return fmt.Errorf("main routine of %d calls is too long", r.ncalls)
	}
	line = line[:0]
	for i, f := range r.calls[:r.ncalls] {
		if i > 0 {
			line = append(line, ',')
		}
		line = append(line, byte('A'+f))
	}
	if err := pushLine(m, line); err != nil {
		return err
	}
	for f := range 3 {
		line = line[:0]
		if r.start[f] < 0 || r.start[f]+r.length[f] > r.n {
			return fmt.Errorf("function %c runs outside the route", 'A'+f)
		}
		for i, mv := range r.moves[r.start[f] : r.start[f]+r.length[f]] {
			if i > 0 {
				line = append(line, ',')
			}
			line = append(line, mv.turn, ',')
			line = strconv.AppendInt(line, int64(mv.steps), 10)
		}
		if err := pushLine(m, line); err != nil {
			return err
		}
	}
	return pushLine(m, []byte{'n'})
}

// dust runs the woken robot to the end and returns its last output.
func dust(m *machine) (int, error) {
	last := -1
	for range maxView {
		status, out, err := m.run()
		if err != nil {
			return 0, err
		}
		if status == halted {
			if last <= 127 {
				return 0, fmt.Errorf("the robot ended with %d, not a dust count", last)
			}
			return last, nil
		}
		if status == needInput {
			return 0, errors.New("the robot asked for more input")
		}
		last = out
	}
	return 0, fmt.Errorf("the robot gave more than %d outputs", maxView)
}
