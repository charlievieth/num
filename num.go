package num

import (
	"bytes"
	"io"
)

type Num struct {
	buf  *bytes.Buffer
	scan *scanner
}

func New() *Num {
	n := &Num{
		buf:  bytes.NewBuffer(make([]byte, 0, 4096)),
		scan: &scanner{},
	}
	n.scan.reset()
	return n
}

func (n *Num) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	var (
		i, j int
		c    byte
	)
	for i, c = range b {
		n.scan.bytes++
		switch n.scan.step(n.scan, int(c)) {
		case scanBeginNum:
			j = i
		case scanEndNum:
			n.buf.Write(Expand(b[j:i]))
			n.buf.WriteByte(b[i])
		case scanNotNum:
			n.buf.Write(b[j : i+1])
		case scanError:
			return i, n.scan.err
		default:
			if n.scan.parseState != parseNum {
				n.buf.WriteByte(c)
			}
		}
	}
	if n.scan.parseState == parseNum {
		n.buf.Write(Expand(b[j:]))
	}
	return len(b), nil
}

func (n *Num) WriteTo(w io.Writer) (int64, error) {
	return n.buf.WriteTo(w)
}

func (n *Num) Read(p []byte) (int, error) {
	return n.buf.Read(p)
}

// Expand a number with commas
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

// Trims decimals to sf significant figures.
func TrimDecimal(sf int, b []byte) []byte {
	n := bytes.IndexByte(b, '.') + 1
	if n == 0 {
		return b
	}
	if sf == 0 {
		if n == 1 {
			return roundDec(b[n:])
		}
		return b[:n-1]
	}
	if len(b)-n < sf {
		return b
	}
	return b[:n+sf]
}

// Round to negative infinity, for simplicity.
func roundDec(b []byte) []byte {
	if '0' <= b[0] && b[0] < '5' {
		return []byte{'0'}
	}
	if '5' <= b[0] && b[0] <= '9' {
		return []byte{'1'}
	}
	return []byte{}
}
