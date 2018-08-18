# GopherCon 2018 Performance Tuning Workshop

## Instructors

- David Cheney <dave@cheney.net>
- Francesc Campoy <campoy@golang.org>



## Overview

The goal for this workshop 

We will work upwards from the basics of profiling and benchmarking. 


The general theme is to start from the smallest pieces, profiling a small piece of code, the write a benchmark to quantify the changes, take a small detour into escape analysis, inlining, as they are related to microbenchmarking. Then step out and talk about the execution tracer, including tracing running applications. The last two sections are really a grab back of advice, and tips.


## Schedule (approximate)

| Start | Description |
| --- | --- |
| 09:00 | [Welcome and introduction][1] |
| 09:30 | [Profiling (dfc)][2] |
| 10:15 | Benchmarking (francesc) |
| 11:00 | [Compiler optimisation (dfc)][4]|
| 12:00 | Lunch |
| 13:30 | Execution tracer (francesc) |
| 14:30 | Memory and Garbage collection (francesc) |
| 15:30 | [Tips and tricks (dfc)][6] |
| 16:00 | Exercises |
| 16:45 | [Final Questions and conclusion][8] |
| 17:00 | Close |


## License and Materials

This presentation is licensed under the [Creative Commons Attribution-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-sa/4.0/) licence.

You are encouraged to remix, transform, or build upon the material, providing you give appropriate credit and distribute your contributions under the same license.

[1]: 1-welcome/1-welcome.md
[2]: 2-profiling/1-profiling.md
[4]: 4-compiler-optimisation/1-compiler-optimisation.md
[6]: 6-tips-and-tricks/1-tips-and-tricks.md
[8]: 8-conclusion/1-conclusion.md