// Day 15: Science for Hungry People
//
// Each line of the input lists an ingredient's properties per teaspoon:
// "Butterscotch: capacity -1, durability -2, flavor 6, texture 3, calories 8".
// A recipe uses exactly 100 teaspoons in all. Its score multiplies the totals
// of the four properties other than calories, with negative totals counted as
// zero.
//
// Part 1 is the best score of any recipe.
// Part 2 is the best score among recipes of exactly 500 calories.
//
// With at most four ingredients there are only about 177,000 ways to split
// 100 teaspoons, so every one is scored. The split is enumerated like an
// odometer: the amounts of all but the last ingredient count up in base 101,
// and the last ingredient takes whatever is left of the 100.
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
	inputPath      = "inputs/15.txt"
	maxInputBytes  = 1 << 20
	maxLines       = 100
	maxIngredients = 4
	maxProperty    = 100
	teaspoons      = 100
	targetCalories = 500
	properties     = 5 // capacity, durability, flavor, texture, calories
	calories       = 4 // the index of calories among the properties
)

type ingredient [properties]int

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

// parseIngredient reads the five property values, which follow the name in
// the order capacity, durability, flavor, texture, calories.
func parseIngredient(line string) (ingredient, error) {
	name, list, found := strings.Cut(line, ": ")
	if !found {
		return ingredient{}, fmt.Errorf("malformed line %q", line)
	}
	if name == "" {
		return ingredient{}, fmt.Errorf("ingredient without a name in %q", line)
	}
	fields := strings.Split(list, ", ")
	if len(fields) != properties {
		return ingredient{}, fmt.Errorf("%q does not list %d properties", line, properties)
	}

	var ing ingredient
	for i, field := range fields {
		_, valueText, _ := strings.Cut(field, " ")
		value, err := strconv.Atoi(valueText)
		if err != nil {
			return ingredient{}, fmt.Errorf("invalid property in %q: %w", line, err)
		}
		if value < -maxProperty || value > maxProperty {
			return ingredient{}, fmt.Errorf("property in %q is outside ±%d", line, maxProperty)
		}
		ing[i] = value
	}
	return ing, nil
}

func parseInput(path string) ([]ingredient, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	ingredients := make([]ingredient, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		ing, err := parseIngredient(line)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ing)
	}
	if len(ingredients) == 0 || len(ingredients) > maxIngredients {
		return nil, fmt.Errorf("%s: %d ingredients, expected 1..%d", path, len(ingredients), maxIngredients)
	}
	return ingredients, nil
}

// score returns a recipe's score and its calories.
func score(ingredients []ingredient, amounts [maxIngredients]int) (int, int, error) {
	if len(ingredients) > maxIngredients {
		return 0, 0, fmt.Errorf("%d ingredients, more than %d", len(ingredients), maxIngredients)
	}
	var totals [properties]int
	used := 0
	for i, ing := range ingredients {
		if amounts[i] < 0 {
			return 0, 0, fmt.Errorf("negative amount %d", amounts[i])
		}
		used += amounts[i]
		for p := range properties {
			totals[p] += amounts[i] * ing[p]
		}
	}
	if used != teaspoons {
		return 0, 0, fmt.Errorf("recipe uses %d teaspoons, not %d", used, teaspoons)
	}

	product := 1
	for p := range calories {
		product *= max(totals[p], 0)
	}
	return product, totals[calories], nil
}

// bestRecipes tries every split of the teaspoons and returns the best score
// overall and the best with exactly targetCalories.
func bestRecipes(ingredients []ingredient) (best, bestDiet int, err error) {
	n := len(ingredients)
	if n < 1 || n > maxIngredients {
		return 0, 0, fmt.Errorf("%d ingredients is outside 1..%d", n, maxIngredients)
	}

	recipes := 0
	combinations := 1
	for range n - 1 {
		combinations *= teaspoons + 1
	}
	for code := range combinations {
		var amounts [maxIngredients]int
		left := teaspoons
		for i := range n - 1 {
			amounts[i] = code % (teaspoons + 1)
			code /= teaspoons + 1
			left -= amounts[i]
		}
		if left < 0 {
			continue // the first amounts already use more than 100
		}
		amounts[n-1] = left
		recipes++

		s, cal, err := score(ingredients, amounts)
		if err != nil {
			return 0, 0, err
		}
		best = max(best, s)
		if cal == targetCalories {
			bestDiet = max(bestDiet, s)
		}
	}

	if recipes == 0 {
		return 0, 0, errors.New("no recipe adds up to the teaspoons")
	}
	if bestDiet > best {
		return 0, 0, fmt.Errorf("diet score %d beats the best score %d", bestDiet, best)
	}
	return best, bestDiet, nil
}

func day15(path string) error {
	ingredients, err := parseInput(path)
	if err != nil {
		return err
	}
	best, bestDiet, err := bestRecipes(ingredients)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", best, bestDiet); err != nil {
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
