package main

import (
	"context"
	"flag"

	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	threads = flag.Int("threads", 0, "threads to calls")
)

func realMain(ctx context.Context) {
	factory, err := model.MakeFactoryFromFlags(ctx)
	check.Err(err)
	if factory != nil {

	}
}

func main() {
	flag.Parse()
	realMain(context.Background())
}
