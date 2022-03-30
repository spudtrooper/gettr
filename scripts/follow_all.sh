#!/bin/sh

set -e

other="$@"
if [[ -z "$msg" ]]; then
    other="mikepompeo"
fi
go run main.go --other $other --actions FollowAll
