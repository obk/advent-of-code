// Day 22: Crab Combat
//
// The input is two players' decks. Each round both draw their top card;
// the higher card wins and goes, with the other, under the winner's deck.
// A player with no cards loses.
//
// Part 1 is the winner's score (each card times its position counted from
// the bottom, summed).
// Part 2 plays Recursive Combat: a position seen before in the same game
// ends it, won by player 1; and if both players have at least as many cards
// left as the values they drew, the round is decided by a sub-game on
// copies of that many cards.
//
// Sub-games are frames on an explicit stack. Each depth has its own set of
// seen positions, stored as 64-bit hashes of both decks and allocated up
// front; a game that outgrows it is an error.
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
	ring          = 64 // deck capacity, a power of two above the card count
	maxDepth      = 64
	maxSeen       = 1 << 14
	maxRounds     = 100_000_000
)

// deck is a ring buffer of cards.
type deck struct {
	cards   [ring]int
	head, n int
}

// push puts a card at the bottom.
func (d *deck) push(c int) error {
	if d.n == ring {
		return errors.New("deck is full")
	}
	if c < 1 || c >= ring {
		return fmt.Errorf("card %d is out of range", c)
	}
	d.cards[(d.head+d.n)%ring], d.n = c, d.n+1
	return nil
}

// pop takes the top card.
func (d *deck) pop() (int, error) {
	if d.n == 0 {
		return 0, errors.New("deck is empty")
	}
	c := d.cards[d.head]
	d.head, d.n = (d.head+1)%ring, d.n-1
	if c < 1 {
		return 0, fmt.Errorf("drew card %d", c)
	}
	return c, nil
}

// score sums each card times its place from the bottom.
func (d *deck) score() (int, error) {
	if d.n < 1 || d.n > ring {
		return 0, fmt.Errorf("the winner holds %d cards", d.n)
	}
	s := 0
	for i := range d.n {
		s += d.cards[(d.head+i)%ring] * (d.n - i)
	}
	if s < d.n {
		return 0, fmt.Errorf("score %d is too low for %d cards", s, d.n)
	}
	return s, nil
}

// hash is FNV-1a over both decks, with a separator between them.
func hash(a, b *deck) uint64 {
	h := uint64(14695981039346656037)
	for _, d := range [2]*deck{a, b} {
		for i := range d.n {
			h = (h ^ uint64(d.cards[(d.head+i)%ring])) * 1099511628211
		}
		h = (h ^ 0xff) * 1099511628211
	}
	return h
}

// frame is one game in progress, and the cards of a round waiting on a
// sub-game.
type frame struct {
	p       [2]deck
	drawn   [2]int
	waiting bool
	result  int
}

// readDecks parses both players' decks.
func readDecks(path string) ([2]deck, error) {
	var decks [2]deck
	if path == "" {
		return decks, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return decks, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return decks, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	players := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(players) != 2 {
		return decks, fmt.Errorf("%s: %d players, expected 2", path, len(players))
	}
	var seen [ring]bool
	for p, text := range players {
		lines := strings.Split(text, "\n")
		if len(lines) < 2 {
			return decks, fmt.Errorf("%s: player %d has no cards", path, p+1)
		}
		for _, l := range lines[1:] {
			c, err := strconv.Atoi(l)
			if err != nil || c < 1 || c >= ring {
				return decks, fmt.Errorf("%s: invalid card %q", path, l)
			}
			if seen[c] {
				return decks, fmt.Errorf("%s: card %d is dealt twice", path, c)
			}
			seen[c] = true
			if err := decks[p].push(c); err != nil {
				return decks, err
			}
		}
	}
	return decks, nil
}

// award gives the drawn cards to the winner, winner's card first.
func (f *frame) award(winner int) error {
	if winner != 0 && winner != 1 {
		return fmt.Errorf("player %d cannot win", winner)
	}
	if f.drawn[0] == f.drawn[1] {
		return fmt.Errorf("both players drew %d", f.drawn[0])
	}
	if err := f.p[winner].push(f.drawn[winner]); err != nil {
		return err
	}
	return f.p[winner].push(f.drawn[1-winner])
}

// combat plays the simple game and returns the winner's score.
func combat(decks [2]deck) (int, error) {
	total := decks[0].n + decks[1].n
	f := frame{p: decks}
	for range maxRounds {
		for w := range 2 {
			if f.p[1-w].n == 0 && f.p[w].n != total {
				return 0, fmt.Errorf("the winner holds %d of %d cards", f.p[w].n, total)
			}
			if f.p[1-w].n == 0 {
				return f.p[w].score()
			}
		}
		for p := range 2 {
			var err error
			if f.drawn[p], err = f.p[p].pop(); err != nil {
				return 0, err
			}
		}
		winner := 0
		if f.drawn[1] > f.drawn[0] {
			winner = 1
		}
		if err := f.award(winner); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("no winner in %d rounds", maxRounds)
}

// game is the stack of recursive games and each depth's seen positions.
type game struct {
	frames [maxDepth]frame
	seen   [maxDepth]map[uint64]bool
	depth  int
}

// round plays one round of the game on top of the stack. It returns the
// winner of the whole game once the outermost game ends, else -1.
func (g *game) round() (int, error) {
	d := g.depth
	if d < 0 || d >= maxDepth {
		return 0, fmt.Errorf("depth %d is out of range", d)
	}
	f := &g.frames[d]
	winner := -1
	switch {
	case f.waiting:
		f.waiting = false
		return -1, f.award(f.result)
	case f.p[1].n == 0 || g.seen[d][hash(&f.p[0], &f.p[1])]:
		winner = 0
	case f.p[0].n == 0:
		winner = 1
	}
	if winner >= 0 { // this game is over: report to the game below
		if d == 0 {
			return winner, nil
		}
		g.depth--
		g.frames[d-1].result = winner
		return -1, nil
	}
	if len(g.seen[d]) == maxSeen {
		return 0, fmt.Errorf("game at depth %d ran past %d positions", d, maxSeen)
	}
	g.seen[d][hash(&f.p[0], &f.p[1])] = true
	for p := range 2 {
		var err error
		if f.drawn[p], err = f.p[p].pop(); err != nil {
			return 0, err
		}
	}
	if f.p[0].n >= f.drawn[0] && f.p[1].n >= f.drawn[1] {
		return -1, g.subgame()
	}
	if f.drawn[1] > f.drawn[0] {
		return -1, f.award(1)
	}
	return -1, f.award(0)
}

// subgame starts a game on copies of as many cards as each player drew.
func (g *game) subgame() error {
	if g.depth+1 == maxDepth {
		return fmt.Errorf("sub-games nest deeper than %d", maxDepth)
	}
	f := &g.frames[g.depth]
	if f.drawn[0] > f.p[0].n || f.drawn[1] > f.p[1].n {
		return errors.New("not enough cards for a sub-game")
	}
	f.waiting = true
	sub := &g.frames[g.depth+1]
	*sub = frame{}
	for p := range 2 {
		for i := range f.drawn[p] {
			if err := sub.p[p].push(f.p[p].cards[(f.p[p].head+i)%ring]); err != nil {
				return err
			}
		}
	}
	g.depth++
	clear(g.seen[g.depth])
	return nil
}

// recursive plays Recursive Combat and returns the winner's score.
func recursive(decks [2]deck) (int, error) {
	if decks[0].n == 0 || decks[1].n == 0 {
		return 0, errors.New("a player starts with no cards")
	}
	g := new(game)
	for d := range g.seen {
		g.seen[d] = make(map[uint64]bool, maxSeen)
	}
	g.frames[0] = frame{p: decks}
	for range maxRounds {
		winner, err := g.round()
		if err != nil {
			return 0, err
		}
		if winner >= 0 {
			return g.frames[0].p[winner].score()
		}
	}
	return 0, fmt.Errorf("no winner in %d rounds", maxRounds)
}

func day22(path string) error {
	decks, err := readDecks(path)
	if err != nil {
		return err
	}
	simple, err := combat(decks)
	if err != nil {
		return err
	}
	deep, err := recursive(decks)
	if err != nil {
		return err
	}

	if decks[0].n+decks[1].n == 0 {
		return errors.New("no cards were dealt")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", simple, deep); err != nil {
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
