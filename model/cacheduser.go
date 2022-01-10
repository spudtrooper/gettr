package model

import (
	"encoding/json"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/or"
)

type CachedUser struct {
	*factory
	username  string
	userInfo  api.UserInfo
	followers []string
	following []string
}

func (u *CachedUser) Username() string { return u.username }

func (u *CachedUser) UserInfo() (api.UserInfo, error) {
	if u.userInfo.Username == "" {
		bytes, err := u.cache.Get("users", u.username, "userInfo")
		if err != nil {
			return api.UserInfo{}, err
		}
		var v api.UserInfo
		if err := json.Unmarshal(bytes, &v); err != nil {
			return api.UserInfo{}, err
		}
		u.userInfo = v
	}
	return u.userInfo, nil
}

func (u *CachedUser) Followers(fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	return cachedFollowers(u.username, u.cache, u.factory, &u.followers, fOpts...)
}

func cachedFollowers(username string, cache Cache, factory Factory, existingFollowers *[]string, fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), 100)

	go func() {
		if len(*existingFollowers) == 0 {
			bytes, err := cache.Get("users", username, "followers")
			if err != nil {
				errs <- err
				return
			}
			var v []string
			if err := json.Unmarshal(bytes, &v); err != nil {
				errs <- err
				return
			}
			*existingFollowers = v
		}
		for _, f := range *existingFollowers {
			followers <- f
		}
		close(followers)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for f := range followers {
					users <- factory.MakeUser(f)
				}
			}()
		}
		wg.Wait()
		close(users)
	}()

	return users, errs
}

func (u *CachedUser) Following(fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	return cachedFollowing(u.username, u.cache, u.factory, &u.followers, fOpts...)
}

func cachedFollowing(username string, cache Cache, factory Factory, existingFollowers *[]string, fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), 100)

	go func() {
		if len(*existingFollowers) == 0 {
			bytes, err := cache.Get("users", username, "following")
			if err != nil {
				errs <- err
				return
			}
			var v []string
			if err := json.Unmarshal(bytes, &v); err != nil {
				errs <- err
				return
			}
			*existingFollowers = v
		}
		for _, f := range *existingFollowers {
			following <- f
		}
		close(following)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for f := range following {
					users <- factory.MakeUser(f)
				}
			}()
		}
		wg.Wait()
		close(users)
	}()

	return users, errs
}
