#/bin/sh

go build -gcflags="-m" -ldflags="-s -w" -o statusline
upx statusline
