// Day 21: Allergen Assessment
//
// Each line of the input is a food: its ingredients, then some of the
// allergens it contains, "mxmxvkd kfcds (contains dairy, fish)". Each
// allergen is in exactly one ingredient, and each ingredient has at most
// one allergen.
//
// Part 1 counts appearances of ingredients that cannot hold any allergen.
// Part 2 lists the ingredients holding allergens, ordered by allergen name
// and joined with commas.
//
// An allergen's ingredient must be in every food listing it, so its
// candidates are the intersection of those foods' ingredient sets (bit
// sets). Allergens with one candidate left fix it and remove it from the
// others, until all are fixed.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 16
	maxFoods      = 100
	maxAllergens  = 16
	words         = 4 // 64-bit words per ingredient set
	maxNames      = 64 * words
)

// set is a set of ingredients.
type set [words]uint64

// count returns the set's size.
func (s set) count() int {
	n := 0
	for _, w := range s {
		n += bits.OnesCount64(w)
	}
	return n
}

// menu is the parsed input.
type menu struct {
	ingredients []string
	ids         map[string]int
	allergens   []string
	foods       []set
	appear      []int // how many foods each ingredient is in
	candidates  []set // per allergen
	sortedAll   []int // allergens in name order
}

// id returns a name's number in list, adding it if new.
func id(ids map[string]int, list []string, name string) (int, []string, error) {
	if name == "" {
		return 0, list, errors.New("empty name")
	}
	if i, ok := ids[name]; ok {
		return i, list, nil
	}
	if len(list) == cap(list) {
		return 0, list, fmt.Errorf("more than %d names", cap(list))
	}
	ids[name] = len(list)
	return len(list), append(list, name), nil
}

// addFood parses one food.
func (m *menu) addFood(line string, allergenIDs map[string]int) error {
	left, right, found := strings.Cut(strings.TrimSuffix(line, ")"), " (contains ")
	if !found {
		return fmt.Errorf("food %q lists no allergens", line)
	}
	if strings.TrimSpace(left) == "" || strings.TrimSpace(right) == "" {
		return fmt.Errorf("food %q has no ingredients or no allergens", line)
	}
	var food set
	for _, name := range strings.Fields(left) {
		i, list, err := id(m.ids, m.ingredients, name)
		if err != nil {
			return err
		}
		m.ingredients = list
		food[i/64] |= 1 << (i % 64)
		m.appear[i]++
	}
	for _, name := range strings.Split(right, ", ") {
		a, list, err := id(allergenIDs, m.allergens, name)
		if err != nil {
			return err
		}
		if a > len(m.candidates) {
			return fmt.Errorf("allergen %d skipped a number", a)
		}
		if a == len(m.candidates) { // first time: every ingredient
			m.candidates = append(m.candidates, food)
		}
		m.allergens = list
		for w := range words {
			m.candidates[a][w] &= food[w]
		}
	}
	if len(m.foods) == cap(m.foods) {
		return fmt.Errorf("more than %d foods", cap(m.foods))
	}
	m.foods = append(m.foods, food)
	return nil
}

// load parses the foods.
func load(path string) (*menu, error) {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxFoods {
		return nil, fmt.Errorf("%s: more than %d foods", path, maxFoods)
	}
	m := &menu{
		ingredients: make([]string, 0, maxNames),
		ids:         make(map[string]int, maxNames),
		allergens:   make([]string, 0, maxAllergens),
		foods:       make([]set, 0, len(lines)),
		appear:      make([]int, maxNames),
		candidates:  make([]set, 0, maxAllergens),
		sortedAll:   make([]int, 0, maxAllergens),
	}
	allergenIDs := make(map[string]int, maxAllergens)
	for _, line := range lines {
		if err := m.addFood(line, allergenIDs); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	if len(m.allergens) == 0 {
		return nil, fmt.Errorf("%s: no allergens", path)
	}
	return m, nil
}

// safe counts appearances of ingredients no allergen could be in.
func (m *menu) safe() (int, error) {
	var risky set
	for _, c := range m.candidates {
		if c.count() == 0 {
			return 0, errors.New("an allergen has no possible ingredient")
		}
		for w := range words {
			risky[w] |= c[w]
		}
	}
	total, all := 0, 0
	for i := range m.ingredients {
		all += m.appear[i]
		if risky[i/64]>>(i%64)&1 == 0 {
			total += m.appear[i]
		}
	}
	if total >= all {
		return 0, errors.New("no ingredient can hold an allergen")
	}
	return total, nil
}

// resolve fixes each allergen's ingredient, leaving one bit per candidate
// set.
func (m *menu) resolve() error {
	if len(m.candidates) != len(m.allergens) {
		return errors.New("candidate sets do not match the allergens")
	}
	var fixed [maxAllergens]bool
	for range m.allergens {
		for a, c := range m.candidates {
			if fixed[a] || c.count() != 1 {
				continue
			}
			fixed[a] = true
			for b := range m.candidates {
				if b == a {
					continue
				}
				for w := range words {
					m.candidates[b][w] &^= c[w]
				}
			}
		}
	}
	for a := range m.allergens {
		if !fixed[a] {
			return fmt.Errorf("allergen %q is not pinned down", m.allergens[a])
		}
	}
	return nil
}

// dangerous writes the ingredients by allergen name into out.
func (m *menu) dangerous(out []byte) ([]byte, error) {
	if len(m.allergens) > maxAllergens {
		return nil, errors.New("too many allergens")
	}
	// Insertion sort of allergen numbers by name.
	m.sortedAll = m.sortedAll[:0]
	for a := range m.allergens {
		m.sortedAll = append(m.sortedAll, a)
		for k := len(m.sortedAll) - 1; k > 0 && m.allergens[m.sortedAll[k-1]] > m.allergens[m.sortedAll[k]]; k-- {
			m.sortedAll[k-1], m.sortedAll[k] = m.sortedAll[k], m.sortedAll[k-1]
		}
	}
	for k := 1; k < len(m.sortedAll); k++ {
		if m.allergens[m.sortedAll[k-1]] >= m.allergens[m.sortedAll[k]] {
			return nil, errors.New("allergens are not in order")
		}
	}
	out = out[:0]
	for n, a := range m.sortedAll {
		c := m.candidates[a]
		i := -1
		for w := range words {
			if c[w] != 0 {
				i = w*64 + bits.TrailingZeros64(c[w])
			}
		}
		if i < 0 || len(out)+len(m.ingredients[i])+1 > cap(out) {
			return nil, errors.New("no ingredient, or no room to list it")
		}
		if n > 0 {
			out = append(out, ',')
		}
		out = append(out, m.ingredients[i]...)
	}
	return out, nil
}

func day21(path string) error {
	m, err := load(path)
	if err != nil {
		return err
	}
	count, err := m.safe()
	if err != nil {
		return err
	}
	if err := m.resolve(); err != nil {
		return err
	}
	list, err := m.dangerous(make([]byte, 0, maxInputBytes))
	if err != nil {
		return err
	}

	if count < 1 {
		return errors.New("every ingredient might hold an allergen")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", count, list); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
