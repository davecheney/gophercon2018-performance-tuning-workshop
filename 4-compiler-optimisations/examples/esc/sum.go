package main

import "fmt"

// START OMIT

// Sum returns the sum of the numbers 1 to 100
func Sum() int {
	const count = 100
	numbers := make([]int, 100)
	for i := range numbers {
		numbers[i] = i + 1
	}

	var sum int
	for _, i := range numbers {
		sum += i
	}
	return sum
}

// END OMIT

func main() {
	fmt.Println(Sum())
}
