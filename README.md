# What it does?
It functions similarly to a taskbar (in an environment where there is no taskbar).
- display the day of the week, date and time
- display current volume if recently changed
- show a warning if the microphone is not muted, or if there are multiple microphones other programs might confuse
- show a warning if any process is listening on a non-local TCP address (make sure local development stays local)
- show a warning if a compatibility layer is running

![screenshot of the bar](/bar.png)

# How it does it?
`statusline` is an executable that takes no inputs and outputs JSON-encoded data to standard output whenever something important changes.<br/>
Sound-related changes subscribe to published events, this is done via a single long-running (os) process.<br/>
Other changes operate on a timer and also gather data by calling other executables.<br/>

# What to expect?
Code optimized for
- reduced resource usage
- minimal count and size of runtime allocations
- performance (where it doesn't increase resource usage)
- a lot of benchmarks

# What NOT to expect?
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

<summary>benchmark</summary>

> go test -bench=UpdateAndRender -run=^# -benchmem -benchtime=10s ./... | tee out<br/>
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

<br/>

> [!NOTE]
> That's nearly **100x** less memory allocated per operation and a few times fewer allocations.

<br/>

> [!IMPORTANT]
> You would be able to go most of the way without resorting to as extreme measures as I did.

## Performance characteristics
At the end, processing the data and outputting results doesn't allocate.<br/>
I'll dig into it deeper later on, but the allocations are caused by calling other executables.<br/>

<details>

<summary>benchmark</summary>

> GOGC=off go test -bench=. -run=^# -benchmem -benchtime=10s ./internal/... | tee out

```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkBarRenderHeader-16    	34269534	       346.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkBarRenderAll-16       	30892675	       384.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkUpdateAll-16          	     484	  23466704 ns/op	    2414 B/op	      56 allocs/op
BenchmarkGetDate-16            	55408135	       192.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetIP-16              	    4033	   2893564 ns/op	     664 B/op	      15 allocs/op
BenchmarkGetSinks-16           	    2445	   5177240 ns/op	     659 B/op	      15 allocs/op
BenchmarkGetSources-16         	    2316	   5083940 ns/op	     662 B/op	      15 allocs/op
BenchmarkSubscribe-16          	    3372	   3589968 ns/op	    1094 B/op	      24 allocs/op
BenchmarkEventLine-16          	14473698	       791.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetXWayland-16        	    1282	   9754279 ns/op	     400 B/op	      11 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	124.993s
```

</details>

![screenshot of pprof](/pprof.png)

> [!NOTE]
> Highlighted nodes are responsible for calling outside executables.<br/>
> They represent a custom abstraction over `os.StartProcess`, implemented to reduce the overall cost.

# Steps taken

## Caching
The protocol dictates that each update needs to send the complete current state.<br/>
It's preferable to update each part independently, e.g., update the clock without fetching a list of open ports.<br/>

## Writing the output
Writing the output is an important part of this program.<br/>
`fmt` and `log` packages are very convenient, but they are not performant.<br/>
There's a variety of functions that allow you to print bytes or byte slices that don't allocate extra memory.<br/>

| Version | ns/op | B/op | allocs/op |
| ---: | ---: | ---: | ---: |
| old | 813.5 ns/op | 120 B/op | 5 allocs/op |
| new | 394.6 ns/op | 0 B/op | 0 allocs/op |

> `BenchmarkBarRenderAll` between commits `ec8346f` and `40917a5`

The most performant way of writing the data to stdout is to buffer it and then flush it in a single call.<br/>

## Reusing output buffers
One of the easiest and most beneficial optimizations in this repository.<br/>

```diff
- out, err := exec.Command("pactl", "--format=json", "list", "sources").Output()
+ cmd := exec.Command("pactl", "--format=json", "list", "sources")
+ cmd.Stdout = &s.buffer
```

> [!NOTE]
> The example above is trivial, unfortunately, it's usually a fair bit harder.

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 57141 B/op | 92 allocs/op |
| new |  8079 B/op | 73 allocs/op |

<details>

<summary>
benchmarks
</summary>

> go test -bench=GetSources -run=^# -benchmem -benchtime=10s . | tee old<br/>
> commit eaf6d30
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetSources 	    2335	   6550233 ns/op	   57141 B/op	      92 allocs/op
PASS
ok  	github.com/mjw6i/statusline	15.818s
```
> go test -bench=GetSources -run=^# -benchmem -benchtime=10s . | tee new<br/>
> commit b493778 (edited to remove path resolution optimization)
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetSources 	    2310	   5429670 ns/op	    8079 B/op	      73 allocs/op
PASS
ok  	github.com/mjw6i/statusline	13.071s
```

</details>

## JSON parsing
JSON prioritizes readability and portability over space and computation efficiency.<br/>
That being said, there are multiple ways of improving the performance of JSON parsing.<br/>
Beyond the standard library, there are multiple competing fast struct-oriented decoders.<br/>
In this project, I'm working with relatively large JSON objects with very little data I'm interested in inside of them.<br/>
This relatively unique combination allows me to parse incoming JSON as a stream of bytes, which should be the most efficient way of approaching this problem.</br>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 1070 B/op | 27 allocs/op |
| new | 662 B/op | 15 allocs/op |

<details>

<summary>
benchmarks
</summary>

> go test -bench=GetSources -run=^# -benchmem -benchtime=10s ./internal/... | tee old
> commit 49fd1fc
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetSources-16    	    2314	   5637531 ns/op	    1070 B/op	      27 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.575s
```
> go test -bench=GetSources -run=^# -benchmem -benchtime=10s ./internal/... | tee new
> commit a27323f
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetSources-16    	    2342	   5603520 ns/op	     662 B/op	      15 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.646s
```

</details>

![screenshot of pprof](/pprof_json_parsing.png)

> [!NOTE]
> Using this parser can be a lot faster, and allocates no memory on its own.<br/>
> In this case, it drops nearly half of the allocations and a third of the allocated memory for the whole operation.<br/>

## JSON encoding
Usually, solutions for JSON decoding have an encoding counterpart.<br/>
This time it's not the case, but since I'm using fairly small and uniform objects I'm just writing them manually.<br/>
This change doesn't make a huge difference in overall performance, but every bit helps.<br/>
Surprisingly, replacing the std implementation of JSON reduced the binary size from 2.3 MB to 1.7 MB.<br/>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 5993 B/op | 34 allocs/op |
| new | 5807 B/op | 31 allocs/op |

<details>

<summary>
benchmarks
</summary>

Benchmarks measure only `UpdateXWayland`.

> go test -bench=UpdateAll -run=^# -benchmem -benchtime=10s ./internal/... | tee old<br/>
> commit 621dd54
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkUpdateAll-16    	    1230	  10230095 ns/op	    5993 B/op	      34 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.574s
```
> go test -bench=UpdateAll -run=^# -benchmem -benchtime=10s ./internal/... | tee new<br/>
> commit 58d4572
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkUpdateAll-16    	    1237	   9723971 ns/op	    5807 B/op	      31 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.013s
```

</details>

## Improving on os.Exec
`os.Exec` provides a generic implementation supporting a lot of use cases.<br/>
Writing a more specialized implementation can reduce memory footprint.<br/>

### Avoid repeated work where API allows it
Passing the executable name to `exec.Command()` causes it to resolve the absolute path on each run.<br/>
This work can be done once during startup.<br/>
Doing so reduces the number of allocations with a relatively small cost of increased complexity.<br/>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 7104 B/op | 48 allocs/op |
| new | 5803 B/op | 31 allocs/op |

> the difference in context of a complete operation

Similarly, specifying ENV avoids the default behavior of copying the entire parent environment.<br/>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 13988 B/op | 114 allocs/op |
| new |  9443 B/op | 111 allocs/op |

> the difference in context of a complete operation

### Using os.StartProcess directly
At this point further reductions in runtime allocated memory require dropping the `os.Exec` abstraction layer.<br/>

| Version | B/op | allocs/op |
| ---: | ---: | ---: |
| old | 1304 B/op | 29 allocs/op |
| new | 400 B/op | 11 allocs/op |

> the difference in context of a complete operation

<details>

<summary>
benchmarks
</summary>

> old
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetXWayland-16    	    1210	   9917418 ns/op	    1304 B/op	      29 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.006s
```
> new
```
goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkGetXWayland-16    	    1237	  10197449 ns/op	     400 B/op	      11 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	13.598s
```

</details>

# What else is there?
Details were omitted, implementation also features:
- parsing tab-separated output
- dealing with a long-running stream of JSON events while reusing a small buffer
