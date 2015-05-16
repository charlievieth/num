package num

import "strconv"

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
	parseEnd
)

type ScannerError struct {
	msg   string
	bytes int64
}

func (s ScannerError) Error() string {
	return s.msg
}

func (s ScannerError) Bytes() int64 {
	return s.bytes
}

type scanner struct {
	step       func(*scanner, int) int
	parseState int
	bytes      int64
	err        error
}

func (s *scanner) reset() {
	s.step = stateBeginValue
	s.parseState = 0
	s.bytes = 0
	s.err = nil
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
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
			s.parseState = parseEnd
			return scanEndNum
		}
		s.parseState = parseValue
		s.step = stateInValue
		return scanNotNum
	case parseValue:
		s.step = stateBeginValue
		s.parseState = parseEnd
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

func stateError(s *scanner, c int) int {
	return scanError
}

func (s *scanner) error(c int, context string) int {
	s.step = stateError
	s.err = &ScannerError{context, s.bytes}
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
