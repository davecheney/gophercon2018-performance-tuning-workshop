# Welcome

Hello and welcome! :tada:

The goal for this workshop is to give you the tools you need to diagnose performance problems in your Go applications and fix them.

Through the day we'll work from the small -- learning how to write benchmarks, then profiling a small piece of code. Then step out and talk about the execution tracer, the garbage collector  and tracing running applications. The remainder of the day will be a chance for you to ask questions, experiement with your own code.

## Prerequisites

The are several software downloads you will need today. 

### Laptop, power supplies, etc.

The workshop material targets Go 1.10.

[**Download Go 1.10.3**][0]

_Note_: Although there is a strong chance Go 1.11 will have shipped by Gophercon, we're sticking to Go 1.10. If you want to use Go 1.11, you'll have to do the translations in your head.

### Graphviz

The section on pprof requires the `dot` program which ships with the `graphviz` suite of tools.

- Linux: `[sudo] apt-get install graphviz`
- OSX:
  - MacPorts: `sudo port install graphviz`
  - Homebrew: `brew install graphviz`
- [Windows][1] (untested) 

### Google Chrome

The section on the execution tracer requires Google Chrome. It will not work with Safari, Edge, Firefox, or IE 4.01. Sorry. 

[**Download Google Chrome**][2]

### Your own code to profile and optimise

The final section of the day will be an open session where you can experiment with the tools you've learnt.

## One more thing ...

This isn't a lecture, it's a conversation. We'll have lots of breaks to ask questions.

If you don't understand something, or think what you're hearing is incorrect, please ask.

[0]: https://golang.org/dl/#go1.10.3
[1]: https://graphviz.gitlab.io/download/#Windows
[2]: https://www.google.com/chrome/