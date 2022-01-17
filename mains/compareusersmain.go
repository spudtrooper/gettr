package main

import (
	"context"
	"flag"

	"github.com/spudtrooper/gettr/mains/compareusers"
)

func main() {
	flag.Parse()
	compareusers.Main(context.Background())
}
