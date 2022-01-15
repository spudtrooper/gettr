#!/bin/sh

set -e
for f in *.go; do
    echo "=== building $f ==="
    go build $f
done
go test api/*.go
go test util/*.go