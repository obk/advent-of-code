// Day 15: Beverage Bandits
//
// The input is a cave map: '#' walls, '.' open floor, 'G' goblins and 'E'
// elves, each unit with 200 hit points and 3 attack power. Every round each
// unit, in reading order, moves one step towards the nearest square next to
// an enemy (ties in reading order, first step in reading order) unless it is
// already next to one, and then hits the adjacent enemy with the fewest hit
// points (ties in reading order). Combat ends when a unit finds no enemies.
//
// Part 1 is the outcome: full rounds completed times the hit points left.
// Part 2 gives the elves the lowest attack power (at least 4) that lets them
// win with no elf dying, and asks for that battle's outcome.
//
// The cave is a flat array indexed y*width + x, so reading order is index
// order, and the four neighbours in reading order are up, left, right, down.
// Moving uses two breadth-first searches over a fixed queue: one from the
// unit to find the nearest target square, one back from that square to pick
// the first step.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 16
	maxCells      = 64 * 64
	maxUnits      = 64
	maxRounds     = 1_000
	startHP       = 200
)

// unit is one goblin or elf.
type unit struct {
	pos, hp, power int
	elf            bool
}

// battle is the cave, the units, and scratch space for the searches.
type battle struct {
	walls      [maxCells]bool
	occupant   [maxCells]int // unit index, -1 if empty
	units      [maxUnits]unit
	n, width   int
	order      [maxUnits]int
	dist, back [maxCells]int
	queue      [maxCells]int
}

// load parses the map and gives the elves the given attack power.
func (b *battle) load(text string, elfPower int) error {
	rows := strings.Split(strings.TrimSpace(text), "\n")
	if len(rows) < 3 {
		return fmt.Errorf("map of %d rows is too small", len(rows))
	}
	b.width = len(strings.TrimSpace(rows[0]))
	if b.width == 0 || b.width*len(rows) > maxCells || elfPower < 1 {
		return fmt.Errorf("map of %dx%d (elf power %d) is not supported", b.width, len(rows), elfPower)
	}
	b.n = 0
	for y, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) != b.width {
			return fmt.Errorf("row %d is %d wide, expected %d", y, len(row), b.width)
		}
		for x := range b.width {
			i := y*b.width + x
			b.walls[i], b.occupant[i] = row[x] == '#', -1
			if row[x] != 'G' && row[x] != 'E' {
				continue
			}
			if b.n == maxUnits {
				return fmt.Errorf("more than %d units", maxUnits)
			}
			b.units[b.n] = unit{pos: i, hp: startHP, power: 3, elf: row[x] == 'E'}
			if row[x] == 'E' {
				b.units[b.n].power = elfPower
			}
			b.occupant[i] = b.n
			b.n++
		}
	}
	elves := 0
	for i := range b.n {
		if b.units[i].elf {
			elves++
		}
	}
	if elves == 0 || elves == b.n {
		return fmt.Errorf("%d elves among %d units: nobody to fight", elves, b.n)
	}
	return nil
}

// neighbours returns the four cells around i in reading order (some may be
// off the map, which the caller must check with open).
func (b *battle) neighbours(i int) [4]int {
	return [4]int{i - b.width, i - 1, i + 1, i + b.width}
}

// open reports whether a unit could stand on cell i.
func (b *battle) open(i int) bool {
	return i >= 0 && i < maxCells && !b.walls[i] && b.occupant[i] < 0
}

// search fills dist with the step count from start over open cells (-1 where
// unreachable).
func (b *battle) search(start int, dist *[maxCells]int) error {
	if start < 0 || start >= maxCells || dist == nil {
		return fmt.Errorf("cannot search from %d", start)
	}
	if b.walls[start] {
		return fmt.Errorf("cannot search from wall %d", start)
	}
	for i := range dist {
		dist[i] = -1
	}
	dist[start], b.queue[0] = 0, start
	tail := 1
	for head := 0; head < tail; head++ {
		for _, nb := range b.neighbours(b.queue[head]) {
			if !b.open(nb) || dist[nb] >= 0 {
				continue
			}
			if tail == len(b.queue) {
				return errors.New("search queue overflowed")
			}
			dist[nb], b.queue[tail] = dist[b.queue[head]]+1, nb
			tail++
		}
	}
	return nil
}

// adjacentEnemy reports whether unit u has an enemy next to it.
func (b *battle) adjacentEnemy(u int) (bool, error) {
	if u < 0 || u >= b.n {
		return false, fmt.Errorf("unit %d does not exist", u)
	}
	for _, nb := range b.neighbours(b.units[u].pos) {
		if nb >= 0 && nb < maxCells && b.occupant[nb] >= 0 && b.units[b.occupant[nb]].elf != b.units[u].elf {
			return true, nil
		}
	}
	return false, nil
}

// step moves unit u one square towards the nearest square in range of an
// enemy, if it can reach one.
func (b *battle) step(u int) error {
	if u < 0 || u >= b.n {
		return fmt.Errorf("unit %d does not exist", u)
	}
	if b.units[u].hp <= 0 {
		return fmt.Errorf("unit %d is dead and cannot move", u)
	}
	if err := b.search(b.units[u].pos, &b.dist); err != nil {
		return err
	}
	target := -1
	for e := range b.n {
		if b.units[e].hp <= 0 || b.units[e].elf == b.units[u].elf {
			continue
		}
		for _, nb := range b.neighbours(b.units[e].pos) {
			if !b.open(nb) || b.dist[nb] < 0 {
				continue
			}
			if target < 0 || b.dist[nb] < b.dist[target] || (b.dist[nb] == b.dist[target] && nb < target) {
				target = nb
			}
		}
	}
	if target < 0 {
		return nil // nothing reachable: stay put
	}
	if err := b.search(target, &b.back); err != nil {
		return err
	}
	from, best := b.units[u].pos, -1
	for _, nb := range b.neighbours(from) {
		if b.open(nb) && b.back[nb] >= 0 && (best < 0 || b.back[nb] < b.back[best]) {
			best = nb
		}
	}
	if best < 0 {
		return errors.New("a reachable target has no first step")
	}
	if b.back[best] != b.dist[target]-1 {
		return fmt.Errorf("first step is %d from the target, expected %d", b.back[best], b.dist[target]-1)
	}
	b.occupant[from], b.occupant[best], b.units[u].pos = -1, u, best
	return nil
}

// attack hits the weakest adjacent enemy of unit u, if any. It reports
// whether an elf died.
func (b *battle) attack(u int) (bool, error) {
	if u < 0 || u >= b.n {
		return false, fmt.Errorf("unit %d does not exist", u)
	}
	victim := -1
	for _, nb := range b.neighbours(b.units[u].pos) {
		if nb < 0 || nb >= maxCells || b.occupant[nb] < 0 {
			continue
		}
		e := b.occupant[nb]
		if b.units[e].elf != b.units[u].elf && (victim < 0 || b.units[e].hp < b.units[victim].hp) {
			victim = e
		}
	}
	if victim < 0 {
		return false, nil
	}
	if b.units[u].power < 1 {
		return false, fmt.Errorf("unit %d has no attack power", u)
	}
	if b.units[victim].hp <= 0 {
		return false, fmt.Errorf("unit %d attacked the dead unit %d", u, victim)
	}
	b.units[victim].hp -= b.units[u].power
	if b.units[victim].hp > 0 {
		return false, nil
	}
	b.occupant[b.units[victim].pos] = -1
	return b.units[victim].elf, nil
}

// round plays one round. It reports whether combat ended during it and
// whether any elf died.
func (b *battle) round() (ended, elfDied bool, err error) {
	if b.n > maxUnits {
		return false, false, fmt.Errorf("%d units, more than %d", b.n, maxUnits)
	}
	alive := 0
	for i := range b.n {
		if b.units[i].hp > 0 {
			b.order[alive], alive = i, alive+1
		}
	}
	for i := 1; i < alive; i++ { // insertion sort into reading order
		for j := i; j > 0 && b.units[b.order[j]].pos < b.units[b.order[j-1]].pos; j-- {
			b.order[j], b.order[j-1] = b.order[j-1], b.order[j]
		}
	}
	for _, u := range b.order[:alive] {
		if b.units[u].hp <= 0 {
			continue
		}
		enemies := 0
		for e := range b.n {
			if b.units[e].hp > 0 && b.units[e].elf != b.units[u].elf {
				enemies++
			}
		}
		if enemies == 0 {
			return true, elfDied, nil
		}
		adjacent, err := b.adjacentEnemy(u)
		if err != nil {
			return false, false, err
		}
		if !adjacent {
			if err := b.step(u); err != nil {
				return false, false, err
			}
		}
		died, err := b.attack(u)
		if err != nil {
			return false, false, err
		}
		elfDied = elfDied || died
	}
	return false, elfDied, nil
}

// fight plays to the end and returns the outcome and whether an elf died.
func (b *battle) fight() (outcome int, elfDied bool, err error) {
	for rounds := 0; rounds < maxRounds; rounds++ {
		ended, died, err := b.round()
		if err != nil {
			return 0, false, err
		}
		elfDied = elfDied || died
		if ended {
			hp := 0
			for i := range b.n {
				hp += max(b.units[i].hp, 0)
			}
			if hp < 1 {
				return 0, false, errors.New("combat ended with nobody left standing")
			}
			return rounds * hp, elfDied, nil
		}
	}
	return 0, false, fmt.Errorf("combat lasted more than %d rounds", maxRounds)
}

func day15(path string) error {
	if path == "" {
		return errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	b := new(battle)
	if err := b.load(string(content), 3); err != nil {
		return err
	}
	outcome, _, err := b.fight()
	if err != nil {
		return err
	}
	flawless := -1
	for power := 4; power <= startHP && flawless < 0; power++ {
		if err := b.load(string(content), power); err != nil {
			return err
		}
		result, elfDied, err := b.fight()
		if err != nil {
			return err
		}
		if !elfDied {
			flawless = result
		}
	}
	if flawless < 0 {
		return errors.New("part 2: no attack power saves every elf")
	}
	if outcome < 1 {
		return fmt.Errorf("part 1 outcome %d is not positive", outcome)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", outcome, flawless); err != nil {
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
