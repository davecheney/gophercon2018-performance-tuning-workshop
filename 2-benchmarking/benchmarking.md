# Benchmarking

This section focuses on how to construct useful benchmarks using the Go testing framework, and gives practical tips for avoiding the pitfalls.

## Benchmarking ground rules

Before you benchmark, you must have a stable environment to get repeatable results.

- The machine must be idle—don't profile on shared hardware, don't browse the web while waiting for a long benchmark to run.
- Watch out for power saving and thermal scaling.
- Avoid virtual machines and shared cloud hosting; they are too noisy for consistent measurements.

If you can afford it, buy dedicated performance test hardware. Rack it, disable all the power management and thermal scaling and never update the software on those machines.

For everyone else, have a before and after sample and run them multiple times to get consistent results.

## Using the testing package for benchmarking
The `testing` package has built in support for writing benchmarks. If we have a simple function like this
```go
// Fib computes the n'th number in the Fibonacci series.
func Fib(n int) int {
        switch n {
        case 0:
                return 0
        case 1:
                return 1
        default:
                return Fib(n-1) + Fib(n-2)
        }
}
```
The we can use the `testing` package to write a _benchmark_ for the function using this form. The benchmark function lives alongside your tests in a `_test.go` file.
```go
func BenchmarkFib20(b *testing.B) {
        for n := 0; n < b.N; n++ {
                Fib(20) // run the Fib function b.N times
        }
}
```
Benchmarks are similar to tests. The only real difference is they take a `*testing.B` rather than a `*testing.T`. Both of these types implement the `testing.TB` interface which provides crowd favorites like `Errorf()`, `Fatalf()`, and `FailNow()`

### Running a package's benchmarks

As benchmarks use the `testing` package they are executed via the `go test` subcommand. However, by default when you invoke `go test` the benchmarks are excluded. 

To explicitly run benchmarks in a package use the `-bench` flag. `-bench` takes a regular expression that matches the names of the benchmarks you want to run, so the most common way to invoke all benchmarks in a package is `-bench=.`. Here is an example:
```
% go test -bench=. ./examples/fib/
goos: darwin
goarch: amd64
BenchmarkFib20-8           30000             44514 ns/op
PASS
ok      _/Users/dfc/devel/gophercon2018-performance-tuning-workshop/2-benchmarking/examples/fib 1.795s
```
_Note_: `go test` will also run all the tests in a package before matching benchmarks, so if you have a lot of tests in a package, or they take a long time to run, you can exclude them by providing `go test`'s `-run` flag with a regex that matches nothing; ie. `go test -run=^$`.

### How benchmarks work

Each benchmark function is called with a valid of `b.N`, this is the number of iterations the benchmark should run for.

`b.N` starts at 1, if the benchmark function completes in under 1 second (the default), then `b.N` is increased and the benchmark function run again.

`b.N` increases in the approximate sequence; 1, 2, 3, 5, 10, 20, 30, 50, 100, and so on. The benchmark framework tries to be smart and if it sees small values of `b.N` are completing quickly, it will start to increase the sequence faster.

Looking at the example above, `BenchmarkFib20-8` found that around 30,000 iterations of the loop took just over a second. From there the benchmark framework computed that 

_Note_: The `-8` suffix relates to the value of `GOMAXPROCS` that was used to run this test. This number, like `GOMAXPROCS`, defaults to the number of CPUs visible to the Go process on startup. You can change this value with the `-cpu` flag which takes a list of values to run the benchmark with.
```
% go test -bench=. -cpu=1,2,4 ./examples/fib/
goos: darwin
goarch: amd64
BenchmarkFib20             30000             44644 ns/op
BenchmarkFib20-2           30000             44504 ns/op
BenchmarkFib20-4           30000             44848 ns/op
PASS
```
This shows running the benchmark with 1, 2, and 4 cores. In this case the flag has no option because this benchmark is entirely sequential. 

###  Improving benchmark accuracy

The `fib` function is a slightly contrived example -- unless your writing a TechPower web server benchmark, its unlikely your business is going to be gated on how quickly you can compute the 20th number in the Fibonaci sequence. But, the benchmark does fulfil what I consider to be a valid benchmark. 

Specifically you want your benchmark to run for several to several thousand iterations so you get a good average per operation. If your benchmark runs for only 100's or 10's of iterations, the average of those runs may not be stable. 

To increase the number of iterations, the benchmark time can be increased with the `-benchtime` flag. For example:
```
% go test -bench=. -benchtime=10s ./examples/fib/
goos: darwin
goarch: amd64
BenchmarkFib20-8          300000             44616 ns/op
```
Ran the same benchmark until it reached a value of `b.N` that took longer than 10 seconds to return. As we're running for 10x longer, the total number of iterations is 10x larger. The result hasn't changed much, which is what we expected.

If you have a benchmark which runs for millons or billions of iterations resulting in a time per operation in the micro or nano second range, you may find that your benchmark numbers are unstable because  thermal scaling, memory locality, background processing, gc activity, etc.

For times measured in 10 or single digit nanoseconds per operation the relativistic effects of instruction reordering and code alignment will have an impact on your benchmark times.

To address this run benchmarks multiple times with the `-count` flag:
```
% go test -bench=Fib1 -count=10 ./examples/fib/
goos: darwin
goarch: amd64
BenchmarkFib1-8         2000000000               1.99 ns/op
BenchmarkFib1-8         1000000000               1.95 ns/op
BenchmarkFib1-8         2000000000               1.99 ns/op
BenchmarkFib1-8         2000000000               1.97 ns/op
BenchmarkFib1-8         2000000000               1.99 ns/op
BenchmarkFib1-8         2000000000               1.96 ns/op
BenchmarkFib1-8         2000000000               1.99 ns/op
BenchmarkFib1-8         2000000000               2.01 ns/op
BenchmarkFib1-8         2000000000               1.99 ns/op
BenchmarkFib1-8         1000000000               2.00 ns/op
```
A benchmark of `Fib(1)` takes around 2 nano seconds with a variance of +/- 2%. 

_Tip:_ If you find that the defaults that `go test` applies need to be tweaked for a particular package, I suggest codifying those settings in a `Makefile` so everyone who wants to run your benchmarks can do so with the same settings.

## Benchstat

In the previous section I suggested running benchmarks more than once to get more data to average. This is good advice for any benchmark because of the effects of power management, background processes, and thermal management that I mentioned at the start of the chapter.

I'm going to introduce a tool by Russ Cox called [benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat)
```
% go get golang.org/x/perf/cmd/benchstat
```
Benchstat can take a set of benchmark runs and tell you how stable they are. Here is an example of `Fib(20)` on battery power.
```
% go test -bench=Fib20 -count=10 ./examples/fib/ | tee old.txt
goos: darwin
goarch: amd64
BenchmarkFib20-8           30000             46295 ns/op
BenchmarkFib20-8           30000             41589 ns/op
BenchmarkFib20-8           30000             42204 ns/op
BenchmarkFib20-8           30000             43923 ns/op
BenchmarkFib20-8           30000             44339 ns/op
BenchmarkFib20-8           30000             45340 ns/op
BenchmarkFib20-8           30000             45754 ns/op
BenchmarkFib20-8           30000             45373 ns/op
BenchmarkFib20-8           30000             44283 ns/op
BenchmarkFib20-8           30000             43812 ns/op
PASS
ok      _/Users/dfc/devel/gophercon2018-performance-tuning-workshop/2-benchmarking/examples/fib 17.865s
% benchstat old.txt
name     time/op
Fib20-8  44.3µs ± 6%
```
`benchstat` tells us the mean is 44.3 microseconds with a +/- 6% variation across the samples. This is not unexpected on battery power. The first run is the slowest of all because the operating system had the CPU clocked down to save power. The next two runs are the fastest, because the operating system as decided that this isn't a transient spike of work and it has boosted up the clock speed to get through the work as quick as possible in the hope of being able to go back to sleep, then the remaining runs are the operating system and the bios trading power consumption for heat production.

## Comparing benchmarks with benchstat

Determining the performance delta between two sets of benchmarks can be tedious and error prone. Benchstat can help us with this. 

_Tip_: Saving the output from a benchmark run is useful, but you can also save the _binary_ that produced it. This lets you rerun benchmark previous iterations. To do this, use the `-c` flag to save the test binary; I often rename this binary from `.test` to `.golden`.

```
% go test -c
% mv fib.test fib.golden 
```
## Improve `Fib`
The previous `Fib` fuction had hard coded values for the 0th and 1st numbers in the fibonaci series. After that the code calls itself recursively. We'll talk about the cost of recursion later today, but for the moment, assume it has a cost, especially as our algorithm uses exponential time.

As simple fix to this would be to hard code another number from the fibonacci series, reducing the depth of each recusive call by one.
```go
// Fib computes the n'th number in the Fibonacci series.
func Fib(n int) int {
        switch n {
        case 0:
                return 0
        case 1:
                return 1
        case 2:
                return 2
        default:
                return Fib(n-1) + Fib(n-2)
        }
}
```
_Tip_: This file also includes a comprehensive test for `Fib`. Don't try to improve your benchmarks without a test that verifies the current behaviour.

To compare our new version, we compile a new test binary and benchmark both of them and use `benchstat` to compare the outputs.
```
% go test -c
% ./fib.golden -test.bench=. -test.count=10 > old.txt
% ./fib.test -test.bench=. -test.count=10 > new.txt
% benchstat old.txt new.txt
name     old time/op  new time/op  delta
Fib20-8  44.3µs ± 6%  25.6µs ± 2%  -42.31%  (p=0.000 n=10+10)
```

There are three things to check when comparing benchmarks

- The variance ± in the old and new times. 1-2% is good, 3-5% is ok, greater than 5% and some of your samples will be considered unreliable. Be careful when comparing benchmarks where one side has a high variance, you may not be seeing an improvement.
- p value. p values lower than 0.05 are good, greater than 0.05 means the benchmark may not be statistically significant.
- Missing samples. benchstat will report how many of the old and new samples it considered to be valid, sometimes you may find only, say, 9 reported, even though you did `-count=10`. A 10% or lower rejection rate is ok, higher than 10% may indicate your setup is unstable and you may be comparing too few samples.

## Avoiding benchmarking start up costs

Sometimes your benchmark has a once per run setup cost. `b.ResetTimer()` will can be used to ignore the time accrued in setup.
```go
func BenchmarkExpensive(b *testing.B) {
        boringAndExpensiveSetup()
        b.ResetTimer() // HL
        for n := 0; n < b.N; n++ {
                // function under test
        }
}
```
If you have some expensive setup logic _per loop_ iteration, use `b.StopTimer()` and `b.StartTimer()` to pause the benchmark timer.
```go
func BenchmarkComplicated(b *testing.B) {
        for n := 0; n < b.N; n++ {
                b.StopTimer() // HL
                complicatedSetup()
                b.StartTimer() // HL
                // function under test
        }
}
```
## Benchmarking allocations

Allocation count and size is strongly correlated with benchmark time. You can tell the `testing` framework to record the number of allocations made by code under test.
```go
func BenchmarkRead(b *testing.B) {
        b.ReportAllocs() // HL
        for n := 0; n < b.N; n++ {
                // function under test
        }
}
```

DEMO: `go`test`-run=^$`-bench=.`bufio`

_Note:_ you can also use the `go`test`-benchmem` flag to do the same for _all_ benchmarks.

DEMO: `go`test`-run=^$`-bench=.`-benchmem`bufio`

### Watch out for compiler optimisations

This example comes from [issue 14813](https://github.com/golang/go/issues/14813#issue-140603392).
```go
const m1 = 0x5555555555555555
const m2 = 0x3333333333333333
const m4 = 0x0f0f0f0f0f0f0f0f
const h01 = 0x0101010101010101

func popcnt(x uint64) uint64 {
        x -= (x >> 1) & m1
        x = (x & m2) + ((x >> 2) & m2)
        x = (x + (x >> 4)) & m4
        return (x * h01) >> 56
}

func BenchmarkPopcnt(b *testing.B) {
        for i := 0; i < b.N; i++ {
                popcnt(uint64(i))
        }
}
```
How fast do you think this function will benchmark? Let's find out.
```
% go test -bench=. ./examples/popcnt/
goos: darwin
goarch: amd64
BenchmarkPopcnt-8       2000000000               0.30 ns/op
PASS
```
0.3 of a nano second; that's basically one clock cycle. Even assuming that the CPU may have a few instructions in flight per clock tick, this number seems unreasonably low. What happened?

To understand what happened, we have to look at the function under benchmake, `popcnt`. `popcnt` is a leaf function -- it does not call any other functions --  so the compiler can inline it.

Because the function is inlined, the compiler now can see it has no side effects. `popcnt` does not affect the state of any global variable. Thus, the call is eliminated. This is what the compiler sees:
```go
func BenchmarkPopcnt(b *testing.B) {
        for i := 0; i < b.N; i++ {
                // optimised away
        }
}
```
On all versions of the Go compiler, the loop is still generated. But Intel CPUs are really good at optimising loops, especially empty ones. 

### Optimisations are good

The thing to take away is the same optimisations that make real code fast, by removing unnecessary computation, are the same ones that remove benchmarks that have no observable side effects.

This issue was is only going to get more common as the Go compiler improves.

DEMO: show how to fix popcnt

## Benchmark mistakes

The `for` loop is crucial to the operation of the benchmark.

Here are two incorrect benchmarks, can you explain what is wrong with them?

.code examples/benchfib/wrong_test.go /START OMIT/,/END OMIT/

## Profiling benchmarks

The `testing` package has built in support for generating CPU, memory, and block profiles.

- `-cpuprofile=$FILE` writes a CPU profile to `$FILE`.
- `-memprofile=$FILE`, writes a memory profile to `$FILE`, `-memprofilerate=N` adjusts the profile rate to `1/N`.
- `-blockprofile=$FILE`, writes a block profile to `$FILE`.

Using any of these flags also preserves the binary.

    % go test -run=XXX -bench=. -cpuprofile=c.p bytes
    % go tool pprof bytes.test c.p

_Note:_ use `-run=XXX` to disable tests, you only want to profile benchmarks. You can also use `-run=^$` to accomplish the same thing.

## Discussion

Are there any questions?

Perhaps it is time for a break.