package main

import (
	"bytes"
	"fmt"
	"strconv"
)

var (
	_ = fmt.Sprint("")
)

var Test = []byte("a 12345 12a 1 1")

type Pos struct {
	x, y int
}

type Num struct {
	i, j int
	b    []byte
	ok   bool
}

func Expand(b []byte) []byte {
	n := bytes.IndexByte(b, '.')
	if n == -1 {
		n = len(b)
	}
	if n <= 3 {
		return b
	}
	c := 3 - (n % 3)
	if c == 3 {
		c = 0
	}
	l := (n / 3) + len(b)
	buf := make([]byte, 0, l)
	for j := 0; j < n; j++ {
		if c == 3 {
			buf = append(buf, ',')
			c = 0
		}
		buf = append(buf, b[j])
		c++
	}
	return append(buf, b[n:]...)
}

func main() {
	s := &scanner{}
	s.reset()
	var pos []Pos
	var n int
	for i, c := range Test {
		switch s.step(s, int(c)) {
		case scanBeginNum:
			n = i
		case scanEndNum:
			if n == -1 {
				panic("error")
			}
			if i-n > 3 {
				pos = append(pos, Pos{n, i})
			}
		case scanNotNum:
			n = -1
		}
	}
	// Check if we ended in a num
	if s.parseState == parseNum && n-len(Test) > 3 {
		pos = append(pos, Pos{n, len(Test)})
	}
	for _, p := range pos {
		fmt.Println(string(Test[p.x:p.y]))
	}
}

const (
	scanContinue = iota
	scanBeginValue
	scanEndValue
	scanBeginNum
	scanEndNum
	scanNotNum
	scanSkipSpace
	scanEnd
	scanError
)

const (
	parseValue = iota
	parseNum
)

type scanner struct {
	step       func(*scanner, int) int
	state      int
	err        error
	parseState int
}

func (s *scanner) reset() {
	s.step = stateBeginValue
	s.parseState = 0
	s.err = nil
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func isNumEnd(c rune) bool {
	return c == ',' || c == ';' || c == '.'
}

func stateBeginValue(s *scanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	if '1' <= c && c <= '9' { // beginning of 1234.5
		s.step = state1
		s.parseState = parseNum
		return scanBeginNum
	}
	switch c {
	case '0': // beginning of 0.123
		s.step = state0
		s.parseState = parseNum
		return scanBeginNum
	case '-':
		s.step = stateNeg
		s.parseState = parseNum
		return scanBeginValue
	default:
		s.step = stateInValue
		s.parseState = parseValue
		return scanBeginValue
	}
}

func stateEndValue(s *scanner, c int) int {
	switch s.parseState {
	case parseNum:
		if isSpace(rune(c)) {
			s.step = stateBeginValue
			return scanEndNum
		}
		s.parseState = parseValue
		s.step = stateInValue
		return scanNotNum
	case parseValue:
		s.step = stateBeginValue
		return scanEndValue
	}
	return s.error(c, "wtf")
}

func stateInValue(s *scanner, c int) int {
	if !isSpace(rune(c)) {
		s.step = stateInValue
		return scanContinue
	}
	return stateEndValue(s, c)
}

func stateNeg(s *scanner, c int) int {
	if c == '0' {
		s.step = state0
		return scanBeginNum
	}
	if '1' <= c && c <= '9' {
		s.step = state1
		return scanBeginNum
	}
	return stateEndValue(s, c)
}

func state1(s *scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = state1
		return scanContinue
	}
	return state0(s, c)
}

func state0(s *scanner, c int) int {
	if c == '.' {
		s.step = stateDot
		return scanContinue
	}
	return stateEndValue(s, c)
}

func stateDot(s *scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanContinue
	}
	return stateEndValue(s, c)
}

func stateDot0(s *scanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanContinue
	}
	return stateEndValue(s, c)
}

func (s *scanner) error(c int, context string) int {
	// s.step = stateError
	// s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}
	return scanError
}

func quoteChar(c int) string {
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}
