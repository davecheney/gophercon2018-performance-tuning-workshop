package fib

import "testing"

// Fib computes the n'th number in the Fibonacci series.
func Fib(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return Fib(n-1) + Fib(n-2)
	}
}

func BenchmarkFib20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(20) // run the Fib function b.N times
	}
}

func benchmarkFib1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Fib(1)
	}
}

func TestFib(t *testing.T) {
	fibs := []int{0, 1, 1, 2, 3, 5, 8, 13, 21}
	for n, want := range fibs {
		got := Fib(n)
		if want != got {
			t.Errorf("Fib(%d): want %d, got %d", n, want, got)
		}
	}
}

func Fib2(n int) int {
	a, b := 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}
