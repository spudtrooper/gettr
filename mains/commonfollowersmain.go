package main

import (
	"context"
	"flag"

	"github.com/spudtrooper/gettr/mains/commonfollowers"
)

func main() {
	flag.Parse()
	commonfollowers.Main(context.Background())
}
