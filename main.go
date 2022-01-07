package main

import (
	"flag"
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/check"
)

var (
	user  = flag.String("user", "", "auth username")
	token = flag.String("token", "", "auth token")
	debug = flag.Bool("debug", false, "whether to debug requests")
	calls = flag.String("calls", "", "comma-delimited list of calls to make")
)

func realMain() error {
	if *user == "" {
		return errors.Errorf("--user required")
	}
	if *token == "" {
		return errors.Errorf("--token required")
	}

	c := api.MakeClient(*user, *token, api.MakeClientDebug(*debug))

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
		info, err := c.GetMuted("spudtrooper")
		if err != nil {
			return err
		}
		log.Printf("GetMuted: %+v", info)
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
