// Package num provides tools for adding thousands separators to numbers.
package num

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
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

// Reset, resets the internal state of Num.
func (n *Num) Reset() {
	n.buf.Reset()
	if n.scan != nil {
		n.scan.reset()
	}
	n.partial = n.partial[:0]
	n.scratch = n.scratch[:0]
}

// Write, writes formats any numbers in p and writes the results to the
// internal buffer.  A state machine is used to keep track of the format
// state - so if a write ends partially through a number the number will
// be formatted on the next call to Write or Flush.
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
			n.scratch = formatNumber(n.scratch[:0], b[lastWrite:i])
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

// Flush, formats any partially read numbers and flushes them into the internal
// buffer.
func (n *Num) Flush() error {
	if len(n.partial) == 0 {
		return nil
	}
	if n.scan.parseState == parseNum {
		n.scratch = formatNumber(n.scratch[:0], n.partial)
		n.buf.Write(n.scratch)
		n.scan.reset()
		n.partial = n.partial[:0]
	}
	return nil
}

// WriteTo, flushes any partial numbers and writes the contents of Num's
// internal buffer to w.
func (n *Num) WriteTo(w io.Writer) (int64, error) {
	if err := n.Flush(); err != nil {
		return 0, err
	}
	return n.buf.WriteTo(w)
}

// Read, flushes any partial numbers and reads up to len(p) bytes from the
// internal buffer into p.
func (n *Num) Read(p []byte) (int, error) {
	if err := n.Flush(); err != nil {
		return 0, err
	}
	return n.buf.Read(p)
}

// An Encoder is a stream formatter.
type Encoder struct {
	w   io.Writer
	n   Num
	buf []byte
	err error
}

// NewEncoder, returns an Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode, reads from r formatting any numbers and writes the results to the
// underlying io.Writer.
func (e *Encoder) Encode(r io.Reader) error {
	if e.err != nil {
		return e.err
	}
	const bufSize = 32 * 1024
	if len(e.buf) < bufSize {
		e.buf = make([]byte, bufSize)
	}
	for {
		n, err := r.Read(e.buf)
		if err == nil || err == io.EOF {
			e.stream(e.buf[:n])
		}
		if err != nil {
			break
		}
	}
	if e.err != nil {
		return e.err
	}
	if err := e.n.Flush(); err != nil {
		e.err = err
		return err
	}
	if _, err := e.n.WriteTo(e.w); err != nil {
		e.err = err
	}
	return e.err
}

func (e *Encoder) stream(p []byte) error {
	if e.err != nil {
		return e.err
	}
	_, err := e.n.Write(p)
	if err != nil {
		e.err = err
	}
	return e.writeTo()
}

func (e *Encoder) writeTo() error {
	if e.err != nil {
		return e.err
	}
	_, err := e.n.WriteTo(e.w)
	if err != nil {
		e.err = err
	}
	return err
}

// small returns the string for an i with 0 <= i < nSmalls.
func small(i int) string {
	if i < 10 {
		return digits[i : i+1]
	}
	return smallsString[i*2 : i*2+2]
}

const nSmalls = 100

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

const digits = "0123456789abcdefghijklmnopqrstuvwxyz"

func FormatInt(val int64) string {
	if 0 <= val && val < nSmalls {
		return small(int(val))
	}
	return formatBits(uint64(val), val < 0)
}

func FormatUint(val uint64) string {
	if val < nSmalls {
		return small(int(val))
	}
	return formatBits(val, val < 0)
}

func formatBits(val uint64, neg bool) string {
	var buf [26]byte
	if neg {
		val = -val
	}
	i := len(buf) - 1
	n := 0
	for val >= 10 {
		buf[i] = byte(val%10 + '0')
		i--
		val /= 10
		n++
		if n == 3 {
			buf[i] = ','
			i--
			n = 0
		}
	}
	buf[i] = byte(val + '0')
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// func FormatFloat(f float64, fmt byte, prec, bitSize int) string {
func FormatFloat(f float64, fmt byte, prec, bitSize int) string {
	s := strconv.FormatFloat(f, fmt, prec, bitSize)
	if before, after, ok := strings.Cut(s, "."); ok {
		s = string(formatNumber(nil, []byte(before))) + "." + after
	}
	return s
}

// Format, adds thousands separators to string s.  An error is returned is s
// is not a number.
func Format(s string) (string, error) {
	b := []byte(s)
	if !isNumber(b) {
		return "", errors.New("num: cannot format string: " + s)
	}
	var a [64]byte
	return string(formatNumber(a[:0], b)), nil
}

// AppendFormat, adds thousands separators to byte slice b and appends the
// results to dst.  If b is not a number it is not appended to dst.
func AppendFormat(dst, b []byte) []byte {
	if !isNumber(b) {
		return dst
	}
	return formatNumber(dst, b)
}

func isNumber(b []byte) bool {
	if len(b) == 0 || b[0] == '.' {
		return false
	}
	for _, c := range b {
		if ('0' > c || c > '9') && c != '.' {
			return false
		}
	}
	return true
}

func formatNumber(dst, b []byte) []byte {
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
