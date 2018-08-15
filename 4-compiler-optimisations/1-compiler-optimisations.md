# Compiler optimisations

This section gives a brief background on three important optimisations that the Go compiler performs.

- Escape analysis
- Inlining
- Dead code elimination

These are all handled in the front end of the compiler, while the code is still in its AST form; then the code is passed to the SSA compiler for further optimisation. 

## History of the Go compiler

The Go compiler started as a fork of the Plan9 compiler tool chain ~ 2007.

In Go 1.5 the compiler was mechanically translated into Go.

In Go 1.7 the 'back end' of the compiler was rewritten from the Plan9 style code generation to a new form using [SSA][1] techniques.

## Escape analysis

A compliant Go implementation _could_ store every allocation on the heap, but that would put a lot of pressure on the gc.

However, the stack exists as a cheap place to store local variables; there is no need to garbage collect things on the stack.

In some languages, like C and C++, stack/heap allocation is manual, and a common cause of memory corruption bugs.

In Go, the compiler automatically moves local values to the heap if they live beyond the lifetime of the function call. It is said that the value  _escapes_ to the heap.

But the compiler can also do the opposite, it can find things which would be assumed to be allocated on the heap, `new`, `make`, etc, and move them to stack.

## Escape analysis (example)

Sum adds the ints between 1 and 100 and returns the result.

.code examples/esc/sum.go /START OMIT/,/END OMIT/
.caption examples/esc/sum.go

Because the numbers slice is only referenced inside Sum, the compiler will arrange to store the 100 integers for that slice on the stack, rather than the heap. There is no need to garbage collect `numbers`, it is automatically free'd when `Sum` returns.

## Investigating escape analysis

Prove it!

To print the compilers escape analysis decisions, use the `-m` flag.
```
% go build -gcflags=-m examples/esc/sum.go
# command-line-arguments
examples/esc/sum.go:10: Sum make([]int, 100) does not escape
examples/esc/sum.go:25: Sum() escapes to heap
examples/esc/sum.go:25: main ... argument does not escape
```
Line 10 shows the compiler has correctly deduced that the result of `make([]int, 100)` does not escape to the heap.

We'll come back to line 25 soon.

## Escape analysis (example)

This example is a little contrived.

.code examples/esc/center.go /START OMIT/,/END OMIT/

`NewPoint` creates a new `*Point` value, `p`. We pass `p` to the `Center` function which moves the point to a position in the center of the screen. Finally we print the values of `p.X` and `p.Y`.

## Escape analysis (example)
```
% go build -gcflags=-m examples/esc/center.go 
# command-line-arguments
examples/esc/center.go:12: can inline Center
examples/esc/center.go:19: inlining call to Center
examples/esc/center.go:12: Center p does not escape
examples/esc/center.go:20: p.X escapes to heap
examples/esc/center.go:20: p.Y escapes to heap
examples/esc/center.go:18: NewPoint new(Point) does not escape
examples/esc/center.go:20: NewPoint ... argument does not escape
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


[1]: https://en.wikipedia.org/wiki/Static_single_assignment_form
