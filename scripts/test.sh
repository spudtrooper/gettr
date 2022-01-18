#!/bin/sh

set -e

dir=$(dirname $0)

$dir/unit_tests.sh

for f in $(find mains -name '*.go' | xargs); do
    echo "==      building $f"
    go build $f
    bin=$(echo $f | sed 's/.go//' | xargs basename)
    rm -f $bin
done
for f in *.go; do
    echo "==      building $f"
    go build $f
    bin=$(echo $f | sed 's/.go//')
    rm $bin
done

