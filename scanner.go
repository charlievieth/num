/*
	The following part of the code may contain portions of the Go
	standard library (encoding/json/scanner.go), which tells me to
	retain their copyright notice.

	Copyright (c) 2012 The Go Authors. All rights reserved.

	Redistribution and use in source and binary forms, with or without
	modification, are permitted provided that the following conditions are
	met:

	   * Redistributions of source code must retain the above copyright
	notice, this list of conditions and the following disclaimer.
	   * Redistributions in binary form must reproduce the above
	copyright notice, this list of conditions and the following disclaimer
	in the documentation and/or other materials provided with the
	distribution.
	   * Neither the name of Google Inc. nor the names of its
	contributors may be used to endorse or promote products derived from
	this software without specific prior written permission.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
	"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
	LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
	A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
	OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
	SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
	LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
	DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
	THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
	(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
	OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package num

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

func newScanner() *scanner {
	return &scanner{step: stateBeginValue}
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

func isStart(c rune) bool {
	return c == '(' || c == '[' || c == '{' || c == '"' || c == '\''
}

func isEnd(c rune) bool {
	return c == ')' || c == ']' || c == '}' || c == '"' || c == '\''
}

func stateBeginValue(s *scanner, c int) int {
	if c < ' ' || isSpace(rune(c)) || isStart(rune(c)) {
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
		if isSpace(rune(c)) || isEnd(rune(c)) {
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
	return s.error(c, "invalid parse state")
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
