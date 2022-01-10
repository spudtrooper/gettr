#!/bin/sh

set -e

go build main.go
go build html.go
go test api/*.go
go test util/*.go