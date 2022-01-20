#!/bin/sh

set -e

go test api/*.go
go test util/*.go
go test model/*.go