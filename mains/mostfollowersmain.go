package main

import (
	"context"
	"flag"

	"github.com/spudtrooper/gettr/mains/mostfollowers"
)

func main() {
	flag.Parse()
	mostfollowers.Main(context.Background())
}
