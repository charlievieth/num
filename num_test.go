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

const benchmarkResultInput = `
BenchmarkGOROOT-4                 	30000000	        43.2 ns/op	      12 B/op	       0 allocs/op
BenchmarkCorpus_IndexFiles-4      	       5	 332273086 ns/op	174468006 B/op	  780752 allocs/op
BenchmarkCorpus_FindFiles-4       	       5	 319671269 ns/op	167605459 B/op	  680913 allocs/op
BenchmarkCorpus_FindName-4        	       5	 266548976 ns/op	95192496 B/op	  468363 allocs/op
BenchmarkCorpusUpdate_IndexFiles-4	      10	 122262840 ns/op	43831176 B/op	  380475 allocs/op
BenchmarkCorpusUpdate_FindFiles-4 	     100	  10244592 ns/op	 4599907 B/op	   47871 allocs/op
BenchmarkCorpusUpdate_FindName-4  	     200	   8176888 ns/op	 3329827 B/op	   42150 allocs/op
`

const benchmarkResultOutput = `
BenchmarkGOROOT-4                 	30,000,000	        43.2 ns/op	      12 B/op	       0 allocs/op
BenchmarkCorpus_IndexFiles-4      	       5	 332,273,086 ns/op	174,468,006 B/op	  780,752 allocs/op
BenchmarkCorpus_FindFiles-4       	       5	 319,671,269 ns/op	167,605,459 B/op	  680,913 allocs/op
BenchmarkCorpus_FindName-4        	       5	 266,548,976 ns/op	95,192,496 B/op	  468,363 allocs/op
BenchmarkCorpusUpdate_IndexFiles-4	      10	 122,262,840 ns/op	43,831,176 B/op	  380,475 allocs/op
BenchmarkCorpusUpdate_FindFiles-4 	     100	  10,244,592 ns/op	 4,599,907 B/op	   47,871 allocs/op
BenchmarkCorpusUpdate_FindName-4  	     200	   8,176,888 ns/op	 3,329,827 B/op	   42,150 allocs/op
`

var numTests = []testCase{
	{
		In:  "a 123.0 x 1234 abc 12345 12345.a0 a 1234567.1234 317659251 abc",
		Out: "a 123.0 x 1,234 abc 12,345 12345.a0 a 1,234,567.1234 317,659,251 abc",
	},
	{
		In:  "a 123.0 x 1234 abc 12345 12345.a0 a 1234567.1234",
		Out: "a 123.0 x 1,234 abc 12,345 12345.a0 a 1,234,567.1234",
	},
	{
		In:  benchmarkResultInput,
		Out: benchmarkResultOutput,
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
