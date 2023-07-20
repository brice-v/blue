#!/bin/sh
set -o errexit
set -o nounset

govulncheck ./... && go clean -testcache && go test -ldflags '-extldflags "-Wl,--allow-multiple-definition"' -race ./... && go build -ldflags='-s -w -extldflags "-Wl,--allow-multiple-definition"' && strip blue && upx blue && cp blue ~/.blue/bin/

