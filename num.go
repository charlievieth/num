package num

import (
	"bytes"
	"io"
)

const DefaultBufferSize = 4096

type Num struct {
	buf     *bytes.Buffer
	scan    *scanner
	partial []byte
}

func New() *Num {
	n := &Num{
		buf:  bytes.NewBuffer(make([]byte, 0, DefaultBufferSize)),
		scan: &scanner{},
	}
	n.scan.reset()
	return n
}

func (n *Num) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	var b []byte
	if len(n.partial) != 0 {
		b = append(n.partial, p...)
		n.partial = n.partial[0:0]
	} else {
		b = p
	}
	var j int
	dst := make([]byte, 65)
	for i, c := range b {
		n.scan.bytes++
		switch n.scan.step(n.scan, int(c)) {
		case scanBeginNum:
			j = i
		case scanEndNum:
			dst = appendExpand(b[j:i], dst[:0])
			n.buf.Write(dst)
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
		n.partial = append(n.partial, b[j:]...)
	}
	return len(b), nil
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
		return b
	}
	c := (n % 3)
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
