#!/bin/sh

scripts=$(dirname $0)

$scripts/follow_all.sh "$@"
$scripts/like_all.sh "$@"
$scripts/reply_all.sh "$@"
$scripts/share_all.sh "$@"