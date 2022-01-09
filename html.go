package main

import (
	"flag"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	html "github.com/spudtrooper/gettr/htmlgen"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	user                      = flag.String("user", "", "auth username")
	debug                     = flag.Bool("debug", false, "whether to debug requests")
	token                     = flag.String("token", "", "auth token")
	other                     = flag.String("other", "mtg4america", "other username")
	cacheDir                  = flag.String("cache_dir", ".cache", "cache directory")
	userCreds                 = flag.String("user_creds", ".user_creds.json", "file with user credentials")
	limit                     = flag.Int("limit", 0, "only include this many rows")
	writeCSV                  = flag.Bool("write_csv", false, "write CSV file")
	writeHTML                 = flag.Bool("write_html", false, "write HTML file")
	writeSimpleHTML           = flag.Bool("write_simple_html", false, "write HTML file")
	writeDescriptionsHTML     = flag.Bool("write_desc_html", false, "write HTML file for entries with descriptions")
	writeTwitterFollowersHTML = flag.Bool("write_twitter_followers_html", false, "write HTML file for entries with twitter followers")
)

func realMain() error {
	var client *api.Client
	if *user != "" && *token != "" {
		client = api.MakeClient(*user, *token, api.MakeClientDebug(*debug))
	} else if *userCreds != "" {
		c, err := api.MakeClientFromFile(*userCreds, api.MakeClientDebug(*debug))
		if err != nil {
			return err
		}
		client = c
	} else {
		return errors.Errorf("Must set --user & --token or --creds_file")
	}

	cache := model.MakeCache(*cacheDir)

	if err := html.Generate(client, cache, *other,
		html.GenerateLimit(*limit),
		html.GenerateWriteCSV(*writeCSV),
		html.GenerateWriteDescriptionsHTML(*writeDescriptionsHTML),
		html.GenerateWriteHTML(*writeHTML),
		html.GenerateWriteSimpleHTML(*writeSimpleHTML),
		html.GenerateWriteTwitterFollowersHTML(*writeTwitterFollowersHTML),
	); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
