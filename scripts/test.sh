#!/bin/sh

set -e

dir=$(dirname $0)

$dir/unit_tests.sh

for f in mains/*.go; do
    echo "==      building $f"
    go build $f
    bin=$(echo $f | sed 's/.go//' | xargs basename)
    rm $bin
done
for f in *.go; do
    echo "==      building $f"
    go build $f
    bin=$(echo $f | sed 's/.go//')
    rm $bin
done

