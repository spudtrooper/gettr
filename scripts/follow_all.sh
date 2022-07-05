#!/bin/sh

other="$@"
if [[ -z "$msg" ]]; then
    others=(
      stevebannon
      mikepompeo
      kayleighmcenany
      oann
      mtg4america
    )
    RANDOM=$$$(date +%s)
    other=${others[$RANDOM % ${#others[@]}]}
    echo "... other: $other"
fi
go run main.go --other $other --actions FollowAll
