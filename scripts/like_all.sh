#!/bin/sh

set -e

other="$@"
if [[ -z "$msg" ]]; then
    other="tuckercarlson"
fi
go run main.go --other $other --actions LikeAll