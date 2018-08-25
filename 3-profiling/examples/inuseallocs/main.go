// A simple example to demonstrate the difference between alloc_objects and inuse_objects
package main

import (
	"math/rand"
	"runtime"

	"github.com/pkg/profile"
)

// ensure y is live beyond the end of main.
var y []byte

func main() {
	defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
	y = allocate(100000)
	runtime.GC()
}

// allocate allocates count byte slices and returns the first slice allocated.
func allocate(count int) []byte {
	var x [][]byte
	for i := 0; i < count; i++ {
		x = append(x, makeByteSlice())
	}
	return x[0]
}

// makeByteSlice returns a byte slice of a random length in the range [0, 16384).
func makeByteSlice() []byte {
	return make([]byte, rand.Intn(1<<14))
}
