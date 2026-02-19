#!/bin/sh
set -o errexit
set -o nounset

go env -w CGO_ENABLED=1 && govulncheck ./... && go clean -testcache && go test -race ./... && go build -ldflags='-s -w' && strip blue && upx blue && cp blue ~/.blue/bin/

