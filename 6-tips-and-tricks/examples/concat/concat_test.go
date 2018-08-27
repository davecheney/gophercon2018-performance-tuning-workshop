package main

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

// sink to ensure the compiler does not optimise away dead assignments.
var Result string

// fake up some values for request and client.
func setup(b *testing.B) (struct{ ID string }, net.Listener) {
	request := struct {
		ID string
	}{"9001"}
	client, err := net.Listen("tcp", ":0")
	if err != nil {
		b.Fatal(err)
	}
	return request, client
}

func BenchmarkConcatenate(b *testing.B) {
	request, client := setup(b)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()
	var r string
	for n := 0; n < b.N; n++ {
		s := request.ID
		s += " " + client.Addr().String()
		s += " " + time.Now().String()
		r = s
	}
	Result = r
}

func BenchmarkFprintf(b *testing.B) {
	request, client := setup(b)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()
	var r string
	for n := 0; n < b.N; n++ {
		var b bytes.Buffer
		fmt.Fprintf(&b, "%s %v %v", request.ID, client.Addr(), time.Now())
		r = b.String()
	}
	Result = r
}

func BenchmarkSprintf(b *testing.B) {
	request, client := setup(b)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()
	var r string
	for n := 0; n < b.N; n++ {
		r = fmt.Sprintf("%s %v %v", request.ID, client.Addr(), time.Now())
	}
	Result = r
}

func BenchmarkStrconv(b *testing.B) {
	request, client := setup(b)
	defer client.Close()

	b.ResetTimer()
	b.ReportAllocs()
	var r string
	for n := 0; n < b.N; n++ {
		b := make([]byte, 0, 40)
		b = append(b, request.ID...)
		b = append(b, ' ')
		b = append(b, client.Addr().String()...)
		b = append(b, ' ')
		b = time.Now().AppendFormat(b, "2006-01-02 15:04:05.999999999 -0700 MST")
		r = string(b)
	}
	Result = r
}
