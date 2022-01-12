package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	threads = flag.Int("threads", 500, "number of threads to process user dirs")
)

func realMain() {
	factory, err := model.MakeFactoryFromFlags()
	check.Err(err)

	keys, errs, err := factory.Cache().FindKeysChannels("users")
	check.Err(err)

	type follower struct {
		username  string
		followers int
	}
	followers := make(chan follower)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < *threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for k := range keys {
					u := factory.MakeUser(k)
					f, err := u.GetFollowers()
					if err != nil {
						log.Printf("GetFollowers: %v", err)
						continue
					}
					followers <- follower{username: k, followers: f}
				}
				for e := range errs {
					log.Printf("ignoring error: %v", e)
				}
			}()
		}
		wg.Wait()
		close(followers)
	}()

	var max follower
	for f := range followers {
		if f.followers > max.followers {
			max = f
			log.Printf("new max: %+v", max)
		}
	}
	fmt.Printf("max: %+v\n", max)
}

func main() {
	flag.Parse()
	realMain()
}
