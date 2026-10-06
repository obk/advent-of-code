// Day 15: Beacon Exclusion Zone
//
// Each line of the input is a sensor and the beacon closest to it
// (Manhattan distance). No other beacon can be closer to that sensor.
//
// Part 1 counts the positions on row 2000000 where no beacon can be.
// Part 2 finds the one position with x and y in 0..4000000 that no sensor
// covers, and gives x*4000000 + y.
//
// Part 1 merges each sensor's interval on the row (sorted by insertion).
// For part 2, the gap lies just outside some sensors' ranges, on the
// diagonal lines one step beyond their edges; it is where two such lines
// cross, so only those crossings are checked.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 12
	maxSensors    = 64
	row           = 2_000_000
	limit         = 4_000_000
)

// sensor is a sensor, its beacon and its range.
type sensor struct {
	x, y, bx, by, r int
}

// readSensors parses the report.
func readSensors(path string) ([]sensor, error) {
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
	if len(lines) == 0 || len(lines) > maxSensors {
		return nil, fmt.Errorf("%s: more than %d sensors", path, maxSensors)
	}
	sensors := make([]sensor, len(lines))
	for i, l := range lines {
		s := &sensors[i]
		n, err := fmt.Sscanf(l, "Sensor at x=%d, y=%d: closest beacon is at x=%d, y=%d", &s.x, &s.y, &s.bx, &s.by)
		if err != nil || n != 4 {
			return nil, fmt.Errorf("%s: malformed line %q", path, l)
		}
		s.r = max(s.x-s.bx, s.bx-s.x) + max(s.y-s.by, s.by-s.y)
		if s.r == 0 {
			return nil, fmt.Errorf("%s: sensor %q sits on its beacon", path, l)
		}
	}
	return sensors, nil
}

// covered counts positions on row y that no beacon can occupy.
func covered(sensors []sensor, y int, spans [][2]int) (int, error) {
	if cap(spans) < len(sensors) {
		return 0, errors.New("no room for the spans")
	}
	if len(sensors) == 0 {
		return 0, errors.New("no sensors")
	}
	spans = spans[:0]
	for _, s := range sensors {
		w := s.r - max(s.y-y, y-s.y)
		if w < 0 {
			continue
		}
		span := [2]int{s.x - w, s.x + w}
		spans = append(spans, span)
		for k := len(spans) - 1; k > 0 && spans[k-1][0] > spans[k][0]; k-- {
			spans[k-1], spans[k] = spans[k], spans[k-1]
		}
	}
	total, end := 0, -1<<62
	for _, sp := range spans {
		if sp[0] > sp[1] {
			return 0, fmt.Errorf("span %v is backwards", sp)
		}
		if sp[1] > end {
			total += sp[1] - max(sp[0], end+1) + 1
			end = sp[1]
		}
	}
	// Positions holding a known beacon do not count.
	var seen [maxSensors]int
	n := 0
	for _, s := range sensors {
		dup := false
		for _, x := range seen[:n] {
			dup = dup || x == s.bx
		}
		if s.by == y && !dup {
			seen[n], n, total = s.bx, n+1, total-1
		}
	}
	if total < 0 {
		return 0, fmt.Errorf("negative coverage %d", total)
	}
	return total, nil
}

// free reports whether no sensor covers (x, y).
func free(sensors []sensor, x, y int) bool {
	for _, s := range sensors {
		if max(s.x-x, x-s.x)+max(s.y-y, y-s.y) <= s.r {
			return false
		}
	}
	return true
}

// gap finds the one uncovered position in the search area.
func gap(sensors []sensor) (int, error) {
	if len(sensors) == 0 {
		return 0, errors.New("no sensors")
	}
	if len(sensors) > maxSensors {
		return 0, fmt.Errorf("more than %d sensors", maxSensors)
	}
	for _, s := range sensors {
		for _, t := range sensors {
			// Lines y - x = a beyond s, and y + x = b beyond t.
			for _, a := range [2]int{s.y - s.x + s.r + 1, s.y - s.x - s.r - 1} {
				for _, b := range [2]int{t.y + t.x + t.r + 1, t.y + t.x - t.r - 1} {
					if (b-a)%2 != 0 {
						continue
					}
					x, y := (b-a)/2, (a+b)/2
					if x >= 0 && y >= 0 && x <= limit && y <= limit && free(sensors, x, y) {
						return x*limit + y, nil
					}
				}
			}
		}
	}
	return 0, errors.New("no uncovered position")
}

func day15(path string) error {
	sensors, err := readSensors(path)
	if err != nil {
		return err
	}
	count, err := covered(sensors, row, make([][2]int, 0, len(sensors)))
	if err != nil {
		return err
	}
	freq, err := gap(sensors)
	if err != nil {
		return err
	}

	if count < 1 || freq < 1 {
		return fmt.Errorf("coverage %d or frequency %d is not positive", count, freq)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, freq); err != nil {
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
