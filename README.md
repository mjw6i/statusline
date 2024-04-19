# What it does
It functions similarly to a taskbar (in an environment where there is no taskbar).
- display the day of the week, date and time
- display current volume if recently changed
- show a warning if the microphone is not muted, or if there are multiple microphones other programs might confuse
- show a warning if any process is listening on a non-local TCP address (make sure local development stays local)
- show a warning if a compatibility layer is running

![screenshot](/bar.png)

# How it does it
`statusline` is an executable that takes no inputs and outputs JSON-encoded data to standard output whenever something important changes.<br/>
Sound-related changes subscribe to published events, this is done via a single long-running (os) process.<br/>
Other changes operate on a timer and also gather data by calling other executables.<br/>

# What to expect
Code optimized for
- reduced resource usage
- minimal count and size of runtime allocations
- performance (where it doesn't increase resource usage)
- a lot of benchmarks

# What NOT to expect
- production code
- test coverage
- very graceful error handling, in this scope it's preferred to say what's wrong and stop execution
- very clean abstractions

# Goals
Usually, programs are optimized for speed.<br/>
This however is a background utility, its performance is secondary to resource usage.<br/>
Most of the changes will generally improve performance as well, but keep in mind that overall execution speed depends mostly on outside factors.<br/>
The primary goal is to reduce the number and size of runtime allocations.<br/>
This will result in less time spent on GC, and a lower average memory footprint when running in the background.<br/>
The memory limit before GC could be reduced, but it's fairly unlikely to hit the limit before GC timeout.<br/>

# Results
As an attempt to quantify the difference between fairly standard (even slightly naive) go code and optimized code, I pulled an older version of the same functionality and ran a benchmark.<br/>
> Code inside the benchmark starts a new process, parses its JSON output, produces and JSON encodes the output, and then pushes it to the standard output (or dev null) 5 times.<br/>

<details>

<summary>go test -bench=UpdateAndRender -run=^# -benchmem -benchtime=10s ./... | tee out</summary>

> branch old-code-compare

```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkNewUpdateAndRender-16    	    2323	   5100528 ns/op	     662 B/op	      15 allocs/op
BenchmarkOldUpdateAndRender-16    	    2193	   6544635 ns/op	   57554 B/op	      99 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	27.282s
```

</details>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 57554 B/op | 99 allocs/op |
| new | 662 B/op | 15 allocs/op |

> [!NOTE]
> That's nearly 100x less memory allocated per operation and a few times fewer allocations.

You would be able to go most of the way without resorting to as extreme measures as I did.<br/>
