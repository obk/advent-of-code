// Day 21: RPG Simulator 20XX
//
// The input is the boss's hit points, damage and armor. You have 100 hit
// points and buy equipment from a fixed shop: exactly one weapon, at most one
// armor, and up to two different rings. You and the boss take turns hitting,
// you first, each hit doing the attacker's damage minus the defender's armor
// (at least 1).
//
// Part 1 is the least gold that buys a win.
// Part 2 is the most gold the shopkeeper can make you spend and still lose.
//
// There are only 5 x 6 x 22 loadouts, so every one is tried. A fight needs
// no simulation: each side needs ceil(hp / damage per hit) hits to win, and
// you win when you need no more hits than the boss, since you strike first.
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
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 10
	maxStat       = 10_000
	playerHP      = 100
)

// item is one thing the shop sells; a zero item stands for "none".
type item struct {
	cost, damage, armor int
}

// fighter is a combatant's stats.
type fighter struct {
	hp, damage, armor int
}

// readBoss parses the boss's three stats.
func readBoss(path string) (fighter, error) {
	if path == "" {
		return fighter{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fighter{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return fighter{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	names := [3]string{"Hit Points", "Damage", "Armor"}
	if len(lines) != len(names) {
		return fighter{}, fmt.Errorf("%s: %d lines, expected %d", path, len(lines), len(names))
	}
	var stats [3]int
	for i, line := range lines {
		name, valueText, found := strings.Cut(strings.TrimSpace(line), ": ")
		if !found || name != names[i] {
			return fighter{}, fmt.Errorf("%s: expected %q, got %q", path, names[i], line)
		}
		if stats[i], err = strconv.Atoi(valueText); err != nil {
			return fighter{}, fmt.Errorf("%s: invalid %s: %w", path, name, err)
		}
		if stats[i] < 0 || stats[i] > maxStat {
			return fighter{}, fmt.Errorf("%s: %s %d is outside 0..%d", path, name, stats[i], maxStat)
		}
	}
	return fighter{hp: stats[0], damage: stats[1], armor: stats[2]}, nil
}

// hitsToKill is how many hits the attacker needs to bring the defender to 0.
func hitsToKill(attacker, defender fighter) (int, error) {
	if defender.hp < 1 {
		return 0, fmt.Errorf("defender has %d hit points", defender.hp)
	}
	if attacker.damage < 0 || defender.armor < 0 {
		return 0, errors.New("negative damage or armor")
	}
	perHit := max(attacker.damage-defender.armor, 1)
	hits := (defender.hp + perHit - 1) / perHit
	if hits < 1 || hits > defender.hp {
		return 0, fmt.Errorf("%d hits to take %d hit points", hits, defender.hp)
	}
	return hits, nil
}

// cheapestWinAndDearestLoss tries every loadout.
func cheapestWinAndDearestLoss(boss fighter) (cheapest, dearest int, err error) {
	if boss.hp < 1 {
		return 0, 0, fmt.Errorf("boss has %d hit points", boss.hp)
	}
	// The shop's stock. Armor and rings list "none" first, so a pair of ring
	// indices picks zero to two rings.
	weapons := [5]item{{8, 4, 0}, {10, 5, 0}, {25, 6, 0}, {40, 7, 0}, {74, 8, 0}}
	armors := [6]item{{}, {13, 0, 1}, {31, 0, 2}, {53, 0, 3}, {75, 0, 4}, {102, 0, 5}}
	rings := [7]item{{}, {25, 1, 0}, {50, 2, 0}, {100, 3, 0}, {20, 0, 1}, {40, 0, 2}, {80, 0, 3}}

	cheapest, dearest = -1, -1
	for _, w := range weapons {
		for _, a := range armors {
			for r1 := range rings {
				for r2 := r1; r2 < len(rings); r2++ {
					if r1 == r2 && r1 != 0 {
						continue // only one of each ring
					}
					kit := [4]item{w, a, rings[r1], rings[r2]}
					me := fighter{hp: playerHP}
					cost := 0
					for _, it := range kit {
						cost, me.damage, me.armor = cost+it.cost, me.damage+it.damage, me.armor+it.armor
					}
					myHits, err := hitsToKill(me, boss)
					if err != nil {
						return 0, 0, err
					}
					bossHits, err := hitsToKill(boss, me)
					if err != nil {
						return 0, 0, err
					}
					if win := myHits <= bossHits; win && (cheapest < 0 || cost < cheapest) {
						cheapest = cost
					} else if !win {
						dearest = max(dearest, cost)
					}
				}
			}
		}
	}

	if cheapest < 0 || dearest < 0 {
		return 0, 0, fmt.Errorf("no winning or no losing loadout (%d, %d)", cheapest, dearest)
	}
	return cheapest, dearest, nil
}

func day21(path string) error {
	boss, err := readBoss(path)
	if err != nil {
		return err
	}
	cheapest, dearest, err := cheapestWinAndDearestLoss(boss)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", cheapest, dearest); err != nil {
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
