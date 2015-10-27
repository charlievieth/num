package num

import (
	"bytes"
	"io/ioutil"
	"testing"
)

type testCase struct {
	In  string
	Out string
}

func TestNilNum(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()
	{
		var n Num
		n.Flush()
	}
	{
		var n Num
		n.Reset()
	}
	{
		var n Num
		n.init()
	}
	{
		var n Num
		n.Write([]byte{1, 2, 3})
	}
	{
		var n Num
		n.WriteTo(new(NopWriter))
	}
	{
		var n Num
		n.Read(make([]byte, 8))
	}
}

var numTests = []testCase{
	{
		In:  "a 123.0 x 1234 abc 12345 12345.a0 a 1234567.1234 abc",
		Out: "a 123.0 x 1,234 abc 12,345 12345.a0 a 1,234,567.1234 abc",
	},
	{
		In:  "a 123.0 x 1234 abc 12345 12345.a0 a 1234567.1234",
		Out: "a 123.0 x 1,234 abc 12,345 12345.a0 a 1,234,567.1234",
	},
}

func TestNum(t *testing.T) {
	buf := new(bytes.Buffer)
	for _, x := range numTests {
		num := New()
		n, err := num.Write([]byte(x.In))
		num.Flush()
		if err != nil {
			t.Errorf("Num: Error (%+v) %s", x, err)
		}
		if n != len(x.In) {
			t.Errorf("Num: Bad Write (%+v) Written: (%d) Expected: (%d)", x, n, len(x.In))
		}
		buf.Reset()
		num.WriteTo(buf)
		out := buf.String()
		if out != x.Out {
			t.Errorf("Num: Output (%+v) %s", x, out)
		}
		testStream(x, 2, t)
	}
}

func testStream(x testCase, writeSize int, t *testing.T) {
	num := New()
	buf := new(bytes.Buffer)
	var i int
	for i = 0; i < len(x.In)-writeSize; i += writeSize {
		num.Write([]byte(x.In[i : i+writeSize]))
		num.WriteTo(buf)
	}
	if i < len(x.In) {
		num.Write([]byte(x.In[i:]))
		num.WriteTo(buf)
	}
	num.Flush()
	num.WriteTo(buf)
	out := buf.String()
	if out != x.Out {
		t.Errorf("Stream (%+v) (%d):\n\tExp: %s\n\tOut: %s", x, writeSize, x.Out, out)
	}
}

var expandTests = []struct {
	In  string
	Exp string
}{
	{"0.001", "0.001"},
	{"0.01", "0.01"},
	{"0.1", "0.1"},
	{"1", "1"},
	{"12", "12"},
	{"123", "123"},
	{"1234", "1,234"},
	{"12345", "12,345"},
	{"12345.001", "12,345.001"},
	{"12345.12345", "12,345.12345"},
	{"1234567.1", "1,234,567.1"},
}

func TestExpand(t *testing.T) {
	for _, x := range expandTests {
		out := string(Expand([]byte(x.In)))
		if out != x.Exp {
			t.Errorf("Expand: Expected (%s) got (%s)", x.Exp, out)
		}
	}
}

func TestAppendExpand(t *testing.T) {
	var b []byte
	for _, x := range expandTests {
		b = appendExpand([]byte(x.In), b)
		if string(b) != x.Exp {
			t.Errorf("Expand: Expected (%s) got (%s)", x.Exp, string(b))
		}
		b = b[0:0]
	}
}

func BenchmarkExpand_All(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, x := range expandTests {
			Expand([]byte(x.In))
		}
	}
}

func BenchmarkExpand_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Expand([]byte("1234567890.1234"))
	}
}

func BenchmarkAppendExpand_All(b *testing.B) {
	dst := make([]byte, 65)
	for i := 0; i < b.N; i++ {
		for _, x := range expandTests {
			dst = appendExpand([]byte(x.In), dst[:0])
		}
	}
}

func BenchmarkAppendExpand_Large(b *testing.B) {
	dst := make([]byte, 65)
	for i := 0; i < b.N; i++ {
		dst = appendExpand([]byte("1234567890.1234"), dst[:0])
	}
}

func BenchmarkNum(b *testing.B) {
	src, err := ioutil.ReadFile("testdata/test.dat")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	n := New()
	for i := 0; i < b.N; i++ {
		n.Write(src)
		n.Reset()
	}
}

func BenchmarkNumStream(b *testing.B) {
	p, err := ioutil.ReadFile("testdata/test.dat")
	if err != nil {
		b.Fatal(err)
	}
	size := bytes.MinRead
	w := &NopWriter{}
	b.ResetTimer()
	n := New()
	for j := 0; j < b.N; j++ {
		var i int
		for i = 0; i < len(p)-size; i += size {
			n.Write(p[i : i+size])
			n.WriteTo(w)
		}
		if i < len(p) {
			n.Write(p[i:])
			n.WriteTo(w)
		}
		n.Reset()
	}
}

func writeStream(n *Num, size int, p []byte) {
	w := &NopWriter{}
	var i int
	for i = 0; i < len(p)-size; i += size {
		n.Write(p[i : i+size])
		n.WriteTo(w)
	}
	if i < len(p) {
		n.Write(p[i:])
		n.WriteTo(w)
	}
}

type NopWriter struct{}

func (n *NopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkScanner(b *testing.B) {
	src, err := ioutil.ReadFile("testdata/test.dat")
	if err != nil {
		b.Fatal(err)
	}
	scan := newScanner()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scan.reset()
		for _, c := range src {
			scan.bytes++
			switch scan.step(scan, int(c)) {
			case scanError:
				b.Fatal(scan.err)
			}
		}
	}
}
