package main

import (
	"context"
	"flag"

	"github.com/spudtrooper/gettr/mains/wordcloud"
)

func main() {
	flag.Parse()
	wordcloud.Main(context.Background())
}
