#!/bin/sh
set -o errexit
set -o nounset

go clean -testcache && go test ./... && go build -ldflags='-s -w' && strip blue && upx blue && cp blue ~/.blue/bin/

