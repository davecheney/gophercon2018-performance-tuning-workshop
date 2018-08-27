# Compiler optimisations

This section covers three important optimisations that the Go compiler performs.

- Escape analysis
- Inlining
- Dead code elimination

## History of the Go compiler

The Go compiler started as a fork of the Plan9 compiler tool chain circa 2007. The compiler at that time bore a strong resemblance to Aho and Ullman's [_Dragon Book_][0].

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

The reason line 22 reports that `answer` escapes to the heap is `fmt.Println` is a _variadic_ function. The parameters to a variadic function are _boxed_ into a slice, in this case a `[]interface{}`, so `answer` is placed into a interface value because it is referenced by the call to `fmt.Println`. Since Go 1.6 (??) the garbage collector requires _all_ values passed via an interface to be pointers, what the compiler sees is _approximately_:
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
### Exercises

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

_Exercise_: Write a benchmark to provide that `Sum` does not allocate

## Inlining 

In Go function calls in have a fixed overhead; stack and preemption check.

Some of this is ameliorated by hardware branch predictors, but it's still a cost in terms of function size and clock cycles.

Inlining is the classical optimisation to avoid these costs. 

Inlining only works on _leaf functions_, a function that does not call another. The justification for this is:

- If your function does a lot of work, then the preamble overhead will be negligible. That's why functions over a certain size (currently some count of instructions, plus a few operations which prevent inlining all together (eg. switch before Go 1.7)
- Small functions on the other hand pay a fixed overhead for a relatively small amount of useful work performed. These are the functions that inlining targets as they benefit the most. 

The other reason is that heavy inlining makes stack traces harder to follow. 

## Inlining (example)

```go
func Max(a, b int) int {
        if a > b {
                return a
        }
        return b
}

func F() {
        const a, b = 100, 20
        if Max(a, b) == b {
                panic(b)
        }
}
```
Again we use the `-gcflags=-m` flag to view the compilers optimisation decision.
```
% go build -gcflags=-m examples/max/max.go
# command-line-arguments
examples/max/max.go:3:6: can inline Max
examples/max/max.go:12:8: inlining call to Max
```
The compiler printed two lines. 

- The first at line 3, the declaration of `Max`, telling us that it can be inlined.
- The second is reporting that the body of `Max` has been inlined into the caller at line 12.

_Exercise_: _Without_ using the `//go:noinline` comment, rewrite `Max` such that it still returns the right answer, but is no longer considered inlineable by the compiler.

### What does inlining look like?

Compile `max.go` and see what the optimised version of `F()` became.
```
% go build -gcflags=-S examples/max/max.go 2>&1 | grep -A5 '"".F STEXT'
"".F STEXT nosplit size=1 args=0x0 locals=0x0
        0x0000 00000 (/Users/dfc/devel/gophercon2018-performance-tuning-workshop/4-compiler-optimisations/examples/max/max.go:10)       TEXT    "".F(SB), NOSPLIT, $0-0
        0x0000 00000 (/Users/dfc/devel/gophercon2018-performance-tuning-workshop/4-compiler-optimisations/examples/max/max.go:10)       FUNCDATA        $0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
        0x0000 00000 (/Users/dfc/devel/gophercon2018-performance-tuning-workshop/4-compiler-optimisations/examples/max/max.go:10)       FUNCDATA        $1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
        0x0000 00000 (<unknown line number>)    RET
        0x0000 c3
```
This is the body of `F` once `Max` has been inlined into it -- there's nothing happening in this function. I know there's a lot of text on the screen for nothing, but take my word for it, the only  thing happening is the `RET`. In effect `F` became:
```go
func F() {
        return
}
```

_Note_: The output from `-S` is not the final machine code that goes into your binary. The linker does some processing during the final link stage. Lines like `FUNCDATA` and `PCDATA` are metadata for the garbage collector which are moved elsewhere when linking. If you're reading the output of `-S`, just ignore `FUNCDATA` and `PCDATA` lines; they're not part of the final binary.

### Discussion

Why did I declare `a` and `b` in `F()` to be constants?

_Exercise_: Experiment with the output of What happens if `a` and `b` are declared as  are variables?  What happens if `a` and `b` are passing into `F()` as parameters?

_Note_: `-gcflags=-S` doesn't prevent the final binary being build in your working directory. If you find that subsequent runs of `go build ...` produce no output, delete the `./max` binary in your working directory.

### Adjusting the level of inlining

Adjusting the _inlining level_ is performed with the `-gcflags=-l` flag. Somewhat confusingly passing a single `-l` will disable inlining, and two or more will enable inlining at more aggressive settings.

- `-gcflags=-l`, inlining disabled
- nothing, regular inlining.
- `-gcflags='-l -l'` inlining level 2, more aggressive, might be faster, may make bigger binaries
- `-gcflags='-l -l -l'` inlining level 3, more aggressive again, binaries definitely bigger, maybe faster again, but might also be buggy.
- `-gcflags=-l=4` (four `-l`s) in Go 1.11 will enable the experimental [_mid stack_ inlining optimisation][5].

### Inlining future (Go 1.12)

We're using Go 1.10 for this workshop. A bunch of work happened under the hood for 1.11, but the rules around what is inlined and when have not changed substantively with one exception

A lot of what we've discussed with respect to leaf functions _may_ change in a future release of Go.

There has been work going on in the background since Go 1.8 to enable, so called, mid-stack inlining.

As a Go programmer, this should not be a cause for alarm, mid-

.link https://github.com/golang/proposal/blob/master/design/19348-midstack-inlining.md Proposal: Mid-stack inlining in the Go compiler

## Dead code elimination

Why is it important that `a` and `b` are constants?

To understand what happened lets look at what the compiler sees once its inlined `Max` into `F`. We can't get this from the compiler easily, but it's straight forward to do it by hand.

Before:
```go
func Max(a, b int) int {
        if a > b {
                return a
        }
        return b
}

func F() {
        const a, b = 100, 20
        if Max(a, b) == b {
                panic(b)
        }
}
```
After:
```go
func F() {
        const a, b = 100, 20
        var result int
        if a > b {
                result = a
        } else {
                result = b
        }
        if result == b {
                panic(b) 
        }
}
```
Because `a` and `b` are constants the compiler can prove at compile time that the branch will never be false; `100` is always greater than `20`. So it can further optimise `F` to 
```go
func F() {
        const a, b = 100, 20
        var result int
        if true {
                result = a
        } else {
                result = b
        }
        if result == b {
                panic(b) 
        }
}
```
Now that the result of the branch is know then then the contents of `result` are also known. This is call _branch elimination_.
```go
func F() {
        const a, b = 100, 20
        const result = a
        if result == b {
                panic(b) 
        }
}
```
Now the branch is eliminated we know that `result` is always equal to `a`, and because `a` was a constant, we know that `result` is a constant. The compiler applies this proof to the second branch
```go
func F() {
        const a, b = 100, 20
        const result = a
        if false {
                panic(b) 
        }
}
```
And using branch elimination again the final form of `F` is reduced to.
```go
func F() {
        const a, b = 100, 20
        const result = a
}
```
And finally just
```go
func F() {
}
```

### Dead code elimination (cont.)

Branch elimination is one of a category of optimisations known as _dead code elimination_. In effect, using static proofs to show that a piece of code is never reachable, colloquially known as _dead_, therefore it need not be compiled, optimised, or emitted in the final binary.

We saw out dead code elimination works together with inlining to reduce the amount of code generated by removing loops and branches that are proven unreachable.

You can take advantage of this to implement expensive debugging, and hide it behind
```
const debug = false 
```
Combined with build tags this can be very useful.

### Further reading

- [Using // +build to switch between debug and release builds][7]
- [How to use conditional compilation with the go build tool][8]

### Compiler flags Exercises

Compiler flags are provided with:
```
go build -gcflags=$FLAGS
```

Investigate the operation of the following compiler functions:

- `-S` prints the (Go flavoured) assembly of the _package_ being compiled.
- `-l` controls the behaviour of the inliner; `-l` disables inlining, `-l -l` increases it (more `-l` 's increases the compiler's appetite for inlining code). Experiment with the difference in compile time, program size, and run time.
- `-m` controls printing of optimisation decision like inlining, escape analysis. `-m`-m` prints more details about what the compiler was thinking.
- `-l`-N` disables all optimisations.

_Note_: If you find that subsequent runs of `go build ...` produce no output, delete the `./max` binary in your working directory.

### Further reading

- [Codegen Inspection by Jaana Burcu Dogan][6]


[0]: https://www.goodreads.com/book/show/112269.Principles_of_Compiler_Design
[1]: https://en.wikipedia.org/wiki/Static_single_assignment_form
[2]: https://golang.org/doc/go1.5#c
[3]: https://blog.golang.org/go1.7
[4]: https://golang.org/ref/spec
[5]: https://github.com/golang/go/issues/19348#issuecomment-393654429
[6]: http://go-talks.appspot.com/github.com/rakyll/talks/gcinspect/talk.slide#1
[7]: http://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release
[8]: http://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool