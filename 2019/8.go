// Day 8: Space Image Format
//
// The input is an image 25 pixels wide and 6 tall, given as digits in
// layers of 150 one after the other.
//
// Part 1 takes the layer with the fewest 0s and multiplies its count of 1s
// by its count of 2s.
// Part 2 stacks the layers (0 black, 1 white, 2 transparent, the first layer
// in front) and reads the message the image spells. The letters are 4 pixels
// wide with 1 of spacing; each is matched against the known glyphs of the
// Advent of Code font, and one that matches none is an error.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 20
	width         = 25
	height        = 6
	layerSize     = width * height
)

// readImage returns the digits, which must fill whole layers.
func readImage(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	image := strings.TrimSpace(string(content))
	if len(image) == 0 || len(image)%layerSize != 0 || strings.Trim(image, "012") != "" {
		return "", fmt.Errorf("%s: image is not whole layers of 0, 1 and 2", path)
	}
	return image, nil
}

// checksum finds the layer with fewest 0s and returns ones * twos on it.
func checksum(image string) (int, error) {
	if len(image) == 0 || len(image)%layerSize != 0 {
		return 0, errors.New("image is not whole layers")
	}
	best, result := layerSize+1, 0
	for start := 0; start < len(image); start += layerSize {
		var counts [3]int
		for i := start; i < start+layerSize; i++ {
			if image[i] < '0' || image[i] > '2' {
				return 0, fmt.Errorf("pixel %q is not 0, 1 or 2", image[i])
			}
			counts[image[i]-'0']++
		}
		if counts[0] < best {
			best, result = counts[0], counts[1]*counts[2]
		}
	}
	if best > layerSize {
		return 0, errors.New("no layer was counted")
	}
	return result, nil
}

// decode stacks the layers and reads the letters into out.
func decode(image string, out []byte) ([]byte, error) {
	if len(image) == 0 || len(image)%layerSize != 0 {
		return nil, errors.New("image is not whole layers")
	}
	var top [layerSize]byte
	for k := range layerSize {
		top[k] = '2'
		for start := 0; start < len(image) && top[k] == '2'; start += layerSize {
			top[k] = image[start+k]
		}
		if top[k] == '2' {
			return nil, fmt.Errorf("pixel %d is transparent in every layer", k)
		}
	}
	// The font: each letter followed by its six rows of four pixels.
	font := [17]string{
		"A.##.#..##..######..##..#", "B###.#..####.#..##..####.", "C.##.#..##...#...#..#.##.",
		"E#####...###.#...#...####", "F#####...###.#...#...#...", "G.##.#..##...#.###..#.###",
		"H#..##..######..##..##..#", "I.###..#...#...#...#..###", "J..##...#...#...##..#.##.",
		"K#..##.#.##..#.#.#.#.#..#", "L#...#...#...#...#...####", "O.##.#..##..##..##..#.##.",
		"P###.#..##..####.#...#...", "R###.#..##..####.#.#.#..#", "S.####...#....##....####.",
		"U#..##..##..##..##..#.##.", "Z####...#..#..#..#...####",
	}
	out = out[:0]
	var cell [24]byte
	for l := range width / 5 {
		for r := range height {
			for c := range 4 {
				cell[r*4+c] = '.'
				if top[r*width+l*5+c] == '1' {
					cell[r*4+c] = '#'
				}
			}
		}
		letter := byte(0)
		for _, g := range font {
			if g[1:] == string(cell[:]) {
				letter = g[0]
			}
		}
		if letter == 0 {
			return nil, fmt.Errorf("letter %d does not match the font: %s", l, cell[:])
		}
		out = append(out, letter)
	}
	if len(out) != width/5 {
		return nil, fmt.Errorf("read %d letters, expected %d", len(out), width/5)
	}
	return out, nil
}

func day8(path string) error {
	image, err := readImage(path)
	if err != nil {
		return err
	}
	sum, err := checksum(image)
	if err != nil {
		return err
	}

	message, err := decode(image, make([]byte, 0, width/5))
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %s\n", sum, message); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
