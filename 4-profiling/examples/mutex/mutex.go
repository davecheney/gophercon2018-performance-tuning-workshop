package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/pkg/profile"
)

// aapated from https://commandercoriander.net/blog/2018/04/10/dont-lock-around-io/

const payloadBytes = 1024 * 1024

var (
	mu    sync.Mutex
	count int
)

func main() {
	defer profile.Start(profile.MutexProfile).Stop()
	http.HandleFunc("/", root)
	http.ListenAndServe(":9999", nil)
}

func root(w http.ResponseWriter, r *http.Request) {

	mu.Lock()
	defer mu.Unlock()

	count++

	msg := []byte(strings.Repeat(fmt.Sprintf("%d", count), payloadBytes))
	w.Write(msg)
}
