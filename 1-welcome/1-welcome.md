# Welcome

Hello and welcome!

The general theme of this workshop to start from the smallest pieces--profiling a small piece of code, the write a benchmark to quantify the changes--take a small detour into escape analysis, inlining, as they are related to microbenchmarking. Then step out and talk about the execution tracer, including tracing running applications. The last two sections are really a grab back of advice, and tips.

## Prerequisites

Please prepare and bring the following with you on the day. The are several software downloads you will need. Please make sure you have them installed _before_ the workshop as you are large in number and we are small in number.

### Laptop, power supplies, etc.

The workshop material targets Go 1.10, if you have an older version of Go on your laptop, please upgrade to Go 1.10.3 before arriving.

_Note_: Although there is a strong chance Go 1.11 will have shipped by Gophercon, we're sticking to Go 1.10. If you want to use Go 1.11, you'll have to do the translations in your head.

WiFi will be provided. Expect it to be the usual convention center fair and make sure you read the rest of this page _before_ you arrive.

### Graphviz

The section on pprof requires the `dot` program which ships with the `graphviz` suite of tools. Please make sure you have it installed _before_ arriving.

- Linux: `[sudo] apt-get install graphviz`
- OSX:
  - MacPorts: `sudo port install graphviz`
  - Homebrew: `brew install graphviz`
- Windows: See [https://graphviz.gitlab.io/download/#Windows] (untested) 

### Google Chrome

The section on the execution tracer requires Google Chrome. It will not work with Safari, Edge, Firefox, or IE 4.01. Sorry. 

[https://www.google.com/chrome/]

We'll be dedicating a large portion of the workshop to the execution tracer so please make sure you have Chrome installed _before_ setting off for Denver. Don't get caught short by needing to raise a ticket with your IT Helpdesk on the day.

### Your own code to profile and optimise

The final section of the day will be an open 

## One more thing ...

This isn't a lecture, it's a conversation. We'll have lots of breaks to ask questions.

If you don't understand something, or think what you're hearing is incorrect, please ask.

Okay, [let's get started](../1-profiling/1-profiling.md)
