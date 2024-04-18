# Goals
Usually, programs are optimized for speed.<br/>
This however is a background utility, its performance is secondary to resource usage.<br/>
Most of the changes will generally improve performance as well, but keep in mind that overall execution speed depends mostly on outside factors.<br/>
The primary goal is to reduce the number and size of runtime allocations.<br/>
This will result in less time spent on GC, and a lower average memory footprint when running in the background.<br/>
The memory limit before GC could be reduced, but it's fairly unlikely to hit the limit before GC timeout.<br/>

# Results
As an attempt to quantify the difference between fairly standard (even slightly naive) go code and optimized code, I pulled an older version of the same functionality and ran a benchmark.<br/>
Code inside the benchmark starts a new process, parses its JSON output, produces and JSON encodes the output, and then pushes it to the standard output (or dev null) 5 times.<br/>

```
branch old-code-compare
go test -bench=UpdateAndRender -run=^# -benchmem -benchtime=10s ./... | tee out

goos: linux
goarch: amd64
pkg: github.com/mjw6i/statusline/internal
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkNewUpdateAndRender-16    	    2323	   5100528 ns/op	     662 B/op	      15 allocs/op
BenchmarkOldUpdateAndRender-16    	    2193	   6544635 ns/op	   57554 B/op	      99 allocs/op
PASS
ok  	github.com/mjw6i/statusline/internal	27.282s
```
> [!NOTE]
> That's nearly 100x less memory allocated per operation and a few times fewer allocations.

You would be able to go most of the way without resorting to as extreme measures as I did.<br/>
