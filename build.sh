#/bin/sh

go build -gcflags="-m" -ldflags="-s -w"
#upx statusline
