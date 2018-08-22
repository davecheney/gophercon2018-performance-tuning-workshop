# Compiler optimisations

This section covers three important optimisations that the Go compiler performs.

- Escape analysis
- Inlining
- Dead code elimination

## History of the Go compiler

The Go compiler started as a fork of the Plan9 compiler tool chain circa 2007. The compiler at that time bore a strong resemblence to Aho and Ullman's [_Dragon Book_][0].

In 2015 the then Go 1.5 compiler was mechanically translated from [C into Go][2].

A year later, Go 1.7 introduced a [new compiler backend][3] based on [SSA][1] techniques replaced the previous  Plan 9 style code generation. This new backend introduced many opportunities for generic and architecture specific optimistions.

## Escape analysis

The first optimisation we're doing to discuss is _escape analysis_. 

To illustrate what escape analysis does recall that the [Go spec][4] does not mention the heap or the stack. It only mentions that the language is garbage collected in the introduction, and gives no hints as to how this is to be achieved.

A compliant Go implementation of the Go spec _could_ store every allocation on the heap. That would put a lot of pressure on the the garbage collector, but it is in no way incorrect -- for several years, gccgo had very limited support for escape analysis so could effectively be considered to be operating in this mode. 

However, a goroutine's stack exists as a cheap place to store local variables; there is no need to garbage collect things on the stack. Therefore, where it is safe to do so, an allocation placed on the stack will be more efficient.

In some languages, for example C and C++, the choice of allocating on the stack or on the heap is a manual exercise for the programmer--heap allocations are made with `malloc` and `free`, stack allocation is via `alloca`. Mistakes using these mechanisms are a common cause of memory corruption bugs.

In Go, the compiler automatically moves a value to the heap if if lives beyond the lifetime of the function call. It is said that the value  _escapes_ to the heap.
```
type Foo struct {
	a, b, c, d int
}

func NewFoo() *Foo {
	return &Foo{a: 3, b: 1, c: 4, d: 7}
}
```
In this example the `Foo` allocated in `NewFoo` will be moved to the heap so its contents remain valid after `NewFoo` has returned.

This has been present since the earliest days of Go. It  isn't so much an optimisation as an automatic correctness feature. Accidentally returning the address of a stack allocated variable is not possible in Go.

But the compiler can also do the opposite; it can find things which would be assumed to be allocated on the heap, and move them to stack.

## Escape analysis (example)

Let's have a look at an example
```go
// Sum returns the sum of the numbers 1 to 100
func Sum() int {
        const count = 100
        numbers := make([]int, count)
        for i := range numbers {
                numbers[i] = i + 1
        }

        var sum int
        for _, i := range numbers {
                sum += i
        }
        return sum
}
```
`Sum` adds the `int`s between 1 and 100 and returns the result.

Because the `numbers` slice is only referenced inside `Sum`, the compiler will arrange to store the 100 integers for that slice on the stack, rather than the heap. There is no need to garbage collect `numbers`, it is automatically freed when `Sum` returns.

## Investigating escape analysis

Prove it!

To print the compilers escape analysis decisions, use the `-m` flag.
```
% go build -gcflags=-m examples/esc/sum.go
# command-line-arguments
examples/esc/sum.go:8:17: Sum make([]int, count) does not escape
examples/esc/sum.go:22:13: answer escapes to heap
examples/esc/sum.go:22:13: main ... argument does not escape
```
Line 8 shows the compiler has correctly deduced that the result of `make([]int, 100)` does not escape to the heap. The reason it did no

The reason line 22 reports that `answer` escapes to the heap is `fmt.Println` is a _variadic_ function. The parameters to a variadic function are _boxed_ into a slice, in this case a `[]interface{}`, so `answer` is placed into a interface value because it is referenced by the call to `fmt.Println`. Since Go 1.6 (??) the garbage collector requires _all_ values passed via an interface to be pointers, what the complier sees is _approximately_:
```
var answer = Sum()
fmt.Println([]interface{&answer}...)
```
We can confirm this using the `-gcflags="-m -m"` flag. Which returns
```
examples/esc/sum.go:22:13: answer escapes to heap
examples/esc/sum.go:22:13:      from ... argument (arg to ...) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:      from *(... argument) (indirection) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13:      from ... argument (passed to call[argument content escapes]) at examples/esc/sum.go:22:13
examples/esc/sum.go:22:13: main ... argument does not escape
```
In short, don't worry about line 22, its not important to this discussion.
### Execises

- Does this optimisation hold true for all values of `count`?
- Does this optimisation hold true if `count` is a variable, not a constant?
- Does this optimisation hold true if `count` is a parameter to `Sum`?

## Escape analysis (example)

This example is a little contrived. It is not intended to be real code, just an example.

```go
package main

import "fmt"

type Point struct{ X, Y int }

const Width = 640
const Height = 480

func Center(p *Point) {
        p.X = Width / 2
        p.Y = Height / 2
}

func NewPoint() {
        p := new(Point)
        Center(p)
        fmt.Println(p.X, p.Y)
}
```

`NewPoint` creates a new `*Point` value, `p`. We pass `p` to the `Center` function which moves the point to a position in the center of the screen. Finally we print the values of `p.X` and `p.Y`.

```
% go build -gcflags=-m examples/esc/center.go
# command-line-arguments
examples/esc/center.go:10:6: can inline Center
examples/esc/center.go:17:8: inlining call to Center
examples/esc/center.go:10:13: Center p does not escape
examples/esc/center.go:18:15: p.X escapes to heap
examples/esc/center.go:18:20: p.Y escapes to heap
examples/esc/center.go:16:10: NewPoint new(Point) does not escape
examples/esc/center.go:18:13: NewPoint ... argument does not escape
# command-line-arguments
```
Even though `p` was allocated with the `new` function, it will not be stored on the heap, because no reference `p` escapes the `Center` function.

- Question: What about line 20, if `p` doesn't escape, what is escaping to the heap?

 examples/esc/center.go:20: p.X escapes to heap
 examples/esc/center.go:20: p.Y escapes to heap

.link https://github.com/golang/go/issues/7714 Escape analysis is not perfect

## Exercise

- What is happening on line 25? Open up `examples/esc/sum.go` and see.
- Write a benchmark to provide that `Sum` does not allocate

## Inlining 

In Go function calls in have a fixed overhead; stack and preemption check.

Some of this is ameliorated by hardware branch predictors, but it's still a cost in terms of function size and clock cycles.

Inlining is the classical optimisation to avoid these costs. 

Inlining only works on _leaf_functions_, a function that does not call another. The justification for this is:

- If your function does a lot of work, then the preamble overhead will be negligible. That's why functions over a certain size (currently some count of instructions, plus a few operations which prevent inlining all together (eg. switch before Go 1.7)
- Small functions on the other hand pay a fixed overhead for a relatively small amount of useful work performed. These are the functions that inlining targets as they benefit the most. 

The other reason is it makes stack traces harder to follow.

## Inlining (example)

.play examples/max/max.go /START OMIT/,/END OMIT/

## Inlining (cont.)

Again we use the `-m` flag to view the compilers optimisation decision.
```
% go build -gcflags=-m examples/max/max.go 
# command-line-arguments
examples/max/max.go:4: can inline Max
examples/max/max.go:13: inlining call to Max
```
Compile `max.go` and see what the optimised version of `F()` became.

DEMO: `go`build`-gcflags="-m`-S"`examples/max/max.go`2>&1`|`less`

## Discussion

- Why did I declare `a` and `b` in `F()` to be constants?
- What happens if they are variables?
- What happens if they are passing into `F()` as parameters?

## Inlining future

A lot of what we've discussed with respect to leaf functions _may_ change in a future release of Go.

There has been work going on in the background since Go 1.8 to enable, so called, mid-stack inlining.

As a Go programmer, this should not be a cause for alarm, mid-

.link https://github.com/golang/proposal/blob/master/design/19348-midstack-inlining.md Proposal: Mid-stack inlining in the Go compiler

## Dead code elimination

Why is it important that `a` and `b` are constants?

After inlining, this is what the compiler saw

.play examples/max/max2.go /START OMIT/,/END OMIT/

- The call to `Max` has been inlined.
- If `a`>`b` then there is nothing to do, so the function returns. 
- If `a`<`b` then the branch is false and we fall through to `panic`
- But, because `a` and `b` are constants, we know that the branch will never be false, so the compiler can optimise `F()` to a return.

## Dead code elimination (cont.)

Dead code elimination work together with inlining to reduce the amount of code generated by removing loops and branches that are proven unreachable.

You can take advantage of this to implement expensive debugging, and hide it behind
```
const debug = false 
```
Combined with build tags this can be very useful.

Further reading:

.link http://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release Using // +build to switch between debug and release builds
.link http://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool How to use conditional compilation with the go build tool

### Compiler flags Exercises

Compiler flags are provided with:
```
go build -gcflags=$FLAGS
```

Investigate the operation of the following compiler functions:

- `-S` prints the (Go flavoured) assembly of the _package_ being compiled.
- `-l` controls the behaviour of the inliner; `-l` disables inlining, `-l`-l` increases it (more `-l` 's increases the compiler's appetite for inlining code). Experiment with the difference in compile time, program size, and run time.
- `-m` controls printing of optimisation decision like inlining, escape analysis. `-m`-m` prints more details about what the compiler was thinking.
- `-l`-N` disables all optimisations.

.link http://go-talks.appspot.com/github.com/rakyll/talks/gcinspect/talk.slide#1 Further reading: Codegen Inspection by Jaana Burcu Dogan

[0]: https://www.goodreads.com/book/show/112269.Principles_of_Compiler_Design
[1]: https://en.wikipedia.org/wiki/Static_single_assignment_form
[2]: https://golang.org/doc/go1.5#c
[3]: https://blog.golang.org/go1.7
[4]: https://golang.org/ref/spec
