package main

import (
	"flag"
	"log"

	"github.com/pkg/errors"
	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/model"
	"github.com/spudtrooper/goutil/check"
)

var (
	user          = flag.String("user", "", "auth username")
	token         = flag.String("token", "", "auth token")
	debug         = flag.Bool("debug", false, "whether to debug requests")
	calls         = flag.String("calls", "", "comma-delimited list of calls to make")
	pause         = flag.Duration("pause", 0, "pause amount between follows")
	offset        = flag.Int("offset", 0, "offset for calls that take offsets")
	other         = flag.String("other", "mtg4america", "other username")
	usernamesFile = flag.String("usernames_file", "", "file containing usernames")
	cacheDir      = flag.String("cache_dir", ".cache", "cache directory")
	max           = flag.Int("max", 0, "max to calls")
	threads       = flag.Int("threads", 0, "threads to calls")
	force         = flag.Bool("force", false, "force things")
	userCreds     = flag.String("user_creds", ".user_creds.json", "file with user credentials")
)

func realMain() error {

	var client *api.Client
	if *user != "" && *token != "" {
		client = api.MakeClient(*user, *token, api.MakeClientDebug(*debug))
	} else if *userCreds != "" {
		c, err := api.MakeClientFromFile(*userCreds, api.MakeClientDebug(*debug))
		check.Err(err)
		client = c
	} else {
		return errors.Errorf("Must set --user & --token or --creds_file")
	}

	cache := model.MakeCache(*cacheDir)
	factory := model.MakeFactory(cache, client)
	u := factory.MakeCachedUser(*other)

	followers := make(chan model.User)
	go func() {
		users, _ := u.Followers(api.AllFollowersMax(*max), api.AllFollowersMax(*threads))
		for u := range users {
			followers <- u
		}
		close(followers)
	}()

	i := 0
	for f := range followers {
		userInfo, err := f.UserInfo()
		check.Err(err)
		following := userInfo.Following()
		followers := userInfo.Followers()
		log.Printf("followers[%d]: %s following=%d followers=%d", i, f.Username(), following, followers)
		i++
	}

	return nil
}

func main() {
	flag.Parse()
	check.Err(realMain())
}
