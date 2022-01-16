#!/bin/sh

set -e

for f in *.go; do
    echo "=== building $f ==="
    go build $f
    bin=$(echo $f | sed 's/.go//')
    rm $bin
done
go test api/*.go
go test util/*.go