// Day 12: JSAbacusFramework.io
//
// The input is one JSON document. No string in it contains a number.
//
// Part 1 is the sum of every number in the document.
// Part 2 leaves out every object - and everything inside it - that has a
// property whose value is the string "red". Arrays are not affected.
//
// The document is scanned once, character by character, with an explicit
// stack of the arrays and objects currently open. Each entry keeps its own
// sum and whether it has been marked red; closing it adds its sum to the
// enclosing one unless it is a red object. Part 1's total ignores the red
// marks altogether.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 20
	maxDepth      = 1_000
	maxNumberLen  = 15 // digits; keeps every number and sum inside int
)

// container is one array or object on the stack.
type container struct {
	isObject bool
	red      bool
	sum      int // part 2 sum of what it holds so far
}

// scanner walks the document, keeping the stack of open containers.
type scanner struct {
	doc   string
	pos   int
	stack [maxDepth + 1]container // entry 0 is the top level
	depth int
	total int // part 1: every number
}

// readDocument returns the JSON document.
func readDocument(path string) (string, error) {
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
	doc := strings.TrimSpace(string(content))
	if doc[0] != '[' && doc[0] != '{' {
		return "", fmt.Errorf("%s: document is not an array or object", path)
	}
	return doc, nil
}

// readNumber reads an integer starting at s.pos and adds it to both sums.
func (s *scanner) readNumber() error {
	if s.pos >= len(s.doc) {
		return errors.New("number expected past the end of the document")
	}
	start := s.pos
	if s.doc[s.pos] == '-' {
		s.pos++
	}
	digits := 0
	value := 0
	for i := s.pos; i < len(s.doc) && s.doc[i] >= '0' && s.doc[i] <= '9'; i++ {
		value = value*10 + int(s.doc[i]-'0')
		digits++
	}
	s.pos += digits
	if digits == 0 || digits > maxNumberLen {
		return fmt.Errorf("malformed number at %d", start)
	}
	if s.doc[start] == '-' {
		value = -value
	}
	s.total += value
	s.stack[s.depth].sum += value
	return nil
}

// readString skips a string starting at s.pos, marking the enclosing object
// red if this is a property value equal to "red".
func (s *scanner) readString() error {
	if s.pos >= len(s.doc) || s.doc[s.pos] != '"' {
		return fmt.Errorf("string expected at %d", s.pos)
	}
	start := s.pos
	end := strings.IndexByte(s.doc[start+1:], '"')
	if end < 0 {
		return fmt.Errorf("unterminated string at %d", start)
	}
	text := s.doc[start+1 : start+1+end]
	if strings.IndexByte(text, '\\') >= 0 {
		return fmt.Errorf("escaped string at %d is not supported", start)
	}
	s.pos = start + end + 2

	// A property value is the string right after a colon.
	isValue := start > 0 && s.doc[start-1] == ':'
	if isValue && text == "red" && s.stack[s.depth].isObject {
		s.stack[s.depth].red = true
	}
	return nil
}

// open pushes a new array or object.
func (s *scanner) open(isObject bool) error {
	if s.pos >= len(s.doc) {
		return errors.New("bracket expected past the end of the document")
	}
	if s.depth >= maxDepth {
		return fmt.Errorf("nesting deeper than %d at %d", maxDepth, s.pos)
	}
	s.depth++
	s.stack[s.depth] = container{isObject: isObject}
	s.pos++
	return nil
}

// close pops the innermost container, adding its sum to its parent unless it
// is a red object.
func (s *scanner) close(isObject bool) error {
	if s.pos >= len(s.doc) {
		return errors.New("bracket expected past the end of the document")
	}
	if s.depth == 0 {
		return fmt.Errorf("unmatched closing bracket at %d", s.pos)
	}
	inner := s.stack[s.depth]
	if inner.isObject != isObject {
		return fmt.Errorf("mismatched closing bracket at %d", s.pos)
	}
	s.depth--
	if !inner.red {
		s.stack[s.depth].sum += inner.sum
	}
	s.pos++
	return nil
}

// scan walks the whole document. Every step consumes at least one character.
func (s *scanner) scan() error {
	for range len(s.doc) + 1 {
		if s.pos >= len(s.doc) {
			break
		}
		var err error
		switch c := s.doc[s.pos]; {
		case c == '{' || c == '[':
			err = s.open(c == '{')
		case c == '}' || c == ']':
			err = s.close(c == '}')
		case c == '"':
			err = s.readString()
		case c == '-' || (c >= '0' && c <= '9'):
			err = s.readNumber()
		case c == ',' || c == ':' || c == ' ' || c == '\n':
			s.pos++
		default:
			return fmt.Errorf("unexpected %q at %d", c, s.pos)
		}
		if err != nil {
			return err
		}
	}

	if s.pos < len(s.doc) {
		return errors.New("document was not fully read")
	}
	if s.depth != 0 {
		return fmt.Errorf("%d brackets left open", s.depth)
	}
	return nil
}

func day12(path string) error {
	doc, err := readDocument(path)
	if err != nil {
		return err
	}
	s := &scanner{doc: doc}
	if err := s.scan(); err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", s.total, s.stack[0].sum); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
