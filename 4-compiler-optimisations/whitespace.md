# Whitespace exercise

Protocols like [HTTP][0] and [JSON][1] define whitespace as a small set of characters which are either noise (whitespace) or used to split words. For example `encoding/json/scanner.go` defines an [`isSpace`][2] function like this:

```go
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
```

While small, and straightforward, `isSpace` is called at least once for every byte of JSON input. If we can improve the speed of `isSpace` those savings _should_ for through to all users of the JSON decoder.

## Establishing a baseline

First we need to figure out how expensive it is to call `isSpace`. Let's write a quick benchmark which we can use to get a baseline.

```go
var B bool

func BenchmarkIsSpaceTrue(b *testing.B) {
        benchmarkIsSpace(b, '\t')
}

func BenchmarkIsSpaceFalse(b *testing.B) {
        benchmarkIsSpace(b, 'a')
}

func benchmarkIsSpace(b *testing.B, c byte) {
        var r bool
        for i := 0; i < b.N; i++ {
                r = isSpace(c)
        }
        B = r
}
```

### Baseline results
Here are the baseline results for my machine; yours, and mine, will vary on the day.
```
% go test -count=10 -bench=. >  old.txt
% benchstat old.txt
name                 time/op
IsSpaceTrue-8   0.59ns ± 5%
IsSpaceFalse-8  1.89ns ± 2%
```
## Switch
Looking at the body of `isSpace` its really asking the question, _if `c` a space, tab, linefeed, or carage return, return true, otherwise return false_. There is a construct in Go which matches these semantics quite well, `switch`. This is what `isSpace` looks like rewritten as a `switch` statement.

```
func isSpace(c byte) bool {
        switch c {
        case ' ', '\t', '\n', '\r':
                return true
        default:
                return false
        }
}
```
Let's benchmark this new version and see if there is an improvement
```
% go test -count=10 -bench=. > tee new.txt
% benchstat old.txt new.txt
name                 old time/op  new time/op  delta
IsSpaceTrue-8   0.59ns ± 5%  0.59ns ± 4%     ~     (p=0.429 n=9+9)
IsSpaceFalse-8  1.89ns ± 2%  0.88ns ± 3%  -53.23%  (p=0.000 n=10+10)
```
This is promising, we've reducted the time to assert a character is not whitespace by 53%. 

## Bytes are just numbers
One interesting property the switch statement has that `if` doesn't necessarily have is the compiler can recognise that because we're making a choice between a set of _related_ options, those options can be considered to be a _sparse range_. 

Specifically `\\t` has the ASCII value 0x09, `\\n` is 0x0a, and so on up to space which is 0x20. So the compiler is able to recognise that the values for which `isSpace` could be true lie within the range 0x09 to 0x20--if `c` is outside that range then it is by definition, false.

This is what the compiler is doing. You can check this by adding `-gcflags=-S` to your `go test` command and searching the output for `isSpace`. 

This gives us a hint for how we can speed up `isSpace`. We can take advantage of _short circuit evaluation_ and test if the character is greater than space and come up with this monstrosity
```go
func isSpace(ch byte) bool {
        return !(ch > ' '+1) && (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r')
}
```
Benchmarking this version we find that the `true` case has regressed, but the `false` case has improved significantly.
```
% benchstat old.txt new.txt
name            old time/op  new time/op  delta
IsSpaceTrue-8   0.59ns ± 5%  0.87ns ± 3%  +47.92%  (p=0.000 n=9+9)
IsSpaceFalse-8  1.89ns ± 2%  0.58ns ± 3%  -69.15%  (p=0.000 n=10+10)
```

_Note_: There is a way to write `isSpace` in a way that is less compliant, but will pass most tests, and faster. Can you figure out what it it could be?

### Mapping
Steping back, what we have is a map; some values of `c` map to `true`, others map to `false`. What if we used a map?
```go
var wsm = map[byte]bool{
        ' ':  true,
        '\t': true,
        '\n': true,
        '\r': true,
}

func isSpace(ch byte) bool { return wsm[ch] }
```
How did that turn out?
```
% benchstat old.txt new.txt
name            old time/op  new time/op   delta
IsSpaceTrue-8   0.59ns ± 5%  21.31ns ± 4%  +3518.68%  (p=0.000 n=9+10)
IsSpaceFalse-8  1.89ns ± 2%  19.11ns ± 3%   +911.11%  (p=0.000 n=10+10)
```
Not good, the overhead of map lookups is too high. But, the idea of trading space for time perhaps has some merit, if we can find a data structure with a cheep enough lookup cost.

### Arrays
Fortunately such a structure exists, the humble array. 
```go
var wsp = [256]bool{
        ' ':  true,
        '\t': true,
        '\n': true,
        '\r': true,
}

func isSpace(c byte) bool { return wsp[c] }
```
Not only is the cost of an array lookup O(1), accessing the array is a question of adding the value of `c` to the base address of `wsp` in memory. Because we've defined the size of `wsp` to be the number of possible values a `byte` can represent, the compiler can elide the bounds check because all values of `c` are within the bounds of the array.

Let's look at the results.
```
% benchstat old.txt new.txt
name            old time/op  new time/op  delta
IsSpaceTrue-8   0.59ns ± 5%  0.38ns ±17%  -34.79%  (p=0.000 n=9+10)
IsSpaceFalse-8  1.89ns ± 2%  0.37ns ± 2%  -80.26%  (p=0.000 n=10+10)
```
Excellent. For a modest tradeoff in space -- 256 bytes -- we've improved the speed of `isSpace` for both cases.

## Real world word counting
0.38ns is effectively one clock at 3Ghz, this is an unrealistic number. Likely the the code and data being accessed lives within the L1 cache, and the result of the benchmark always being the same. This is something that branch predictors in the CPU really like.

In real world scenarios, the input to `isSpace` would vary over the letter frequency of the input text. We can write a benchmark that consumes a large input text.
```go
var Result int

func BenchmarkCountWhitespacePrideAndPrejudice(b *testing.B) {
        buf := setup(b, "testdata/pnp.txt")
        var r int
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
                r = 0
                for _, c := range buf {
                        if isSpace(c) {
                                r++
                        }
                }
                Result = r
        }
}
```
Run this benchmark against the various methods we've seen and see if the improvements demonstrated in our synthetic benchmarks hold true for a larger set of input data.

## Taking this further

In the example we are lucky because the input to `isSpace` is a byte. However, in Go, we talk in unicode `rune`s. 

Rewrite `isSpace` to work on a `rune` type and re-evaluate the performance and tradeoffs of each method -- or invent your own.

## Homework

The `isSpace` pattern appears several times in the Go standard library and many more times in the wider corpus of Go code--grep over your `GOPATH` for `isWhitespace` and `isSpace`.

If you're at a loss for something to do on Community Day and are looking to get your Github streak up, 

A word of warning, I applied these techniques to `encoding/json` and while it had a 7% improvement in on decodin benchmark, it also caused regressions of 2-5% in other benchmarks. Make sure you benchmark the macro, as well as the micro, effects of your change.

[0]: https://tools.ietf.org/html/rfc7230#section-3.2.3
[ 1]: https://tools.ietf.org/html/rfc7159#section-2
[ 2]: https://github.com/golang/go/blob/5d11838654464c42de48958ff163360da38ab850/src/encoding/json/scanner.go#L160