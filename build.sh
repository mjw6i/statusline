#/bin/sh

go build -ldflags="-s -w" -o statusline
upx statusline
