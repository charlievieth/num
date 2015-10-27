package num

import (
	"io/ioutil"
	"testing"
)

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
	for i := 0; i < b.N; i++ {
		n := New()
		n.Write(src)
	}
}
