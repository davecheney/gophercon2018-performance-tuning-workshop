# Conclusion

To summarise the advice from today

## Keep it simple

Start with the simplest possible code.

_Measure_. Profile your code to identify the bottlenecks, _do_not_guess_.

If performance is good, _stop_. You don't need to optimise everything, only the hottest parts of your code.

## Not every part of your application needs to be fast

For most applications where performance is a concern, the 80/20 rule applies. 80% of the time will be spent in 20% of the code--Knuth argues the number is closer to 3%.

As your application grows, or your traffic pattern evolves, these performance hot spots will change.

Don't leave complex code that is not performance critical, rewrite it with simpler operations if the bottleneck moves elsewhere.

## The Go compiler is optimised for simple code

Always write the simplest code you can, the compiler is optimised for _normal_ code. I'm not going to say _Idiomatic_ because I don't like how we use that word when discussing Go. So I'll just say simple, not clever, code.

Shorter code is faster code; Go is not C++, do not expect the compiler to unravel complicated abstractions.

Shorter code is _smaller_ code; which is important for the CPU's cache.

## Be on the lookout for quadratic operations

> If a program is too slow, it must have a loop -- Ken Thompson

Most programs perform well with small amounts of data.

I'm a big fan of metrics, rather than testing with 

Limit the communication and points of co-ordination between the parts of your program to ride Ahmdawls law.

## Performance rule of thumb:

Network/disk io >> allocations >> function calls

If your program is dominated by network or disk access, don’t bother optimising allocations. 

If your program is allocation bound, don’t bother optimising functions for inlining, loop unrolling, etc.

Pay very close attention to allocations, avoid unnecessary allocation where possible.

## Don't trade performance for reliability

> I can make things very fast if they don't have to be correct. -- Russ Cox

Finally, don't trade performance for reliability.

> Readable means reliable -- Rob Pike

Performance and reliability are equally important. I see little value in making a very fast server that panics, deadlocks or OOMs on a regular basis.
