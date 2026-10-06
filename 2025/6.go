// Day 6: Trash Compactor
//
// The input is one wide block of text holding several sums printed side by
// side, each separated from the next by a column of blanks. One row of each
// sum carries its operator:
//
//	 123    45
//	+ 45   * 6
//	  78     7
//
// Part 1 reads each row of a sum as one operand, left to right, and folds them
// with the operator. Part 2 reads each column as one operand instead, top to
// bottom, and takes the columns right to left. Both parts sum their results
// over all the sums on the display.
//
// The lines are padded with trailing spaces first so the display is a perfect
// rectangle, which lets any column be indexed without a length check. The sums
// are then found as the maximal runs of columns that are not entirely blank.
// Operands are folded in as they are read rather than collected first, so
// solving a sum allocates nothing.
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
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 20
	maxLines      = 100
	maxDigits     = 18 // the most digits an int64 operand can safely have
)

// A block is a maximal run of columns that are not entirely blank: one sum.
type block struct {
	startCol int // inclusive
	endCol   int // exclusive
}

// readDisplay reads the input and pads every line with trailing spaces so the
// display is a perfect rectangle and any column can be indexed safely.
func readDisplay(path string) (lines []string, width int, err error) {
	if path == "" {
		return nil, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}

	lines = strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) < 2 || len(lines) > maxLines {
		return nil, 0, fmt.Errorf("%s: %d lines, expected 2..%d", path, len(lines), maxLines)
	}
	for _, line := range lines {
		width = max(width, len(line))
	}
	if width == 0 {
		return nil, 0, fmt.Errorf("%s: display is blank", path)
	}
	for i, line := range lines {
		lines[i] = line + strings.Repeat(" ", width-len(line))
	}
	return lines, width, nil
}

func isBlankColumn(lines []string, col int) (bool, error) {
	if len(lines) == 0 {
		return false, errors.New("display is empty")
	}
	for _, line := range lines {
		if col < 0 || col >= len(line) {
			return false, fmt.Errorf("column %d is off the display", col)
		}
		if line[col] != ' ' {
			return false, nil
		}
	}
	return true, nil
}

// findBlocks splits the display into its individual sums.
func findBlocks(lines []string, width int) ([]block, error) {
	if len(lines) == 0 || len(lines[0]) != width {
		return nil, fmt.Errorf("display is not %d columns wide", width)
	}

	// A block takes at least one column plus a blank one, so this never grows.
	blocks := make([]block, 0, width/2+1)
	start := -1
	for col := range width + 1 {
		blank := true // one column past the edge ends the last block
		if col < width {
			var err error
			if blank, err = isBlankColumn(lines, col); err != nil {
				return nil, err
			}
		}
		switch {
		case blank && start >= 0:
			blocks = append(blocks, block{startCol: start, endCol: col})
			start = -1
		case !blank && start < 0:
			start = col
		}
	}

	if len(blocks) == 0 {
		return nil, errors.New("display holds no sums")
	}
	return blocks, nil
}

// findOperator returns the operator of a block and the row it sits on, so that
// row can be left out when collecting the operands.
func findOperator(lines []string, b block) (operator byte, row int, err error) {
	if b.startCol < 0 || b.startCol >= b.endCol {
		return 0, 0, fmt.Errorf("block %v is empty", b)
	}
	for r, line := range lines {
		chunk := line[b.startCol:b.endCol]
		if strings.Contains(chunk, "+") {
			return '+', r, nil
		}
		if strings.Contains(chunk, "*") {
			return '*', r, nil
		}
	}
	return 0, 0, fmt.Errorf("block %v has no operator", b)
}

// combine folds one more operand into a running result.
func combine(result, operand int64, operator byte) (int64, error) {
	if operator != '+' && operator != '*' {
		return 0, fmt.Errorf("unknown operator %q", operator)
	}
	if result < 0 || operand < 0 {
		return 0, fmt.Errorf("negative value in %d %c %d", result, operator, operand)
	}
	if operator == '+' {
		if result > math.MaxInt64-operand {
			return 0, fmt.Errorf("%d + %d overflows int64", result, operand)
		}
		return result + operand, nil
	}
	if operand != 0 && result > math.MaxInt64/operand {
		return 0, fmt.Errorf("%d * %d overflows int64", result, operand)
	}
	return result * operand, nil
}

// foldRows evaluates a block the part 1 way: one operand per row, read left to
// right.
func foldRows(lines []string, b block, operator byte, operatorRow int) (int64, error) {
	if operatorRow < 0 || operatorRow >= len(lines) {
		return 0, fmt.Errorf("operator row %d is off the display", operatorRow)
	}
	if b.startCol < 0 || b.startCol >= b.endCol || b.endCol > len(lines[operatorRow]) {
		return 0, fmt.Errorf("block %v is outside the display", b)
	}

	var result int64
	operands := 0
	for r, line := range lines {
		text := strings.TrimSpace(line[b.startCol:b.endCol])
		if r == operatorRow || text == "" {
			continue
		}
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid operand %q: %w", text, err)
		}
		if operands == 0 {
			result = value
		} else if result, err = combine(result, value, operator); err != nil {
			return 0, err
		}
		operands++
	}

	if operands == 0 {
		return 0, fmt.Errorf("block %v has no row operands", b)
	}
	return result, nil
}

// foldColumns evaluates a block the part 2 way: each column is one operand,
// read top to bottom, and the columns are taken right to left.
func foldColumns(lines []string, b block, operator byte, operatorRow int) (int64, error) {
	if operatorRow < 0 || operatorRow >= len(lines) {
		return 0, fmt.Errorf("operator row %d is off the display", operatorRow)
	}
	if b.startCol < 0 || b.startCol >= b.endCol || b.endCol > len(lines[operatorRow]) {
		return 0, fmt.Errorf("block %v is outside the display", b)
	}

	var result int64
	operands := 0
	for c := b.endCol - 1; c >= b.startCol; c-- {
		var value int64
		digits := 0
		for r, line := range lines {
			if ch := line[c]; r != operatorRow && ch >= '0' && ch <= '9' {
				value = value*10 + int64(ch-'0')
				digits++
			}
		}
		if digits > maxDigits {
			return 0, fmt.Errorf("column %d has more than %d digits", c, maxDigits)
		}
		if digits == 0 {
			continue
		}

		var err error
		if operands == 0 {
			result = value
		} else if result, err = combine(result, value, operator); err != nil {
			return 0, err
		}
		operands++
	}

	if operands == 0 {
		return 0, fmt.Errorf("block %v has no column operands", b)
	}
	return result, nil
}

// solveBlock evaluates one sum both ways: by rows for part 1, by columns for
// part 2.
func solveBlock(lines []string, b block) (byRow, byColumn int64, err error) {
	operator, operatorRow, err := findOperator(lines, b)
	if err != nil {
		return 0, 0, err
	}
	if byRow, err = foldRows(lines, b, operator, operatorRow); err != nil {
		return 0, 0, err
	}
	if byColumn, err = foldColumns(lines, b, operator, operatorRow); err != nil {
		return 0, 0, err
	}
	return byRow, byColumn, nil
}

func day6(path string) error {
	lines, width, err := readDisplay(path)
	if err != nil {
		return err
	}
	blocks, err := findBlocks(lines, width)
	if err != nil {
		return err
	}

	var totalPart1, totalPart2 int64
	for _, b := range blocks {
		byRow, byColumn, err := solveBlock(lines, b)
		if err != nil {
			return err
		}
		if totalPart1, err = combine(totalPart1, byRow, '+'); err != nil {
			return err
		}
		if totalPart2, err = combine(totalPart2, byColumn, '+'); err != nil {
			return err
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", totalPart1, totalPart2); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day6(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
