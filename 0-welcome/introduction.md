# The past and future of Microprocessor performance

This is a workshop about writing high performance code. In other workshops I talk about decoupled design and maintainability, but we’re here today to talk about performance.

I want to start today with a short lecture on how I think about the history of the evolution of computers and why I think writing high performance software is important .

The reality is that software runs on hardware, so to talk about writing high performance code, first  we need to talk about the hardware that runs our code.

## Mechanical Sympathy 

![image-20180818145606919](images/image-20180818145606919.png)

There is a term in popular use at the moment, you’ll hear people like Martin Thompson or Bill Kennedy talk about “mechanical sympathy”.

The name "Mechanical Sympathy" comes from the great racing car driver Jackie Stewart, who was a 3 times world Formula 1 champion. He believed that the best drivers had enough understanding of how a machine worked so they could work in harmony with it. 

To be a great race car driver, you don’t need to be a great mechanic, but you need to have more than a cursory understanding of how a motor car works.

I believe the same is true for us as software engineers. I don’t think any of us in this room will be a professional CPU designer, but that doesn’t mean we can ignore the problems that CPU designers face.

## Six orders of magnitude

There’s a common internet meme that goes something like this;

![jalopnik](images/jalopnik.png)

Of course this is preposterous, but it underscores just how much has changed in the computing industry.

As software authors all of us in this room have benefited from Moore's Law, the doubling of the number of available transistors on a chip every 18 months, for 50 years. **No other industry has experienced a six order of magnitude improvement in their tools in the space of a lifetime**.

But this is all changing.

## Are computers still getting faster?

So the fundamental question is, confronted with statistic like the ones on the slide page, we should ask the question _are computers still getting faster_?

If computers are still getting faster then maybe we don’t need to care about the performance of our code, we just wait a bit and the hardware manufacturers will solve our performance problems.

### Let's look at the data

This is the classic data you’ll find in textbooks like *Computer Architecture, A Quantitative Approach* by John L. Hennessy and David A. Patterson. This graph was taken from the 5th edition

![graph](/Users/dfc/devel/gophercon2018-performance-tuning-workshop/0-welcome/images/graph.png)

Hennessey and Patterson argue that there are three eras

- The first was the 1970’s and 80’s which were really formative years, microprocessors as we know them today didn’t really exist, computers were built from discrete transistors or small scale integrated circuits. Cost, size, and the limitations in understanding material science were the limiting factor.
- From the mid 80s to 2004 the trend line is clear. Computer integer performance improved by 52% each year. Computer power doubled every two years, hence people conflated Moore’s law — the doubling of the number of transistors on a die, with computer performance.
- Then we come to the third era of computer performance. Things slow down. The aggregate rate of change is 22% per year. 

That previous graph only went up to 2012, but fortuntaly in 2012 [Jeff Preshing][0] wrote a [tool to scrape the Spec website and build your own graph][1].

![int_graph-1345](/Users/dfc/devel/gophercon2018-performance-tuning-workshop/0-welcome/images/int_graph-1345.png)

So this is the same graph using Spec data from 1995 til 2017.

To me, rather than the step change we saw in the 2012 data, I’d say that _single core_ performance is approaching a limit. The numbers are slightly better for floating point, but for us in the room doing line of business applications, this is probably not that relevant.

### Yes, computer are still getting faster, slowly

> The first thing to remember about the ending of Moore's law is something Gordon Moore told me. He said "All exponentials come to an end". -- [John Hennessy][2]

This is Hennessy's quote from Google Next 18 and his Turing Award lecture. His contention is yes, CPU performance is still improving. However, single threaded integer performance is still improving around 2-3% per year. At this rate its going to take 20 years of compounding growth to double integer performance. Compare that to the go-go days of the 90's where performance was doubling every two years.

Why is this happening?

## Clock speeds

![stuttering](images/stuttering.png)

This graph from 2015 demonstrates this well. The top line shows the number of transistors on a die. This has continued in a roughly linear trend line since the 1970's. As this is a log-lin graph this linear series represents exponential growth.

However, If we look at the middle line, we see clock speeds have not increased in a decade, we see that cpu speeds stalled around 2004

The bottom graph shows thermal dissipation power; that is electrical power that is turned into heat follows a same pattern--clock speeds and cpu heat dissipation are correlated.

## Heat

Why does a CPU produce heat? It's a solid state device, there are no moving components, so effects like friction are not (directly) relvant here.

The power consumption of a CMOS device, which is what every transistor in this room, on your desk, and in your pocket, is made from, is combination of three factors.

This digram is taken from a great [data sheet produced by TI][7]. In this model the switch in N typed devices is attracted to a positive voltage P type devices are repelled from a positive voltage.

![cmos-inverter](/Users/dfc/devel/gophercon2018-performance-tuning-workshop/0-welcome/images/cmos-inverter.png)

1. Static power. When a transistor is static, that is, not changing its state, there is a small amount of current that leaks through the transistor to ground. The smaller the transistor, the more leakage. So the power consumed by this leakage is the voltage the chip operates multiplied by the leakage current and the number of transistors on the die. Even a minute amount of leakage adds up when you have billions of transistors!
2. Dynamic power. When a transistor transitions from one state to another, it must charge or discharge the various capacitances it is connected to the gate.
3. Crowbar, or short circuit current. We like to think of transistors as digital devices occupying one state or another, off or on, atomically. In reality a transistor is an analog device. As a switch a transistor starts out _mostly_ off, and transitions, or switches, to a state of being _mostly_ on. This transition or switching time is very fast, in modern processors it is in the order of pico seconds, but that still represents a period of time when there is a low resistance path from Vcc to ground. The faster the transistro transitions, its frequency, the more heat is disipated.

It was postulated around 2004 that if we followed the trend line increasing clock speed and shrinking transistor dimensions then within a processor generation the transistor junction would give off as much heat as the core of a nuclear reactor

![pant-GLSVLSI-talk-1338](images/pant-GLSVLSI-talk-1338.png)

Obviously this is was lunacy. The Pentium 4 [marked the end of the line][3] for single core, high frequency, consumer CPUs.

## Dennard Scaling

To understand what happened next we need to look to a paper written in 1974 co-authored by [Robert H. Dennard](https://en.wikipedia.org/wiki/Robert_H._Dennard). Dennard's Scaling law states, roughly, that as transistors get smaller their [power density](https://en.wikipedia.org/wiki/Power_density) stays constant. 

Since the 1970's, Dennard's scaling rule held up quite well. As transistors became smaller, they could operate at lower voltages, thus reducing the amount of power they consumed as the power consumed by a transistor is a function of the square of the voltage. But as the gate lenth of the transistor approaches the width of a silicon atom, the relationship between 

Thus Smaller transistors are aimed at reducing power consumption not improving performance.

## The end of Dennard scaling

Returning to this graph, we see that the reason clock speeds have stalled is because cpu’s exceeded our ability to cool them. So, now we know that CPU feature size reductions are primarily aimed at reducing power consumption. Reducing power consumption doesn't just mean “green”, like recycle, save the planet. The primary goal is to keep power consumption, and thus heat dissipation, below levels that will damage the CPU.

![stuttering](images/stuttering.png)

However by 2006 the ability for chip makers to continue to reduce the size of the transitor started to 

But, there is one part of the graph that is continuing to increase, the number of transistors on a die. The march of cpu features size, more transistors in the same given area, has both positive and negative effects.

Smaller transistors can run at lower voltages, have lower gate capacitence, and switch faster, which helps reduce the amount of dynamic power.

However smaller transistors tend to be less binary towards being off or on, so the static power, the leakage current increases.

Also, as you can see in the insert, the cost per transistor continued to fall until around 5 years ago, and then the cost per transistor started to go back up again.

![gate-length](images/gate-length.png)

Not only is it getting more expensive to create smaller transistors, it’s getting harder. This report from 2016 shows the prediction of what the chip makers believed would occur in 2013; two years later they had missed all their predictions, and while I don’t have an updated version of this report, there are no signs that they are going to be able to reverse this trend. 

It is costing intel, TSMC, AMD, and Samsung billions of dollars because they have to build new fabs, buy all new process tooling. So while the number of transistors per die continues to increase, their unit cost has started to increase.

_note_: Even the term gate length, measured in nano meters, has become ambigious. Various manufacturers measure the size of their transistors in different ways allowing them to demonstate a smaller number than their competitors without perhaps delivering. This is the Non-GAAP Earning reporting model of CPU manufacturers.

So, what are most of these transistors doing?

https://spectrum.ieee.org/semiconductors/devices/transistors-could-stop-shrinking-in-2021

## More cores

![highrescpudies_fullyc_020-1105](images/highrescpudies_fullyc_020-1105.png)

They’re going towards adding more CPU cores. 

CPUs are not getting faster, but they are getting wider with hyper threading and multiple cores. Dual core on mobile parts, quad core on desktop parts, dozens of cores on server parts.

In truth, the core count of a CPU is dominated by heat dissipation. So much so that the clock speed of a CPU is some arbitrary number between 1 and 4 Ghz depending on how hot it is. We'll see this shortly when we talk about benchmarking.

It’s no longer possible to make a single core run twice as fast, but if you add another cores you can provide twice the processing capacity — if the software can support it.

## Amdahl's law

Amdahl's law, named after the Gene Amdahl is a formula which gives the theoretical speedup in latency of the execution of a task at fixed workload that can be expected of a system whose resources are improved.

![AmdahlsLaw](images/AmdahlsLaw.svg)

Amdahl's law tell sus that the maximum speedup of a program is limited by the sequental parts of the program. If you write a program with 95% of its execution able to be run in parallel, even with thousands of processors the maximum speedup in the programs execution is limited to 20x. 

Think about the programs that you work on every day, how much of their execution is parallisable?

## Dynamic Optimisations

With clock speeds stalled and limited returns from throwing extra cores at the problem, where are the speedups coming from? They are coming from architectural improvements in the chips themselves. These are the big five to seven year projects with names like [Nehalem, Sandy Bridge, and Skylake][9]. 

Much of the improvement in performance in the last two decades has come from architectural improvements:

### Out of order execution

Out of order, also known as super scalar, execution is a way of extracting so called _Instruction level parallism_ from the code the CPU is executing. Modern CPUs effectively do SSA at the hardware level to establish data dependencies between operations, and where possible run independant operations in parallel. 

However there is a limit to the amount of parallism inherent in any piece of code. It's also tremendously power hungry as tracking 


### Speculative execution

To avoid the stalls inherent with branches and loads

(super-scalar) -- requires register renaming
speculative execution -- huge power waste

Cliff Click has a [wonderful presentation][10] that argues out of order and speculative execution is most useful for 



vector (SSE) instructions

All are aimed at extracting instruction level parallism -- the ability to transparently 

Dynamic instruction level parallism

- craps out around 6 in flight operations -- because of ahmdals law, and because of the power spent 

All these optimisations lead to the improvements in single threaded performance we've seen, at the cost of huge numbers of transistors and power.

Another place that the transistors are being spent is expensive dynamic optimisations embedded in the CPU itself.



## Modern CPUs are optimised for bulk operations

> Modern processors are a like nitro fuelled funny cars, they excel at the quarter mile. Unfortunately modern programming languages are like Monte Carlo, they are full of twists and turns. -- 

This a quote from David Ungar, an influential computer scientist and the developer of the SELF programming language that I found online in some very old

Thus, modern CPUs are optimised for bulk transfers and bulk operations. At every level, the setup cost of an operation encourages you to work in bulk. Some examples include

- memory is not loaded per byte, but per multiple of cache lines, this is why alignment is becoming less of an issue than it was in earlier computers.
- Vector instructions like MMX and SSE allow a single instruction to execute against multiple items of data concurrently providing your program can be expressed in that form.

## Modern processors are limited by memory latency not memory capacity

If the situation in CPU land wasn't bad enough, the news from the memory side of the house doesn't get much better.

Physical memory attached to a server has increased geometrically. My first computer in the 1980’s had kilobytes of memory. When I went through high school I wrote all my essays on a 386 with 1.8 megabytes of ram. Now its commonplace to find servers with tens or hundreds of gigabytes of ram, and the cloud providers are pushing into the terabytes of ram.

![processor-memory-gap](images/processor-memory-gap.png)

However, the gap between processor speeds and memory access time continues to grow.

![unnamed](images/unnamed.png)

But, in terms of processor cycles lost waiting for memory, physical memory is still as far away as ever because memory has not kept pace with the increases in CPU speed.

So, most modern processors are limited by memory latency not capacity.

## Cache rules everything around me

![memory-latency](images/memory-latency.png)

For decades CPUs have a cache, a piece of small memory

By caches are limited in size because they are physically large on the CPU die, consume a lot of power

## The free lunch is over

In 2005 Herb Sutter, the C++ committee leader, wrote an article entitled [The free lunch is over][5]. In his article Sutter discussed all the points I covered and asserted that programmers could no longer rely on faster hardware to fix slow programs—or slow programming languages.

Now, more than a decade later, there is no doubt that Herb Sutter was right. Memory is slow, caches are too small, CPU clock speeds are going backwards, and the simple world of a single threaded CPU is long gone.

Moore's Law is still in effect, but for all of us in this room, the free lunch is over.

## Conclusion

Ok, so that's all doom and gloom. What's the upside?

There are many presentations online that rehash this material. They all have the same prediction -- computers in the future will not be programmed like they are today. Some argue it'll look more like graphics cards with hundreds of very dumb, very incoherant processors. Others argue that Very Long Instruction Word (VLIW) computers will become predominant. All agree that our current sequential programming languages will not be compatible with these kinds of processors.

My view is that these predictions are right, the outlook for hardware manufacturers saving us at this point is grim. However, there is _enormous_ scope to optimise the programs today we write for the hardware we have today. Rick Hudson spoke at GopherCon 2015 about [re engaging with a "virtuious cycle"][8] of software that works _with_ the hardware we have today, not in spite of it.

Over from 2015 to 2018 looking at the graphs I showed earlier, with at best 5-8% improvement in integer performance and less than that in memory latency, the Go team decreased the garbage collector pause by [two orders of magnitude][11].

So, for best performance on today's hardware in today's world, you need a programming language which:

- Is compiled, not interpreted, because interpreted programming languages operate poorly with CPU branch predictors and speculative execution.
- You need a language which permits efficient code to be written, it needs to be able to talk about bits and bytes, and the length of an integer efficiently, rather than pretend every number is an ideal float.
- You need a language which lets programmers talk about memory effectively, think structs vs java objects, because all that pointer chasing puts pressure on the CPU cache and cache misses burn hundreds of cycles.
- A programming language that scales to multiple cores as  performance of an application is determined by how efficiently it uses its cache and how efficiently it can parallise work over multiple cores.

### Further reading

- [The future of computing: a conversation with John Hennessy][2]  (Google I/O '18)
- Guy from CPP con 2016 / 2017
- [The Future of Microprocessors][6] JuliaCon 2018

[0]: http://preshing.com/20120208/a-look-back-at-single-threaded-cpu-performance/
[1]: https://github.com/preshing/analyze-spec-benchmarks
[2]: https://www.youtube.com/watch?v=Azt8Nc-mtKM
[3]: https://arstechnica.com/uncategorized/2004/10/4311-2/
[4]: https://www.youtube.com/watch?v=LgLNyMAi-0I&list=PLFls3Q5bBInj_FfNLrV7gGdVtikeGoUc9 Mill Computing
[5]: http://www.gotw.ca/publications/concurrency-ddj.htm
[6]: https://www.youtube.com/watch?v=zX4ZNfvw1cw
[7]: http://www.ti.com/lit/an/scaa035b/scaa035b.pdf
[8]: https://talks.golang.org/2015/go-gc.pdf
[9]: https://en.wikipedia.org/wiki/List_of_Intel_CPU_microarchitectures#Pentium_4_/_Core_Lines
[10]: https://www.youtube.com/watch?v=OFgxAFdxYAQ
[11]: https://blog.golang.org/ismmkeynote


