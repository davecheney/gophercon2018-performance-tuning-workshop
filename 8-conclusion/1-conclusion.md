# Conclusion

Start with the simplest possible code.

_Measure_. Profile your code to identify the bottlenecks, _do_not_guess_.

If performance is good, _stop_. You don't need to optimise everything, only the hottest parts of your code.

As your application grows, or your traffic pattern evolves, the performance hot spots will change.

Don't leave complex code that is not performance critical, rewrite it with simpler operations if the bottleneck moves elsewhere.

Always write the simplest code you can, the compiler is optimised for _normal_ code.

Shorter code is faster code; Go is not C++, do not expect the compiler to unravel complicated abstractions.

Shorter code is _smaller_ code; which is important for the CPU's cache.

Pay very close attention to allocations, avoid unnecessary allocation where possible.

Don't trade performance for reliability

"I can make things very fast if they don't have to be correct."
.caption Russ Cox

"Readable means reliable"
.caption Rob Pike

Performance and reliability are equally important.

I see little value in making a very fast server that panics, deadlocks or OOMs on a regular basis.

Don't trade performance for reliability


