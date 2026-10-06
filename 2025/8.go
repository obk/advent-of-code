// Day 8: Playground
//
// The input is one junction box per line as "x,y,z". Every pair of boxes could
// be wired together, and the wires go in shortest first, each one joining the
// circuits of the two boxes it touches.
//
// Part 1 lets the 1000 shortest wires go in and multiplies the sizes of the
// three biggest circuits that result.
// Part 2 keeps going until every box sits on one circuit, and wants the
// product of the X coordinates of the two boxes the final wire joined.
//
// Which boxes share a circuit is tracked with a union-find, so joining two of
// them and asking whether a wire actually merged anything are both effectively
// constant time. Distances are only ever compared, never used as lengths, so
// they are kept squared and the square root is skipped - it would not change
// the ordering. All 499,500 pairs are generated up front and heapsorted in
// place, ties broken by box index so the order is fully determined, and the
// walk stops as soon as both answers are known.
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
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxBoxes      = 2_000
	maxCoord      = 1_000_000 // keeps every squared distance well inside int
)

const (
	// Part 1 looks at the circuits formed by the 1000 shortest connections.
	part1Connections = 1000
	// Part 1 multiplies the sizes of the three biggest circuits.
	largestCircuits = 3
)

// box is a junction box in 3D space.
type box struct {
	x, y, z int
}

// connection is a possible wire between two boxes, kept as indices into the
// box slice. Distances are compared squared, which avoids the square root
// without changing the ordering.
type connection struct {
	from, to int
	distSq   int
}

// unionFind tracks which boxes already sit on a common circuit.
type unionFind struct {
	parent   []int
	size     []int
	circuits int // how many separate circuits are left
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

// parseBox reads a box written as "x,y,z".
func parseBox(line string) (box, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 3 {
		return box{}, fmt.Errorf("malformed coordinate %q", line)
	}

	var coords [3]int
	for i, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			return box{}, fmt.Errorf("invalid coordinate %q: %w", line, err)
		}
		if value < -maxCoord || value > maxCoord {
			return box{}, fmt.Errorf("coordinate in %q is outside ±%d", line, maxCoord)
		}
		coords[i] = value
	}
	return box{x: coords[0], y: coords[1], z: coords[2]}, nil
}

func parseInput(path string) ([]box, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	boxes := make([]box, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b, err := parseBox(line)
		if err != nil {
			return nil, err
		}
		boxes = append(boxes, b)
	}

	if len(boxes) < 2 {
		return nil, fmt.Errorf("%s: need at least two boxes, found %d", path, len(boxes))
	}
	if len(boxes) > maxBoxes {
		return nil, fmt.Errorf("%s: %d boxes, more than the %d allowed", path, len(boxes), maxBoxes)
	}
	return boxes, nil
}

func newUnionFind(n int) (*unionFind, error) {
	if n < 1 || n > maxBoxes {
		return nil, fmt.Errorf("cannot track %d boxes, expected 1..%d", n, maxBoxes)
	}

	u := &unionFind{
		parent:   make([]int, n),
		size:     make([]int, n),
		circuits: n,
	}
	for i := range u.parent {
		u.parent[i] = i
		u.size[i] = 1
	}
	return u, nil
}

// find returns the representative of i's circuit, halving the path to it on
// the way so later lookups are shorter. No path is longer than the number of
// boxes, so that bounds the walk.
func (u *unionFind) find(i int) (int, error) {
	if i < 0 || i >= len(u.parent) {
		return 0, fmt.Errorf("box %d is outside 0..%d", i, len(u.parent)-1)
	}

	for range len(u.parent) {
		if u.parent[i] < 0 || u.parent[i] >= len(u.parent) {
			return 0, fmt.Errorf("box %d has parent %d outside the union-find", i, u.parent[i])
		}
		if u.parent[i] == i {
			return i, nil
		}
		u.parent[i] = u.parent[u.parent[i]] // path halving
		i = u.parent[i]
	}
	if u.parent[i] != i {
		return 0, errors.New("union-find has a cycle")
	}
	return i, nil
}

// union joins the circuits of i and j, reporting whether they were separate.
func (u *unionFind) union(i, j int) (bool, error) {
	if i == j {
		return false, fmt.Errorf("cannot wire box %d to itself", i)
	}
	rootI, err := u.find(i)
	if err != nil {
		return false, err
	}
	rootJ, err := u.find(j)
	if err != nil {
		return false, err
	}
	if rootI == rootJ {
		return false, nil
	}

	// Hang the smaller circuit off the bigger one.
	if u.size[rootI] < u.size[rootJ] {
		rootI, rootJ = rootJ, rootI
	}
	u.parent[rootJ] = rootI
	u.size[rootI] += u.size[rootJ]
	u.circuits--

	if u.circuits < 1 {
		return false, errors.New("merged more circuits than there were boxes")
	}
	if u.size[rootI] > len(u.parent) {
		return false, fmt.Errorf("circuit of %d boxes is bigger than the %d boxes", u.size[rootI], len(u.parent))
	}
	return true, nil
}

// productOfLargestCircuits multiplies the sizes of the largest circuits, at
// most largestCircuits of them. Boxes no wire has reached yet are circuits of
// one, so they take part too - they just never make it into the largest few.
func productOfLargestCircuits(u *unionFind) (int, error) {
	if len(u.parent) == 0 {
		return 0, errors.New("no circuits to measure")
	}
	var largest [largestCircuits]int // descending; zero means not filled yet
	boxes := 0
	for i := range u.parent {
		if u.parent[i] != i {
			continue // only representatives carry their circuit's size
		}
		boxes += u.size[i]
		size := u.size[i]
		for k := range largest {
			if size > largest[k] {
				largest[k], size = size, largest[k]
			}
		}
	}

	if boxes != len(u.parent) {
		return 0, fmt.Errorf("circuits hold %d boxes, expected %d", boxes, len(u.parent))
	}
	product := 1
	for _, size := range largest {
		if size > 0 {
			product *= size
		}
	}
	if product < 1 || product > boxes*boxes*boxes {
		return 0, fmt.Errorf("circuit product %d is out of range", product)
	}
	return product, nil
}

// less orders connections shortest first, then by box index, so that equal
// distances still come out in one fixed order.
func less(a, b connection) bool {
	if a.distSq != b.distSq {
		return a.distSq < b.distSq
	}
	if a.from != b.from {
		return a.from < b.from
	}
	return a.to < b.to
}

// siftDown restores the max-heap below root, treating c[:end] as the heap.
// Each step moves one level down, so the loop is bounded by the heap's depth.
func siftDown(c []connection, root, end int) error {
	if end < 0 || end > len(c) {
		return fmt.Errorf("heap end %d is outside 0..%d", end, len(c))
	}
	if root < 0 || root >= end {
		return fmt.Errorf("heap root %d is outside 0..%d", root, end-1)
	}

	for child := 2*root + 1; child < end; child = 2*root + 1 {
		if child+1 < end && less(c[child], c[child+1]) {
			child++
		}
		if !less(c[root], c[child]) {
			return nil
		}
		c[root], c[child] = c[child], c[root]
		root = child
	}
	return nil
}

// heapSort sorts the connections in place without recursion or allocation.
func heapSort(c []connection) error {
	if len(c) > maxBoxes*maxBoxes/2 {
		return fmt.Errorf("%d connections is too many to sort", len(c))
	}

	for root := len(c)/2 - 1; root >= 0; root-- {
		if err := siftDown(c, root, len(c)); err != nil {
			return err
		}
	}
	for end := len(c) - 1; end > 0; end-- {
		c[0], c[end] = c[end], c[0]
		if err := siftDown(c, 0, end); err != nil {
			return err
		}
	}

	for i := 1; i < len(c); i++ {
		if less(c[i], c[i-1]) {
			return fmt.Errorf("connections out of order at %d after sorting", i)
		}
	}
	return nil
}

// sortedConnections lists every pair of boxes, shortest first.
func sortedConnections(boxes []box) ([]connection, error) {
	if len(boxes) < 2 || len(boxes) > maxBoxes {
		return nil, fmt.Errorf("%d boxes, expected 2..%d", len(boxes), maxBoxes)
	}

	pairs := len(boxes) * (len(boxes) - 1) / 2
	connections := make([]connection, 0, pairs)
	for i := range boxes {
		for j := i + 1; j < len(boxes); j++ {
			a, b := boxes[i], boxes[j]
			dx, dy, dz := a.x-b.x, a.y-b.y, a.z-b.z
			distSq := dx*dx + dy*dy + dz*dz
			if distSq < 0 {
				return nil, fmt.Errorf("distance between boxes %d and %d overflows", i, j)
			}
			connections = append(connections, connection{from: i, to: j, distSq: distSq})
		}
	}

	if len(connections) != pairs {
		return nil, fmt.Errorf("built %d connections, expected %d", len(connections), pairs)
	}
	if err := heapSort(connections); err != nil {
		return nil, err
	}
	return connections, nil
}

// wireUp adds connections shortest first and reports both answers: the circuit
// product once part1Connections wires are in place, and the connection that
// finally leaves everything on one circuit.
func wireUp(u *unionFind, connections []connection) (product int, closing connection, err error) {
	if len(connections) < part1Connections {
		return 0, connection{}, fmt.Errorf("only %d connections, part 1 needs %d", len(connections), part1Connections)
	}
	if u.circuits != len(u.parent) {
		return 0, connection{}, errors.New("union-find already has wires in it")
	}

	closed := false
	for i, c := range connections {
		// A wire whose ends already share a circuit changes nothing, but it
		// still counts towards part 1's tally of wires laid.
		merged, err := u.union(c.from, c.to)
		if err != nil {
			return 0, connection{}, err
		}

		if i == part1Connections-1 {
			if product, err = productOfLargestCircuits(u); err != nil {
				return 0, connection{}, err
			}
		}
		// Only the wire that merges the last two circuits can leave one
		// behind, so this fires exactly once.
		if merged && u.circuits == 1 {
			closing, closed = c, true
		}

		// Nothing later can change either answer.
		if i >= part1Connections-1 && closed {
			break
		}
	}

	if !closed {
		return 0, connection{}, errors.New("part 2: the boxes never end up on a single circuit")
	}
	return product, closing, nil
}

func day8(path string) error {
	boxes, err := parseInput(path)
	if err != nil {
		return err
	}
	connections, err := sortedConnections(boxes)
	if err != nil {
		return err
	}
	u, err := newUnionFind(len(boxes))
	if err != nil {
		return err
	}
	product, closing, err := wireUp(u, connections)
	if err != nil {
		return err
	}

	// Part 2 wants the X coordinates of the two boxes the final wire joined.
	if closing.from < 0 || closing.to >= len(boxes) || closing.from >= closing.to {
		return fmt.Errorf("closing connection %v does not join two boxes", closing)
	}
	from, to := boxes[closing.from], boxes[closing.to]

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", product, from.x*to.x); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day8(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
