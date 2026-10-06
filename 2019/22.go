// Day 22: Slam Shuffle
//
// Each line of the input is a shuffle step on a deck of 10007 cards: "deal
// into new stack" (reverse), "cut N" (move the top N to the bottom, or the
// bottom -N to the top) or "deal with increment N" (deal card i to position
// i*N, wrapping round).
//
// Part 1 is the position of card 2019 after the shuffle.
// Part 2 shuffles a deck of 119315717514047 cards 101741582076661 times,
// and asks which card ends up at position 2020.
//
// Every step moves a card at position p to a*p + b (mod the deck size), so
// one card is followed through the steps without building a deck.
//
// For part 2 the steps compose into one map p -> a*p + b, and repeating
// it k times is a power of that map, found by repeated squaring. The deck
// size is prime, so the result can be inverted to ask which card lands at
// 2020. Products of two numbers below the deck size need 128 bits, which
// math/bits gives.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 16
	maxSteps      = 1_000
	deckSize      = 10_007
	bigDeck       = 119_315_717_514_047
	shuffles      = 101_741_582_076_661
)

// linear is the map p -> a*p + b modulo n.
type linear struct {
	a, b, n uint64
}

// step is one shuffle technique as p -> a*p + b.
type step struct {
	a, b int
}

// readSteps parses the shuffle; amounts must be below the deck size.
func readSteps(path string) ([]step, error) {
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
	if len(lines) > maxSteps {
		return nil, fmt.Errorf("%s: more than %d steps", path, maxSteps)
	}
	steps := make([]step, len(lines))
	for i, line := range lines {
		f := strings.Fields(line)
		if line == "deal into new stack" {
			steps[i] = step{-1, -1}
			continue
		}
		if len(f) < 2 {
			return nil, fmt.Errorf("%s: unknown step %q", path, line)
		}
		n, err := strconv.Atoi(f[len(f)-1])
		if err != nil || n <= -deckSize || n >= deckSize {
			return nil, fmt.Errorf("%s: invalid amount in %q", path, line)
		}
		switch {
		case f[0] == "cut":
			steps[i] = step{1, -n}
		case strings.HasPrefix(line, "deal with increment ") && n > 0:
			steps[i] = step{n, 0}
		default:
			return nil, fmt.Errorf("%s: unknown step %q", path, line)
		}
	}
	return steps, nil
}

// follow returns where the card at position p ends up in a deck of n.
func follow(steps []step, p, n int) (int, error) {
	if n < 2 || n > deckSize || p < 0 || p >= n {
		return 0, fmt.Errorf("position %d in a deck of %d", p, n)
	}
	for _, s := range steps {
		p = ((s.a*p+s.b)%n + n) % n
	}
	if p < 0 || p >= n {
		return 0, fmt.Errorf("card ended at %d, outside the deck", p)
	}
	return p, nil
}

// mulmod returns x*y mod n.
func mulmod(x, y, n uint64) (uint64, error) {
	if n < 2 || x >= n || y >= n {
		return 0, fmt.Errorf("%d * %d mod %d is out of range", x, y, n)
	}
	hi, lo := bits.Mul64(x, y)
	return bits.Rem64(hi, lo, n), nil
}

// then returns the map that applies f and then g.
func (f linear) then(g linear) (linear, error) {
	if f.n != g.n {
		return linear{}, fmt.Errorf("maps modulo %d and %d do not compose", f.n, g.n)
	}
	if f.b >= f.n || g.b >= g.n {
		return linear{}, errors.New("map offset is not reduced")
	}
	a, err := mulmod(g.a, f.a, f.n)
	if err != nil {
		return linear{}, err
	}
	b, err := mulmod(g.a, f.b, f.n)
	if err != nil {
		return linear{}, err
	}
	return linear{a, (b + g.b) % f.n, f.n}, nil
}

// compose turns the steps into one map modulo n.
func compose(steps []step, n uint64) (linear, error) {
	if n < deckSize || len(steps) == 0 {
		return linear{}, fmt.Errorf("deck of %d cards or no steps", n)
	}
	f := linear{1, 0, n}
	for _, s := range steps {
		a := uint64((int64(s.a)%int64(n) + int64(n)) % int64(n))
		b := uint64((int64(s.b)%int64(n) + int64(n)) % int64(n))
		var err error
		if f, err = f.then(linear{a, b, n}); err != nil {
			return linear{}, err
		}
	}
	return f, nil
}

// power returns f applied k times.
func (f linear) power(k uint64) (linear, error) {
	if f.n < 2 {
		return linear{}, fmt.Errorf("modulus %d is too small", f.n)
	}
	result := linear{1, 0, f.n}
	var err error
	for bit := 0; bit < 64 && k > 0; bit++ {
		if k&1 == 1 {
			if result, err = result.then(f); err != nil {
				return linear{}, err
			}
		}
		if f, err = f.then(f); err != nil {
			return linear{}, err
		}
		k >>= 1
	}
	if k != 0 {
		return linear{}, errors.New("the power did not finish")
	}
	return result, nil
}

// cardAt returns the card that f moves to position p: the x with
// a*x + b = p, using a's inverse a^(n-2) (n is prime).
func (f linear) cardAt(p uint64) (uint64, error) {
	if p >= f.n || f.a == 0 {
		return 0, fmt.Errorf("position %d, or the map is not invertible", p)
	}
	inv, err := linear{f.a, 0, f.n}.power(f.n - 2)
	if err != nil {
		return 0, err
	}
	x, err := mulmod((p+f.n-f.b)%f.n, inv.a, f.n)
	if err != nil {
		return 0, err
	}
	ax, err := mulmod(f.a, x, f.n)
	if err != nil || (ax+f.b)%f.n != p {
		return 0, fmt.Errorf("card %d does not land at %d (%v)", x, p, err)
	}
	return x, nil
}

func day22(path string) error {
	steps, err := readSteps(path)
	if err != nil {
		return err
	}
	pos, err := follow(steps, 2019, deckSize)
	if err != nil {
		return err
	}

	// The composed map must agree with following the card step by step.
	small, err := compose(steps, deckSize)
	if err != nil {
		return err
	}
	if (small.a*2019+small.b)%deckSize != uint64(pos) {
		return fmt.Errorf("the composed shuffle puts 2019 at %d, not %d", (small.a*2019+small.b)%deckSize, pos)
	}
	big, err := compose(steps, bigDeck)
	if err != nil {
		return err
	}
	if big, err = big.power(shuffles); err != nil {
		return err
	}
	card, err := big.cardAt(2020)
	if err != nil {
		return err
	}
	if card >= bigDeck {
		return fmt.Errorf("card %d is not in the deck", card)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", pos, card); err != nil {
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
