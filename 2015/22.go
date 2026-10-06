// Day 22: Wizard Simulator 20XX
//
// The input is the boss's hit points and damage. You have 50 hit points and
// 500 mana, and each turn cast one of five spells - Magic Missile, Drain,
// Shield, Poison or Recharge - while the boss hits back. Effects tick at the
// start of both players' turns, and an effect cannot be recast while active.
//
// Part 1 is the least mana spent on a fight you win.
// Part 2 is the same on "hard", where you lose 1 hit point at the start of
// each of your turns.
//
// The fights form a tree, one branch per spell, searched depth first with an
// explicit stack. A branch is abandoned as soon as its mana spent reaches the
// best win found so far, which keeps the search small.
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
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 10
	maxStat       = 1_000
	maxStack      = 4_096
	maxSteps      = 10_000_000
)

const (
	playerHP   = 50
	playerMana = 500
	spells     = 5
)

// state is one moment of the fight, at the start of the player's turn.
type state struct {
	hp, mana, bossHP         int
	shield, poison, recharge int // turns left on each effect
	spent                    int
}

// spell is what one cast costs and does.
type spell struct {
	cost, damage, heal, shield, poison, recharge int
}

// readBoss returns the boss's hit points and damage.
func readBoss(path string) (hp, damage int, err error) {
	if path == "" {
		return 0, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}

	var stats [2]int
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(stats) {
		return 0, 0, fmt.Errorf("%s: %d lines, expected %d", path, len(lines), len(stats))
	}
	for i, line := range lines {
		_, valueText, found := strings.Cut(line, ": ")
		if !found {
			return 0, 0, fmt.Errorf("%s: malformed line %q", path, line)
		}
		if stats[i], err = strconv.Atoi(strings.TrimSpace(valueText)); err != nil {
			return 0, 0, fmt.Errorf("%s: invalid stat in %q: %w", path, line, err)
		}
		if stats[i] < 1 || stats[i] > maxStat {
			return 0, 0, fmt.Errorf("%s: stat in %q is outside 1..%d", path, line, maxStat)
		}
	}
	return stats[0], stats[1], nil
}

// applyEffects ticks every active effect and returns the armor it gives.
func (s *state) applyEffects() (armor int, err error) {
	if s.shield < 0 || s.poison < 0 || s.recharge < 0 {
		return 0, fmt.Errorf("negative effect timer in %+v", *s)
	}
	if s.shield > 0 {
		armor = 7
		s.shield--
	}
	if s.poison > 0 {
		s.bossHP -= 3
		s.poison--
	}
	if s.recharge > 0 {
		s.mana += 101
		s.recharge--
	}
	return armor, nil
}

// playRound casts spell sp from s and plays out the boss's turn. It reports
// whether the move was legal and, if so, whether the boss died.
func playRound(s *state, sp spell, bossDamage int, hard bool) (legal, won bool, err error) {
	if bossDamage < 1 || sp.cost < 1 {
		return false, false, fmt.Errorf("invalid round: boss damage %d, spell cost %d", bossDamage, sp.cost)
	}
	if hard {
		if s.hp--; s.hp <= 0 {
			return false, false, nil
		}
	}
	if _, err := s.applyEffects(); err != nil {
		return false, false, err
	}
	if s.bossHP <= 0 {
		return true, true, nil // poison won before a spell was needed
	}
	busy := (sp.shield > 0 && s.shield > 0) || (sp.poison > 0 && s.poison > 0) || (sp.recharge > 0 && s.recharge > 0)
	if sp.cost > s.mana || busy {
		return false, false, nil
	}
	s.mana, s.spent = s.mana-sp.cost, s.spent+sp.cost
	s.bossHP, s.hp = s.bossHP-sp.damage, s.hp+sp.heal
	s.shield, s.poison, s.recharge = s.shield+sp.shield, s.poison+sp.poison, s.recharge+sp.recharge
	if s.bossHP <= 0 {
		return true, true, nil
	}

	armor, err := s.applyEffects()
	if err != nil {
		return false, false, err
	}
	if s.bossHP <= 0 {
		return true, true, nil
	}
	s.hp -= max(bossDamage-armor, 1)
	return s.hp > 0, false, nil
}

// leastMana searches every fight for the cheapest win. stack is scratch space.
func leastMana(bossHP, bossDamage int, hard bool, stack []state) (int, error) {
	if len(stack) < spells+1 {
		return 0, errors.New("search stack is too small")
	}
	if bossHP < 1 || bossDamage < 1 {
		return 0, fmt.Errorf("invalid boss: %d hit points, %d damage", bossHP, bossDamage)
	}

	book := [spells]spell{
		{cost: 53, damage: 4},          // Magic Missile
		{cost: 73, damage: 2, heal: 2}, // Drain
		{cost: 113, shield: 6},         // Shield
		{cost: 173, poison: 6},         // Poison
		{cost: 229, recharge: 5},       // Recharge
	}
	best, top := -1, 1
	stack[0] = state{hp: playerHP, mana: playerMana, bossHP: bossHP}
	for range maxSteps {
		if top == 0 {
			if best < 0 {
				return 0, errors.New("no way to win")
			}
			return best, nil
		}
		top--
		current := stack[top]
		for _, sp := range book {
			next := current
			legal, won, err := playRound(&next, sp, bossDamage, hard)
			if err != nil {
				return 0, err
			}
			switch {
			case !legal || (best >= 0 && next.spent >= best):
			case won:
				best = next.spent
			case top >= len(stack):
				return 0, fmt.Errorf("search stack of %d overflowed", len(stack))
			default:
				stack[top] = next
				top++
			}
		}
	}
	return 0, fmt.Errorf("search exceeded %d steps", maxSteps)
}

func day22(path string) error {
	bossHP, bossDamage, err := readBoss(path)
	if err != nil {
		return err
	}

	stack := make([]state, maxStack)
	normal, err := leastMana(bossHP, bossDamage, false, stack)
	if err != nil {
		return err
	}
	hard, err := leastMana(bossHP, bossDamage, true, stack)
	if err != nil {
		return err
	}

	if hard < normal {
		return fmt.Errorf("hard mode cost %d, less than normal %d", hard, normal)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", normal, hard); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
