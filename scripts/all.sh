#!/bin/sh

scripts=$(dirname $0)

$scripts/unfollow_all.sh  --banner "$@"
$scripts/follow_all.sh    --banner "$@"
$scripts/like_all.sh      --banner "$@"
$scripts/reply_all.sh     --banner "$@"
$scripts/share_all.sh     --banner "$@"