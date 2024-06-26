build:
	go build -gcflags="-m" -ldflags="-s -w" ./cmd/statusline/

bench:
	go test -bench=. -run=^# -count=1 ./...

benchcpu:
	go test -cpuprofile cpu.prof -bench=. -run=^# -count=1 ./...

httpcpu:
	go tool pprof -http=127.0.0.1:9000 cpu.prof

benchmem:
	go test -memprofile mem.prof -bench=. -run=^# -count=1 -benchmem ./...

benchmemv:
	GOGC=off go test -memprofile mem.prof -memprofilerate=1 -bench=. -run=^# -count=1 -benchmem -benchtime=10s ./internal/...

httpmem:
	go tool pprof -http=127.0.0.1:9000 mem.prof

st:
	staticcheck ./...
