package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/check"
)

var (
	user  = flag.String("user", "", "auth username")
	token = flag.String("token", "", "auth token")
	debug = flag.Bool("debug", false, "whether to debug requests")
	calls = flag.String("calls", "", "comma-delimited list of calls to make")
	pause = flag.Duration("pause", 0, "pause amount between follows")
)

func realMain() error {
	if *user == "" {
		return errors.Errorf("--user required")
	}
	if *token == "" {
		return errors.Errorf("--token required")
	}

	callMap := map[string]bool{}
	if *calls != "" {
		for _, c := range strings.Split(*calls, ",") {
			callMap[strings.ToLower(c)] = true
		}
	}
	for _, c := range flag.Args() {
		callMap[strings.ToLower(c)] = true
	}
	should := func(s string) bool {
		for k := range callMap {
			if k == "all" {
				return true
			}
			if s == k {
				return callMap[s]
			}
		}
		return callMap[strings.ToLower(s)]
	}

	if len(callMap) == 0 {
		return errors.Errorf("you need to specify at least one call")
	}

	c := api.MakeClient(*user, *token, api.MakeClientDebug(*debug))

	if should("GetUserInfo") {
		info, err := c.GetUserInfo("spudtrooper")
		if err != nil {
			return err
		}
		log.Printf("GetUserInfo: %+v", info)
	}
	if should("GetPublicGlobals") {
		info, err := c.GetPublicGlobals()
		if err != nil {
			return err
		}
		log.Printf("GetPublicGlobals: %+v", info)
	}
	if should("GetSuggestions") {
		info, err := c.GetSuggestions()
		if err != nil {
			return err
		}
		log.Printf("GetSuggestions: %+v", info)
	}
	if should("GetPosts") {
		info, err := c.GetPosts("mikepompeo")
		if err != nil {
			return err
		}
		log.Printf("GetPosts: %+v", info)
	}
	if should("GetComments") {
		info, err := c.GetComments("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetComments: %+v", info)
	}
	if should("GetPost") {
		info, err := c.GetPost("pmyaf4548d")
		if err != nil {
			return err
		}
		log.Printf("GetPost: %+v", info)
	}
	if should("GetMuted") {
		info, err := c.GetMuted()
		if err != nil {
			return err
		}
		log.Printf("GetMuted: %+v", info)
	}
	if should("GetFollowings") {
		info, err := c.GetFollowings("repmattgaetz")
		if err != nil {
			return err
		}
		log.Printf("GetFollowings: %+v", info)
	}
	if should("GetAllFollowings") {
		info, err := c.GetAllFollowings("mtg4america")
		if err != nil {
			return err
		}
		log.Printf("GetAllFollowings: %+v", info)
		for _, u := range info {
			if err := c.Follow(u.Username); err != nil {
				return err
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}
	if should("GetFollowers") {
		info, err := c.GetFollowers("repmattgaetz")
		if err != nil {
			return err
		}
		log.Printf("GetFollowers: %+v", info)
	}
	if should("AllFollowers") {
		username := "mtg4america"
		if err := c.AllFollowers(username, func(offset int, userInfos api.UserInfos) error {
			log.Printf("following users[%d] of %s", offset, username)
			for _, u := range userInfos {
				if err := c.Follow(u.Username); err != nil {
					return err
				}
				if *pause > 0 {
					time.Sleep(*pause)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
