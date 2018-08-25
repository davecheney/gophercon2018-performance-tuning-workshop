package q

// START1 OMIT
func BenchmarkExpensive(b *testing.B) {
	boringAndExpensiveSetup()
	b.ResetTimer() // HL
	for n := 0; n < b.N; n++ {
		// function under test
	}
}

// END1 OMIT

// START2 OMIT
func BenchmarkComplicated(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer() // HL
		complicatedSetup()
		b.StartTimer() // HL
		// function under test
	}
}

// END2 OMIT
