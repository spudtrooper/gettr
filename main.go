package main

import (
	"flag"
	"log"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/gettr/api"
	"github.com/spudtrooper/goutil/check"
)

var (
	user  = flag.String("user", "", "auth username")
	token = flag.String("token", "", "auth token")
	debug = flag.Bool("debug", false, "whether to debug requests")
)

func realMain() error {
	if *user == "" {
		return errors.Errorf("--user required")
	}
	if *token == "" {
		return errors.Errorf("--token required")
	}

	c := api.MakeClient(*user, *token, api.MakeClientDebug(*debug))

	{
		info, err := c.GetUserInfo("spudtrooper")
		if err != nil {
			return err
		}
		log.Printf("GetUserInfo: %+v", info)
	}
	{
		info, err := c.GetPublicGlobals()
		if err != nil {
			return err
		}
		log.Printf("GetPublicGlobals: %+v", info)
	}
	{
		info, err := c.GetSuggestions()
		if err != nil {
			return err
		}
		log.Printf("GetSuggestions: %+v", info)
	}
	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
