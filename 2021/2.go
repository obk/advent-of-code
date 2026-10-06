// Day 2: Dive!
//
// Each line of the input is a command: "forward X", "down X" or "up X".
//
// Part 1 reads down/up as changing depth, and multiplies the final
// horizontal position by the final depth.
// Part 2 reads down/up as changing the aim; forward X also moves the depth
// by aim * X. The answer is the same product.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 16
	maxCommands   = 10_000
	maxAmount     = 1_000
)

func day2(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxCommands {
		return fmt.Errorf("%s: more than %d commands", path, maxCommands)
	}
	x, depth, aim, aimDepth := 0, 0, 0, 0
	for _, line := range lines {
		cmd, arg, found := strings.Cut(line, " ")
		n, err := strconv.Atoi(arg)
		if !found || err != nil || n < 0 || n > maxAmount {
			return fmt.Errorf("%s: malformed command %q", path, line)
		}
		switch cmd {
		case "forward":
			x, aimDepth = x+n, aimDepth+aim*n
		case "down":
			depth, aim = depth+n, aim+n
		case "up":
			depth, aim = depth-n, aim-n
		default:
			return fmt.Errorf("%s: unknown command %q", path, cmd)
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", x*depth, x*aimDepth); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
