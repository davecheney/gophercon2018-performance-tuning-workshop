package fib

import "testing"

// START OMIT
func Fib(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}

func BenchmarkFibWrong(b *testing.B) {
	Fib(b.N)
}

func BenchmarkFibWrong2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(n)
	}
}

// END OMIT
