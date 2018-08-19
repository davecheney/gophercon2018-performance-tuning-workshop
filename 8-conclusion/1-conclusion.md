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

## Be on the lookout for quadratic operations

> If a program is too slow, it must have a loop -- Ken Thompson

When the number of elemen



Limit the communication and points of co-ordination between the parts of your program to ride Ahmdawls law

Performance rule of thumb:

Network/disk io >> allocations >> function calls

If your program is dominated by network or disk access, don’t bother optimising allocations. 

If your program is allocation bound, don’t bother optimising functions for inlining, loop unrolling, etc.

## Don't trade performance for reliability

> I can make things very fast if they don't have to be correct. -- Russ Cox

> Readable means reliable -- Rob Pike

Performance and reliability are equally important.

I see little value in making a very fast server that panics, deadlocks or OOMs on a regular basis.
