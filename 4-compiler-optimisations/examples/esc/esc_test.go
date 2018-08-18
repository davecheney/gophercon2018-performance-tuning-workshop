package main

import "testing"

var result int

func BenchmarkSum(b *testing.B) {
	b.ReportAllocs()
	var r int
	for i := 0; i < b.N; i++ {
		r = Sum()
	}
	result = r
}
