// Day 24: Immune System Simulator 20XX
//
// The input lists two armies, "Immune System:" and "Infection:", each a list
// of groups: "18 units each with 729 hit points (weak to fire; immune to
// cold) with an attack that does 8 radiation damage at initiative 10". Each
// fight, groups pick targets in order of effective power (units * damage,
// ties by initiative), each choosing the enemy it would hurt most (ties by
// the target's effective power, then initiative); then groups attack in
// order of initiative, killing whole units only. Immunity means no damage,
// weakness double.
//
// Part 1 is how many units the winning army has left.
// Part 2 finds the smallest boost to the immune system's attack damage that
// lets it win, and asks how many units it then has left. Some boosts lead to
// a stalemate where nobody can kill anybody; those are not wins.
//
// Groups live in a fixed array; damage types are bits in small masks. Both
// orderings are insertion sorts over the few group indices.
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
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 16
	maxGroups     = 32
	maxTypes      = 8
	maxFights     = 100_000
	maxBoost      = 100_000
)

// group is one group of units.
type group struct {
	units, hp, damage, initiative int
	attack                        int  // damage type, as an index
	weak, immune                  uint // damage types, as bits
	infection                     bool
}

// war is both armies and the scratch arrays for one fight.
type war struct {
	groups  [maxGroups]group
	n       int
	types   [maxTypes]string
	ntypes  int
	order   [maxGroups]int
	target  [maxGroups]int
	claimed [maxGroups]bool
}

// typeID returns the number of a damage type, numbering it if new.
func (w *war) typeID(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty damage type")
	}
	for i := range w.ntypes {
		if w.types[i] == name {
			return i, nil
		}
	}
	if w.ntypes == maxTypes {
		return 0, fmt.Errorf("more than %d damage types", maxTypes)
	}
	w.types[w.ntypes] = name
	w.ntypes++
	return w.ntypes - 1, nil
}

// parseTraits reads "weak to fire; immune to cold, slashing".
func (w *war) parseTraits(text string, g *group) error {
	if text == "" || g == nil {
		return errors.New("empty traits or no group")
	}
	for _, part := range strings.Split(text, "; ") {
		kind, list, found := strings.Cut(part, " to ")
		if !found || (kind != "weak" && kind != "immune") {
			return fmt.Errorf("malformed traits %q", text)
		}
		for _, name := range strings.Split(list, ", ") {
			t, err := w.typeID(name)
			if err != nil {
				return err
			}
			if kind == "weak" {
				g.weak |= 1 << t
			} else {
				g.immune |= 1 << t
			}
		}
	}
	return nil
}

// parseGroup reads one group line.
func (w *war) parseGroup(line string, infection bool) (group, error) {
	g := group{infection: infection}
	if open := strings.IndexByte(line, '('); open >= 0 {
		shut := strings.IndexByte(line, ')')
		if shut < open {
			return group{}, fmt.Errorf("unclosed traits in %q", line)
		}
		if err := w.parseTraits(line[open+1:shut], &g); err != nil {
			return group{}, err
		}
		line = line[:open] + line[shut+1:]
	}
	f := strings.Fields(line)
	if len(f) != 18 || f[1] != "units" || f[14] != "damage" {
		return group{}, fmt.Errorf("malformed group %q", line)
	}
	var n [4]int
	for i, k := range [4]int{0, 4, 12, 17} {
		v, err := strconv.Atoi(f[k])
		if err != nil || v < 1 {
			return group{}, fmt.Errorf("invalid number %q in %q", f[k], line)
		}
		n[i] = v
	}
	g.units, g.hp, g.damage, g.initiative = n[0], n[1], n[2], n[3]
	var err error
	g.attack, err = w.typeID(f[13])
	return g, err
}

// load parses both armies.
func (w *war) load(text string) error {
	infection := false
	w.n, w.ntypes = 0, 0
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "":
			continue
		case "Immune System:":
			infection = false
			continue
		case "Infection:":
			infection = true
			continue
		}
		if w.n == maxGroups {
			return fmt.Errorf("more than %d groups", maxGroups)
		}
		g, err := w.parseGroup(line, infection)
		if err != nil {
			return err
		}
		w.groups[w.n], w.n = g, w.n+1
	}
	if w.n < 2 {
		return fmt.Errorf("%d groups cannot fight", w.n)
	}
	return nil
}

// damageTo is what attacker a would deal defender d.
func (w *war) damageTo(a, d int) (int, error) {
	if a < 0 || d < 0 || a >= w.n || d >= w.n {
		return 0, fmt.Errorf("groups %d and %d do not both exist", a, d)
	}
	ga, gd := w.groups[a], w.groups[d]
	switch {
	case gd.immune>>ga.attack&1 == 1:
		return 0, nil
	case gd.weak>>ga.attack&1 == 1:
		return 2 * ga.units * ga.damage, nil
	}
	return ga.units * ga.damage, nil
}

// selectTargets fills w.target for every living group (-1 for none).
func (w *war) selectTargets() error {
	if w.n > maxGroups {
		return fmt.Errorf("%d groups, more than %d", w.n, maxGroups)
	}
	alive := 0
	for i := range w.n {
		w.target[i], w.claimed[i] = -1, false
		if w.groups[i].units > 0 {
			w.order[alive], alive = i, alive+1
		}
	}
	for i := 1; i < alive; i++ {
		for j := i; j > 0; j-- {
			a, b := w.groups[w.order[j-1]], w.groups[w.order[j]]
			pa, pb := a.units*a.damage, b.units*b.damage
			if pa > pb || (pa == pb && a.initiative > b.initiative) {
				break
			}
			w.order[j-1], w.order[j] = w.order[j], w.order[j-1]
		}
	}
	for _, a := range w.order[:alive] {
		best, bestDamage := -1, 0
		for d := range w.n {
			gd := w.groups[d]
			if gd.units <= 0 || gd.infection == w.groups[a].infection || w.claimed[d] {
				continue
			}
			dmg, err := w.damageTo(a, d)
			if err != nil {
				return err
			}
			if dmg == 0 {
				continue
			}
			stronger := false // whether d outranks best as a target
			if best >= 0 {
				gb := w.groups[best]
				pd, pb := gd.units*gd.damage, gb.units*gb.damage
				stronger = pd > pb || (pd == pb && gd.initiative > gb.initiative)
			}
			if best < 0 || dmg > bestDamage || (dmg == bestDamage && stronger) {
				best, bestDamage = d, dmg
			}
		}
		if best >= 0 && w.groups[best].infection == w.groups[a].infection {
			return fmt.Errorf("group %d chose its ally %d", a, best)
		}
		if best >= 0 {
			w.target[a], w.claimed[best] = best, true
		}
	}
	return nil
}

// attackAll lets every group attack its target in order of initiative and
// returns how many units died in all.
func (w *war) attackAll() (int, error) {
	for i := range w.n {
		w.order[i] = i
	}
	for i := 1; i < w.n; i++ {
		for j := i; j > 0 && w.groups[w.order[j]].initiative > w.groups[w.order[j-1]].initiative; j-- {
			w.order[j-1], w.order[j] = w.order[j], w.order[j-1]
		}
	}
	killed := 0
	for _, a := range w.order[:w.n] {
		d := w.target[a]
		if d < 0 || w.groups[a].units <= 0 {
			continue
		}
		if d >= w.n || w.groups[d].infection == w.groups[a].infection {
			return 0, fmt.Errorf("group %d targets an ally or nothing", a)
		}
		dmg, err := w.damageTo(a, d)
		if err != nil {
			return 0, err
		}
		dead := min(dmg/w.groups[d].hp, w.groups[d].units)
		if dead < 0 {
			return 0, fmt.Errorf("group %d would revive units in %d", a, d)
		}
		w.groups[d].units -= dead
		killed += dead
	}
	return killed, nil
}

// battle fights until one army is gone, with the immune system's damage
// raised by boost. It reports the winner and its units, or a stalemate.
func (w *war) battle(text string, boost int) (immuneWon bool, units int, stalemate bool, err error) {
	if err := w.load(text); err != nil {
		return false, 0, false, err
	}
	if boost < 0 {
		return false, 0, false, fmt.Errorf("boost %d is negative", boost)
	}
	armies := 0
	for i := range w.n {
		if !w.groups[i].infection {
			w.groups[i].damage += boost
			armies |= 1
		} else {
			armies |= 2
		}
	}
	if armies != 3 {
		return false, 0, false, errors.New("one of the armies has no groups")
	}
	for range maxFights {
		var immune, infection int
		for i := range w.n {
			if w.groups[i].infection {
				infection += w.groups[i].units
			} else {
				immune += w.groups[i].units
			}
		}
		if immune == 0 || infection == 0 {
			return immune > 0, immune + infection, false, nil
		}
		if err := w.selectTargets(); err != nil {
			return false, 0, false, err
		}
		killed, err := w.attackAll()
		if err != nil {
			return false, 0, false, err
		}
		if killed == 0 {
			return false, 0, true, nil // nobody can hurt anybody any more
		}
	}
	return false, 0, false, fmt.Errorf("the battle lasted more than %d fights", maxFights)
}

func day24(path string) error {
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
	w := new(war)
	_, units, stalemate, err := w.battle(string(content), 0)
	if err != nil {
		return err
	}
	if stalemate {
		return errors.New("part 1: the battle ends in a stalemate")
	}

	boosted := -1
	for boost := 1; boost <= maxBoost && boosted < 0; boost++ {
		immuneWon, left, stuck, err := w.battle(string(content), boost)
		if err != nil {
			return err
		}
		if immuneWon && !stuck {
			boosted = left
		}
	}
	if boosted < 0 {
		return fmt.Errorf("part 2: no boost up to %d lets the immune system win", maxBoost)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", units, boosted); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
