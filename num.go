package num

import (
	"bytes"
	"io"
)

type Num struct {
	buf     bytes.Buffer
	scan    *scanner
	partial []byte
	scratch []byte
}

func New() *Num {
	return &Num{scan: newScanner()}
}

func (n *Num) init() {
	if n.scan == nil {
		n.scan = newScanner()
	}
	if cap(n.scratch) == 0 {
		n.scratch = make([]byte, 0, 64)
	}
}

func (n *Num) Reset() {
	n.buf.Reset()
	if n.scan != nil {
		n.scan.reset()
	}
	n.partial = n.partial[:0]
	n.scratch = n.scratch[:0]
}

func (n *Num) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n.init()
	start := len(n.partial)
	var b []byte
	if start != 0 {
		n.partial = append(n.partial, p...)
		b = n.partial
	} else {
		b = p
	}
	var lastWrite int
	for i := start; i < len(b); i++ {
		n.scan.bytes++
		switch n.scan.step(n.scan, int(b[i])) {
		case scanBeginNum:
			n.buf.Write(b[lastWrite:i])
			lastWrite = i
		case scanEndNum:
			n.scratch = appendExpand(b[lastWrite:i], n.scratch[:0])
			n.buf.Write(n.scratch)
			lastWrite = i
		case scanError:
			return i, n.scan.err
		}
	}
	if n.scan.parseState == parseNum {
		n.partial = append(n.partial[:0], b[lastWrite:]...)
	} else {
		n.buf.Write(b[lastWrite:])
		n.partial = n.partial[:0]
	}
	return len(p), nil
}

func (n *Num) Flush() error {
	if len(n.partial) == 0 {
		return nil
	}
	if n.scan.parseState == parseNum {
		n.scratch = appendExpand(n.partial, n.scratch[:0])
		n.buf.Write(n.scratch)
		n.scan.reset()
	}
	return nil
}

func (n *Num) WriteTo(w io.Writer) (int64, error) {
	return n.buf.WriteTo(w)
}

func (n *Num) Read(p []byte) (int, error) {
	return n.buf.Read(p)
}

func appendExpand(b, dst []byte) []byte {
	n := bytes.IndexByte(b, '.')
	if n == -1 {
		n = len(b)
	}
	if n <= 3 {
		return append(dst, b...)
	}
	c := (n % 3)
	if c == 0 {
		c = 3
	}
	dst = append(dst, b[:c]...)
	for i := c; i < n; i += 3 {
		dst = append(dst, ',')
		dst = append(dst, b[i:i+3]...)
	}
	return append(dst, b[n:]...)
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
	buf := make([]byte, len(b)+(n/3))
	var o int
	for i := 0; i < n; i++ {
		if c == 3 {
			c = 0
			buf[o] = ','
			o++
		}
		buf[o] = b[i]
		o++
		c++
	}
	copy(buf[o:], b[n:])
	return buf
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
