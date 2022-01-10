package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	actions       = flag.String("actions", "", "comma-delimited list of calls to make")
	pause         = flag.Duration("pause", 0, "pause amount between follows")
	offset        = flag.Int("offset", 0, "offset for calls that take offsets")
	other         = flag.String("other", "mtg4america", "other username")
	usernamesFile = flag.String("usernames_file", "", "file containing usernames")
	max           = flag.Int("max", 0, "max to calls")
	threads       = flag.Int("threads", 0, "threads to calls")
	force         = flag.Bool("force", false, "force things")
	text          = flag.String("text", "", "text for posting")
	postID        = flag.String("post_id", "", "post ID for deletion")
)

func realMain() error {
	client, err := api.MakeClientFromFlags()
	if err != nil {
		return err
	}
	cache, err := model.MakeCacheFromFlags()
	if err != nil {
		return err
	}

	actionMap := map[string]bool{}
	if *actions != "" {
		for _, c := range strings.Split(*actions, ",") {
			actionMap[strings.ToLower(c)] = true
		}
	}
	for _, c := range flag.Args() {
		actionMap[strings.ToLower(c)] = true
	}
	shouldReturnedTrueOnce := false
	should := func(s string) bool {
		for k := range actionMap {
			if k == "all" {
				return true
			}
			if s == k {
				return actionMap[s]
			}
		}
		res := actionMap[strings.ToLower(s)]
		if res {
			shouldReturnedTrueOnce = true
		}
		return res
	}

	if len(actionMap) == 0 {
		return errors.Errorf("you need to specify at least one call")
	}

	if should("GetUserInfo") {
		info, err := client.GetUserInfo(*other)
		if err != nil {
			return err
		}
		log.Printf("GetUserInfo: %+v", info)
	}

	if should("GetPublicGlobals") {
		info, err := client.GetPublicGlobals()
		if err != nil {
			return err
		}
		log.Printf("GetPublicGlobals: %+v", info)
	}

	if should("GetSuggestions") {
		info, err := client.GetSuggestions()
		if err != nil {
			return err
		}
		log.Printf("GetSuggestions: %+v", info)
	}

	if should("GetPosts") {
		info, err := client.GetPosts(*other)
		if err != nil {
			return err
		}
		log.Printf("GetPosts: %+v", info)
	}

	if should("GetComments") {
		info, err := client.GetComments("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetComments: %+v", info)
	}

	if should("GetPost") {
		info, err := client.GetPost("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetPost: %+v", info)
	}

	if should("GetMuted") {
		info, err := client.GetMuted()
		if err != nil {
			return err
		}
		log.Printf("GetMuted: %+v", info)
	}

	if should("GetFollowings") {
		info, err := client.GetFollowings(*other, api.FollowingsOffset(*offset), api.FollowingsMax(*max))
		if err != nil {
			return err
		}
		log.Printf("GetFollowings: %+v", info)
	}

	if should("GetAllFollowings") {
		info, err := client.GetAllFollowings(*other)
		if err != nil {
			return err
		}
		log.Printf("GetAllFollowings: %+v", info)
		for _, u := range info {
			if err := client.Follow(u.Username); err != nil {
				return err
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}

	if should("GetFollowers") {
		info, err := client.GetFollowers(*other, api.FollowersOffset(*offset), api.FollowersMax(*max))
		if err != nil {
			return err
		}
		log.Printf("GetFollowers: %+v", info)
		for _, f := range info {
			log.Println(f.Username)
		}
	}

	if should("GetAllFollowers") {
		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for i, u := range userInfos {
				log.Printf("users[%d][%d]: %v", offset, i, u)
				if *pause > 0 {
					time.Sleep(*pause)
				}
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	if should("AllFollowers") {
		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, u := range userInfos {
				if err := client.Follow(u.Username); err != nil {
					return err
				}
				if *pause > 0 {
					time.Sleep(*pause)
				}
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	if should("PrintAllFollowers") {
		username := *other
		if err := client.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, u := range userInfos {
				fmt.Println(u.Username)
			}
			return nil
		}, api.AllFollowersOffset(*offset)); err != nil {
			return err
		}
	}

	if should("AllFollowersFromFile") {
		usernames := make(chan string)
		errs := make(chan error)
		out := make(chan string)

		f, err := os.Open(*usernamesFile)
		if err != nil {
			return err
		}
		defer f.Close()

		go func() {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if u := scanner.Text(); u != "" {
					usernames <- u
				}
			}
			close(usernames)
		}()

		go func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for u := range usernames {
						if err := client.Follow(u); err != nil {
							errs <- err
						} else {
							out <- u
						}
					}
				}()
			}
			wg.Wait()
			close(out)
			close(errs)
		}()

		for u := range out {
			log.Printf("done: %s", u)
		}
		for err := range errs {
			log.Fatalf("error: %v", err)
		}
	}

	if should("Persist") {
		factory := model.MakeFactory(cache, client)
		user := factory.MakeUser(*other)
		if err := user.Persist(model.UserPersistMax(*max), model.UserPersistThreads(*threads), model.UserPersistForce(*force)); err != nil {
			return err
		}
	}

	if should("Read") {
		factory := model.MakeFactory(cache, client)
		u := factory.MakeUser(*other)

		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			i := 0
			for f := range c {
				log.Printf("followers[%d]: %s", i, f.Username())
				i++
			}
		}
		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Following(api.AllFollowingsMax(*max), api.AllFollowingsMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			i := 0
			for f := range c {
				log.Printf("following[%d]: %s", i, f.Username())
				i++
			}
		}
	}

	if should("PersistAll") {
		factory := model.MakeFactory(cache, client)
		u := factory.MakeUser(*other)

		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			for f := range c {
				if err := f.Persist(); err != nil {
					return err
				}
			}
		}
		{
			c := make(chan *model.User)
			go func() {
				users, _ := u.Following(api.AllFollowingsMax(*max), api.AllFollowingsMax(*threads))
				for u := range users {
					c <- u
				}
				close(c)
			}()

			for f := range c {
				if err := f.Persist(); err != nil {
					return err
				}
			}
		}
	}

	if should("CreatePost") {
		info, err := client.CreatePost(*text)
		if err != nil {
			return err
		}
		log.Printf("CreatePost: %+v", info)
	}

	if should("DeletePost") {
		info, err := client.DeletePost(*postID)
		if err != nil {
			return err
		}
		log.Printf("DeletePost: %+v", info)
	}

	if !shouldReturnedTrueOnce {
		return errors.Errorf("no valid actions in %+v", actionMap)
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
