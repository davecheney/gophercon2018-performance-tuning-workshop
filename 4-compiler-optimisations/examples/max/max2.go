// +build none
package main

// START OMIT
func F() {
	const a, b = 100, 20
	if a > b {
		return
	}
	panic(b)
}

// END OMIT

func main() {
	F()
}
