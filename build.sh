#!/bin/sh

set -e
export GOFLAGS="-mod=vendor"

version=`git tag --sort=-version:refname | head -n 1`
hash=`git rev-parse HEAD`

go build -ldflags "-s -w -X main.version=$version -X main.commitHash=$hash" -tags netgo -o bin cmd/main.go
upx -9 -k bin
