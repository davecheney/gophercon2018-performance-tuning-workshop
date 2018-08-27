package isspace

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var B bool

func BenchmarkIsSpaceTrue(b *testing.B) {
	benchmarkIsSpace(b, '\t')
}

func BenchmarkIsSpaceFalse(b *testing.B) {
	benchmarkIsSpace(b, 'a')
}

func benchmarkIsSpace(b *testing.B, c byte) {
	var r bool
	for i := 0; i < b.N; i++ {
		r = isSpace(c)
	}
	B = r
}

var Result int

func BenchmarkCountWhitespacePrideAndPrejudice(b *testing.B) {
	buf := setup(b, "testdata/pnp.txt")
	var r int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = 0
		for _, c := range buf {
			if isSpace(c) {
				r++
			}
		}
		Result = r
	}
}

func setup(t testing.TB, path string) []byte {
	buf, err := ioutil.ReadFile(path)
	check(t, err)
	return buf
}

func check(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsSpace(t *testing.T) {
	tests := [256]bool{
		'\t': true,
		'\r': true,
		'\n': true,
		' ':  true,
	}
	for tc, want := range tests {
		t.Run(fmt.Sprintf("%0x", tc), func(t *testing.T) {
			got := isSpace(byte(tc))
			if got != want {
				t.Fatalf("expected: %v, got %v", want, got)
			}
		})
	}
}

func TestCountPrideAndPrejudice(t *testing.T) {
	buf := setup(t, "testdata/pnp.txt")
	got := 0
	want := 140795
	for _, c := range buf {
		if isSpace(c) {
			got++

		}
	}
	assert(t, got, want)
}

func assert(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("expected: %d, got %d", want, got)
	}
}
