// Day 14: Space Stoichiometry
//
// Each line of the input is a reaction, "7 A, 1 B => 1 C": whole batches of
// the inputs make a batch of the output. Every chemical but ORE is made by
// exactly one reaction.
//
// Part 1 is the ORE needed for one FUEL.
// Part 2 is the most FUEL that a trillion ORE makes. The ORE needed grows
// with the FUEL made, so a binary search finds it.
//
// Chemicals are put in an order where every chemical comes before the ones
// it is made from (Kahn's algorithm on the reactions, with no recursion).
// Walking that order, each chemical's total need is known before it is
// made, so whole batches can be run and their inputs added to the needs.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 16
	maxChemicals  = 128
	maxInputs     = 16
	maxQuantity   = 1_000_000
	trillion      = 1_000_000_000_000
)

// term is a quantity of a chemical.
type term struct {
	chem, qty int
}

// reaction makes qty of its chemical from inputs.
type reaction struct {
	qty    int
	inputs []term
	known  bool
}

// factory is every reaction, indexed by the chemical it makes, and the
// order to resolve needs in.
type factory struct {
	ids       map[string]int
	reactions [maxChemicals]reaction
	order     [maxChemicals]int
	need      [maxChemicals]int
	ore, fuel int
}

// id returns a chemical's number, numbering it if new.
func (f *factory) id(name string) (int, error) {
	if name == "" {
		return 0, errors.New("empty chemical name")
	}
	if i, ok := f.ids[name]; ok {
		return i, nil
	}
	if len(f.ids) == maxChemicals {
		return 0, fmt.Errorf("more than %d chemicals", maxChemicals)
	}
	f.ids[name] = len(f.ids)
	return f.ids[name], nil
}

// parseTerm reads "7 A".
func (f *factory) parseTerm(text string) (term, error) {
	n, name, found := strings.Cut(strings.TrimSpace(text), " ")
	if name == "" || strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return term{}, fmt.Errorf("chemical name in %q is not capital letters", text)
	}
	qty, err := strconv.Atoi(n)
	if !found || err != nil || qty < 1 || qty > maxQuantity {
		return term{}, fmt.Errorf("malformed quantity %q", text)
	}
	c, err := f.id(name)
	if err != nil {
		return term{}, err
	}
	return term{c, qty}, nil
}

// addReaction parses one line.
func (f *factory) addReaction(line string) error {
	left, right, found := strings.Cut(line, "=>")
	if !found {
		return fmt.Errorf("malformed reaction %q", line)
	}
	out, err := f.parseTerm(right)
	if err != nil {
		return err
	}
	if strings.HasSuffix(strings.TrimSpace(right), " ORE") {
		return fmt.Errorf("reaction %q makes ORE", line)
	}
	r := &f.reactions[out.chem]
	if r.known {
		return fmt.Errorf("two reactions make chemical %q", right)
	}
	parts := strings.Split(left, ",")
	if len(parts) > maxInputs {
		return fmt.Errorf("reaction %q has more than %d inputs", line, maxInputs)
	}
	r.qty, r.known, r.inputs = out.qty, true, make([]term, len(parts))
	for i, p := range parts {
		if r.inputs[i], err = f.parseTerm(p); err != nil {
			return err
		}
	}
	return nil
}

// load parses the reactions.
func load(path string) (*factory, error) {
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
	f := &factory{ids: make(map[string]int, maxChemicals)}
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		if err := f.addReaction(line); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	ore, okOre := f.ids["ORE"]
	fuel, okFuel := f.ids["FUEL"]
	if !okOre || !okFuel || f.reactions[ore].known || !f.reactions[fuel].known {
		return nil, fmt.Errorf("%s: ORE must be raw and FUEL must be made", path)
	}
	f.ore, f.fuel = ore, fuel
	return f, f.sortChemicals()
}

// sortChemicals orders the chemicals so that each comes before everything
// it is made from.
func (f *factory) sortChemicals() error {
	n := len(f.ids)
	var users [maxChemicals]int
	for c := range n {
		for _, t := range f.reactions[c].inputs {
			users[t.chem]++
		}
	}
	tail := 0
	for c := range n {
		if users[c] == 0 {
			f.order[tail], tail = c, tail+1
		}
	}
	for head := 0; head < tail; head++ {
		for _, t := range f.reactions[f.order[head]].inputs {
			users[t.chem]--
			if users[t.chem] == 0 {
				f.order[tail], tail = t.chem, tail+1
			}
		}
	}
	if tail != n {
		return fmt.Errorf("the reactions have a cycle: ordered %d of %d chemicals", tail, n)
	}
	if f.order[n-1] != f.ore {
		return errors.New("ORE is not the last thing everything is made from")
	}
	return nil
}

// oreFor returns the ORE needed to make the given amount of FUEL.
func (f *factory) oreFor(fuel int) (int, error) {
	if fuel < 1 || fuel > trillion {
		return 0, fmt.Errorf("fuel amount %d is out of range", fuel)
	}
	clear(f.need[:])
	f.need[f.fuel] = fuel
	for _, c := range f.order[:len(f.ids)] {
		r := &f.reactions[c]
		if c == f.ore || f.need[c] == 0 {
			continue
		}
		batches := (f.need[c] + r.qty - 1) / r.qty
		for _, t := range r.inputs {
			if batches > (math.MaxInt64-f.need[t.chem])/t.qty {
				return 0, errors.New("the amounts overflowed")
			}
			f.need[t.chem] += batches * t.qty
		}
	}
	if f.need[f.ore] < 1 {
		return 0, errors.New("FUEL needed no ORE")
	}
	return f.need[f.ore], nil
}

// mostFuel finds the most FUEL the given ORE makes, knowing one FUEL needs
// orePerFuel.
func (f *factory) mostFuel(ore, orePerFuel int) (int, error) {
	if orePerFuel < 1 || ore < orePerFuel {
		return 0, fmt.Errorf("%d ORE cannot make even one FUEL at %d each", ore, orePerFuel)
	}
	// Leftovers only help, so ore/orePerFuel is always possible. Doubling it
	// is not, since no FUEL can take less than the ORE that goes into it on
	// average; the search confirms that.
	lo, hi := ore/orePerFuel, 2*(ore/orePerFuel)+1
	need, err := f.oreFor(hi)
	if err != nil {
		return 0, err
	}
	if need <= ore {
		return 0, fmt.Errorf("%d FUEL takes only %d ORE; the search bound is wrong", hi, need)
	}
	for i := 0; i < 64 && hi-lo > 1; i++ {
		mid := lo + (hi-lo)/2
		need, err := f.oreFor(mid)
		if err != nil {
			return 0, err
		}
		if need <= ore {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo, nil
}

func day14(path string) error {
	f, err := load(path)
	if err != nil {
		return err
	}
	ore, err := f.oreFor(1)
	if err != nil {
		return err
	}

	fuel, err := f.mostFuel(trillion, ore)
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", ore, fuel); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
