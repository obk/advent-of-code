// Day 4: Passport Processing
//
// The input is passports separated by blank lines, each a set of key:value
// fields. byr, iyr, eyr, hgt, hcl, ecl and pid are required; cid is not.
//
// Part 1 counts passports with every required field.
// Part 2 also checks the values: byr 1920-2002, iyr 2010-2020, eyr
// 2020-2030, hgt 150-193cm or 59-76in, hcl '#' and six hex digits, ecl one
// of amb blu brn gry grn hzl oth, pid nine digits.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 16
	maxPassports  = 1_000
)

// The required fields, in a fixed order.
const required = "byr iyr eyr hgt hcl ecl pid"

// inRange reports whether text is a number from lo to hi with the given
// number of digits (0 for any).
func inRange(text string, lo, hi, digits int) bool {
	n, err := strconv.Atoi(text)
	return err == nil && n >= lo && n <= hi && (digits == 0 || len(text) == digits)
}

// fieldValid checks one field's value.
func fieldValid(key, value string) (bool, error) {
	if len(key) != 3 {
		return false, fmt.Errorf("field name %q is not three letters", key)
	}
	if value == "" {
		return false, nil
	}
	switch key {
	case "byr":
		return inRange(value, 1920, 2002, 4), nil
	case "iyr":
		return inRange(value, 2010, 2020, 4), nil
	case "eyr":
		return inRange(value, 2020, 2030, 4), nil
	case "hgt":
		if len(value) < 3 {
			return false, nil
		}
		n, unit := value[:len(value)-2], value[len(value)-2:]
		return (unit == "cm" && inRange(n, 150, 193, 0)) || (unit == "in" && inRange(n, 59, 76, 0)), nil
	case "hcl":
		return len(value) == 7 && value[0] == '#' && strings.Trim(value[1:], "0123456789abcdef") == "", nil
	case "ecl":
		return len(value) == 3 && strings.Contains(" amb blu brn gry grn hzl oth ", " "+value+" "), nil
	case "pid":
		return len(value) == 9 && strings.Trim(value, "0123456789") == "", nil
	case "cid":
		return true, nil
	}
	return false, fmt.Errorf("unknown field %q", key)
}

// check returns whether a passport has every required field, and whether
// they are all valid.
func check(passport string) (present, valid bool, err error) {
	if passport == "" {
		return false, false, errors.New("empty passport")
	}
	var have, good [7]bool
	names := strings.Fields(required)
	fields := strings.Fields(passport)
	if len(fields) > len(names)+1 {
		return false, false, fmt.Errorf("passport has %d fields", len(fields))
	}
	for _, field := range fields {
		key, value, found := strings.Cut(field, ":")
		if !found {
			return false, false, fmt.Errorf("malformed field %q", field)
		}
		ok, err := fieldValid(key, value)
		if err != nil {
			return false, false, err
		}
		for i, name := range names {
			if name == key && have[i] {
				return false, false, fmt.Errorf("field %s appears twice", key)
			}
			if name == key {
				have[i], good[i] = true, ok
			}
		}
	}
	return have == [7]bool{true, true, true, true, true, true, true}, good == [7]bool{true, true, true, true, true, true, true}, nil
}

func day4(path string) error {
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
	passports := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(passports) == 0 || len(passports) > maxPassports {
		return fmt.Errorf("%s: more than %d passports", path, maxPassports)
	}
	present, valid := 0, 0
	for _, p := range passports {
		a, b, err := check(p)
		if err != nil {
			return err
		}
		if a {
			present++
		}
		if b {
			valid++
		}
	}

	if present > len(passports) {
		return fmt.Errorf("%d complete passports out of %d", present, len(passports))
	}
	if valid > present {
		return fmt.Errorf("%d valid passports but only %d complete", valid, present)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", present, valid); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
