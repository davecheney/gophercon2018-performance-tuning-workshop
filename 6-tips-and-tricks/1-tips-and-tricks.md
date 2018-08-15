# Tips and Tricks

This final section contains a number of tips to micro optimise Go code.

The key feature of Go that makes it a great fit for modern hardware are goroutines.

Goroutines are so easy to use, and so cheap to create, you could think of them as _almost_ free.

The Go runtime has been written for programs with tens of thousands of goroutines as the norm, hundreds of thousands are not unexpected.

However, each goroutine does consume a minimum amount of memory for the goroutine's stack which is currently at least 2k.

2048 * 1,000,000 goroutines == 2GB of memory, and they haven't done anything yet.

 Maybe this is a lot, maybe it isn't given the other usages of your application

# Know when to stop a goroutine

Goroutines are cheap to start and cheap to run, but they do have a finite cost in terms of memory footprint; you cannot create an infinite number of them.

Every time you use the `go` keyword in your program to launch a goroutine, you must *know* how, and when, that goroutine will exit.

If you don't know the answer, that's a potential memory leak.

In your design, some goroutines may run until the program exits. These goroutines are rare enough to not become an exception to the rule.

**Never start a goroutine without knowing how it will stop**

.link https://www.youtube.com/watch?v=yKQOunhhf4A&index=16&list=PLq2Nv-Sh8EbZEjZdPLaQt1qh_ohZFMDj8 Concurrency Made Easy (video)
.link https://dave.cheney.net/paste/concurrency-made-easy.pdf Concurrency Made Easy (slides)

## Go uses efficient network polling for some requests

The Go runtime handles network IO using an efficient operating system polling mechanism (kqueue, epoll, windows IOCP, etc). Many waiting goroutines will be serviced by a single operating system thread.

However, for local file IO, Go does not implement any IO polling. Each operation on a `*os.File` consumes one operating system thread while in progress.

Heavy use of local file IO can cause your program to spawn hundreds or thousands of threads; possibly more than your operating system allows.

Your disk subsystem does not expect to be able to handle hundreds or thousands of concurrent IO requests.

## Watch out for IO multipliers in your application

If you're writing a server process, its primary job is to multiplex clients connected over the network, and data stored in your application.

Most server programs take a request, do some processing, then return a result. This sounds simple, but depending on the result it can let the client consume a large (possibly unbounded) amount of resources on your server. Here are some things to pay attention to:

- The amount of IO requests per incoming request; how many IO events does a single client request generate? It might be on average 1, or possibly less than one if many requests are served out of a cache.
- The amount of reads required to service a query; is it fixed, N+1, or linear (reading the whole table to generate the last page of results).

If memory is slow, relatively speaking, then IO is so slow that you should avoid doing it at all costs. Most importantly avoid doing IO in the context of a request—don't make the user wait for your disk subsystem to write to disk, or even read.

* Use streaming IO interfaces

Where-ever possible avoid reading data into a `[]byte` and passing it around. 

Depending on the request you may end up reading megabytes (or more!) of data into memory. This places huge pressure on the GC, which will increase the average latency of your application.

Instead use `io.Reader` and `io.Writer` to construct processing pipelines to cap the amount of memory in use per request.

For efficiency, consider implementing `io.ReaderFrom` / `io.WriterTo` if you use a lot of `io.Copy`. These interface are more efficient and avoid copying memory into a temporary buffer.

## Timeouts, timeouts, timeouts

Never start an IO operating without knowing the maximum time it will take.

You need to set a timeout on every network request you make with `SetDeadline`, `SetReadDeadline`, `SetWriteDeadline`.

You need to limit the amount of blocking IO you issue. Use a pool of worker goroutines, or a buffered channel as a semaphore.

.code examples/semaphore.go /START OMIT/,/END OMIT/

## Defer is expensive, or is it?

`defer` is expensive because it has to record a closure for defer's arguments.
```
defer mu.Unlock()
```
is equivalent to
```
defer func() {
	mu.Unlock()
}()
```
`defer` is expensive if the work being done is small, the classic example is `defer` ing a mutex unlock around a struct variable or map lookup. You may choose to avoid `defer` in those situations.

This is a case where readability and maintenance is sacrificed for a performance win. 

#### Always revisit these decisions.

.link https://github.com/golang/go/issues/9704#issuecomment-251003577

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

Old version of Go receive no updates. Do not use them. Use the latest and you will get the best performance.

.link http://dave.cheney.net/2016/04/02/go-1-7-toolchain-improvements Go 1.7 toolchain improvements
.link http://dave.cheney.net/2016/09/18/go-1-8-performance-improvements-one-month-in Go 1.8 performance improvements

## Discussion

Any questions?

