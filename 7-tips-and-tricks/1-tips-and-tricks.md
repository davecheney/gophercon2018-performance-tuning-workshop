# Tips and Tricks

This section contains a number of tips to optimise Go code.

## Reduce allocations

Make sure your APIs allow the caller to reduce the amount of garbage generated.

Consider these two Read methods
```go
func (r *Reader) Read() ([]byte, error)
func (r *Reader) Read(buf []byte) (int, error)
```

The first Read method takes no arguments and returns some data as a `[]byte`. The second takes a `[]byte` buffer and returns the amount of bytes read.

The first Read method will _always_ allocate a buffer, putting pressure on the GC. The second fills the buffer it was given.

_Exercise_: Can you name examples in the std lib which follow this pattern?

## strings vs []bytes

In Go `string` values are immutable, `[]byte` are mutable.

Most programs prefer to work `string`, but most IO is done with `[]byte`.

Avoid `[]byte` to string conversions wherever possible, this normally means picking one representation, either a `string` or a `[]byte` for a value. Often this will be `[]byte` if you read the data from the network or disk.

The [`bytes`][2] package contains many of the same operations— `Split`, `Compare`, `HasPrefix`, `Trim`, etc—as the [`strings`][3] package.

Under the hood `strings` uses same assembly primitives as the `bytes` package.

## Using []byte as a map key

It is very common to use a `string` as a map key, but often you have a `[]byte`.

The compiler implements a specific optimisation for this case
```go
var m map[string]string
v, ok := m[string(bytes)]
```

This will avoid the conversion of the byte slice to a string for the map lookup. This is very specific, it won't work if you do something like
```go
key := string(bytes)
val, ok := m[key]
```
## Optimise string concatenation

Go strings are immutable. Concatenating two strings generates a third. Which of the following is fastest?
```go
s := request.ID
s += " " + client.Addr().String()
s += " " + time.Now().String()
r = s
```

```go
var b bytes.Buffer
fmt.Fprintf(&b, "%s %v %v", request.ID, client.Addr(), time.Now())
r = b.String()
```

```go
r = fmt.Sprintf("%s %v %v", request.ID, client.Addr(), time.Now())
```

```go
b := make([]byte, 0, 40)
b = append(b, request.ID...)
b = append(b, ' ')
b = append(b, client.Addr().String()...)
b = append(b, ' ')
b = time.Now().AppendFormat(b, "2006-01-02 15:04:05.999999999 -0700 MST")
r = string(b)
```

```
% go test -bench=. ./examples/concat/
```

_Bonus question_: All four benchmarks got _slower_ from Go 1.10 to Go 1.11 _on this Mac_. Any ideas why?

## Preallocate slices if the length is known

Append is convenient, but wasteful.

Slices grow by doubling up to 1024 elements, then by approximately 25% after that. What is the capacity of `b` after we append one more item to it?
```go
func main() {
        b := make([]int, 1024)
        b = append(b, 99)
        fmt.Println("len:", len(b), "cap:", cap(b))
}
```
If you use the append pattern you could be copying a lot of data and creating a lot of garbage.

If know know the length of the slice beforehand, then pre-allocate the target to avoid copying and to make sure the target is exactly the right size.

_Before:_
```go
var s []string
for _, v := range fn() {
        s = append(s, v)
}
return s
```
_After:_
```go
vals := fn()
s := make([]string, len(vals))
for i, v := range vals {
        s[i] = v
}
return s
```

_Exercise_: write a benchmark to see which is faster.

## Goroutines

The key feature of Go that makes it a great fit for modern hardware are goroutines. Goroutines are so easy to use, and so cheap to create, you could think of them as _almost_ free.

The Go runtime has been written for programs with tens of thousands of goroutines as the norm, hundreds of thousands are not unexpected.

However, each goroutine does consume a minimum amount of memory for the goroutine's stack which is currently at least 2k.

2048 * 1,000,000 goroutines == 2GB of memory, and they haven't done anything yet.

Maybe this is a lot, maybe it isn't, given the other consumers of memory in your application. 

## Know when a goroutine is going to stop

Goroutines are cheap to start and cheap to run, but they do have a finite cost in terms of memory footprint; you cannot create an infinite number of them.

Every time you use the `go` keyword in your program to launch a goroutine, you must _know_ how, and when, that goroutine will exit.

If you don't know the answer, that's a potential memory leak.

In your design, some goroutines may run until the program exits. These goroutines are rare enough to not become an exception to the rule.

**Never start a goroutine without knowing how it will stop**

A good way to achieve this is to use something like [run.Group][4], [workgroup.Group][5], or similar.

Peter Bourgon has a great presentation on the design behing run.Group from GopherCon EU

### Further reading

- [Concurrency Made Easy][0] (video)
- [Concurrency Made Easy][1] (slides)

## Go uses efficient network polling for some requests

The Go runtime handles network IO using an efficient operating system polling mechanism (kqueue, epoll, windows IOCP, etc). Many waiting goroutines will be serviced by a single operating system thread.

However, for local file IO--_with the exception of pipes_--Go does not implement any IO polling. Each operation on a `*os.File` consumes one operating system thread while in progress.

Heavy use of local file IO can cause your program to spawn hundreds or thousands of threads; possibly more than your operating system allows.

Your disk subsystem does not expect to be able to handle hundreds or thousands of concurrent IO requests.

## Watch out for IO multipliers in your application

If you're writing a server process, its primary job is to multiplex clients connected over the network, and data stored in your application.

Most server programs take a request, do some processing, then return a result. This sounds simple, but depending on the result it can let the client consume a large (possibly unbounded) amount of resources on your server. Here are some things to pay attention to:

- The amount of IO requests per incoming request; how many IO events does a single client request generate? It might be on average 1, or possibly less than one if many requests are served out of a cache.
- The amount of reads required to service a query; is it fixed, N+1, or linear (reading the whole table to generate the last page of results).

If memory is slow, relatively speaking, then IO is so slow that you should avoid doing it at all costs. Most importantly avoid doing IO in the context of a request—don't make the user wait for your disk subsystem to write to disk, or even read.

## Use streaming IO interfaces

Where-ever possible avoid reading data into a `[]byte` and passing it around. 

Depending on the request you may end up reading megabytes (or more!) of data into memory. This places huge pressure on the GC, which will increase the average latency of your application.

Instead use `io.Reader` and `io.Writer` to construct processing pipelines to cap the amount of memory in use per request.

For efficiency, consider implementing `io.ReaderFrom` / `io.WriterTo` if you use a lot of `io.Copy`. These interface are more efficient and avoid copying memory into a temporary buffer.

## Timeouts, timeouts, timeouts

Never start an IO operating without knowing the maximum time it will take.

You need to set a timeout on every network request you make with `SetDeadline`, `SetReadDeadline`, `SetWriteDeadline`.

You need to limit the amount of blocking IO you issue. Use a pool of worker goroutines, or a buffered channel as a semaphore.
```go
var semaphore = make(chan struct{}, 10)

func processRequest(work *Work) {
        semaphore <- struct{}{} // acquire semaphore
        // process request
        <-semaphore // release semaphore
}
```

## Defer is expensive, or is it?

`defer` has a cost because it has to record a closure for its arguments.
```
defer mu.Unlock()
```
is equivalent to
```
defer func() {
	mu.Unlock()
}()
```
`defer` can be expensive if the work being done is small, the classic example is `defer` ing a mutex unlock around a struct variable or map lookup. You may choose to avoid `defer` in those situations.

This is a case where readability and maintenance is sacrificed for a performance win. 

#### Always revisit these decisions.

However, in revising this presentation, I have not been able to write a program that demonstrates a measurable cost from using `defer` -- the compiler is getting very good at eliminating the cost of using defer.

## Avoid Finalisers

Finalisation is a technique to attach behaviour to an object which is just about to be garbage collected.

Thus, finalisation is non deterministic. 

For a finaliser to run, the object must not be reachable by _anything_. If you accidentally keep a reference to the object in the map, it won't be finalised.

Finalisers run as part of the gc cycle, which means it is unpredictable when they will run and puts them at odds with the goal of reducing gc operation.

A finaliser may not run for a long time if you have a large heap and have tuned your appliation to create minimal garbage.

## Minimise cgo

cgo allows Go programs to call into C libraries. 

C code and Go code live in two different universes, cgo traverses the boundary between them.

This transition is not free and depending on where it exists in your code, the cost could be substantial.

cgo calls are similar to blocking IO, they consume a thread during operation.

Do not call out to C code in the middle of a tight loop.

## Actually, avoid cgo

cgo has a high overhead.

For best performance I recommend avoiding cgo in your applications.

- If the C code takes a long time, cgo overhead is not as important.
- If you're using cgo to call a very short C function, where the overhead is the most noticeable, rewrite that code in Go -- by definition it's short.
- If you're using a large piece of expensive C code is called in a tight loop, why are you using Go?

Is there anyone who's using cgo to call expensive C code frequently?

.link http://dave.cheney.net/2016/01/18/cgo-is-not-go Further reading: cgo is not Go.

## Always use the latest released version of Go

Old versions of Go will never get better. They will never get bug fixes or optimisations.

- Go 1.4 should not be used.
- Go 1.5 and 1.6 had a slower compiler, but it produces faster code, and has a faster GC.
- Go 1.7 delivered roughly a 30% improvement in compilation speed over 1.6, a 2x improvement in linking speed (better than any previous version of Go).
- Go 1.8 will deliver a smaller improvement in compilation speed (at this point), but a significant improvement in code quality for non Intel architectures.
- Go 1.9, 1.10, 1.11 continue to drive down the GC pause time and improve the quality of generated code.

Old version of Go receive no updates. Do not use them. Use the latest and you will get the best performance.

## Discussion

Any questions?

[0]: https://www.youtube.com/watch?v=yKQOunhhf4A&index=16&list=PLq2Nv-Sh8EbZEjZdPLaQt1qh_ohZFMDj8
[1]: https://dave.cheney.net/paste/concurrency-made-easy.pdf
[2]: https://golang.org/pkg/bytes
[3]: https://golang.org/pkg/strings
[4]: https://github.com/oklog/run
[5]: https://github.com/heptio/workgroup