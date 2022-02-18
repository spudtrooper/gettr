package main

import (
	"context"
	"os"

	"github.com/spudtrooper/gettr/cli"
	"github.com/spudtrooper/goutil/check"
)

func main() {
	check.Err(cli.Main(context.Background(), os.Args))
}
