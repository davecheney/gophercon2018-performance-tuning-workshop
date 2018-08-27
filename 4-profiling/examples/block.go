package main

import (
	"fmt"
	"time"

	"github.com/pkg/profile"
)

func generate(in <-chan int, out chan<- int) {
	i := <-in
	time.Sleep(100 * time.Millisecond)
	out <- i + i
}

func main() {
	defer profile.Start(profile.BlockProfile).Stop()
	in := make(chan int, 1)
	in <- 1
	var out chan int
	for i := 0; i < 5; i++ {
		out = make(chan int)
		go generate(in, out)
		in = out
	}
	fmt.Println(<-out)
}
