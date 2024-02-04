#!/bin/sh
set -o errexit
set -o nounset

govulncheck ./... && go clean -testcache && go env -w CGO_ENABLED=1 && go test -race ./... && go env -w CGO_ENABLED=0 && go build -ldflags='-s -w -extldflags="static"' -tags="static" -o blues && strip blues && upx blues && cp blues ~/.blue/bin/

