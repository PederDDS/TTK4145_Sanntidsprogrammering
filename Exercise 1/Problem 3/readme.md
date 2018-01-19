# Reasons for concurrency and parallelism


To complete this exercise you will have to use git. Create one or several commits that adds answers to the following questions and push it to your groups repository to complete the task.

When answering the questions, remember to use all the resources at your disposal. Asking the internet isn't a form of "cheating", it's a way of learning.

 ### What is concurrency? What is parallelism? What's the difference?
  > Concurrency is the ability of different units of a program/an algorithm to run in parallell and without a specified order. Having a concurrent program means that several units can be executed at the same time.
 > Parallelism means that several tasks are being performed simultaneously. This mainly applies to 
 > We don't need to have concurrency to have parallelism. We can still have parallelism at instruction- and bit-level without writing concurrent code.
 
 ### Why have machines become increasingly multicore in the past decade?
 > Multi core processors can run several parallell operations, which increases speed
 
 ### What kinds of problems motivates the need for concurrent execution?
 (Or phrased differently: What problems do concurrency help in solving?)
 > It solves mutual exclusion(?)
 > 
 
 ### Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
 (Come back to this after you have worked on part 4 of this exercise)
 > *Your answer here*
 
 ### What are the differences between processes, threads, green threads, and coroutines?
 > Process: OS-managed program? Is concurrent
 > Thread: OS-managed, 
 > Green thread: user controlled threads? 
 > Coroutines: user managed co-operatively multitasking threads?
 
 ### Which one of these do `pthread_create()` (C/POSIX), `threading.Thread()` (Python), `go` (Go) create?
 > 1. creates thread
 > 2. ???
 > 3. ???
 
 ### How does pythons Global Interpreter Lock (GIL) influence the way a python Thread behaves?
 > *Your answer here*
 
 ### With this in mind: What is the workaround for the GIL (Hint: it's another module)?
 > *Your answer here*
 
 ### What does `func GOMAXPROCS(n int) int` change? 
 > GOMAXPROCS sets the maximum number of CPUs that can be executing simultaneously and returns the previous setting. If n < 1, it does not change the current setting
