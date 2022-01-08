package model

import (
	"encoding/json"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/or"
)

type cachedUser struct {
	*factory
	username  string
	userInfo  api.UserInfo
	followers []string
	following []string
}

func (u *cachedUser) Username() string { return u.username }

func (u *cachedUser) UserInfo() (api.UserInfo, error) {
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

func (u *cachedUser) Followers(fOpts ...api.AllFollowersOption) (chan User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), 100)

	go func() {
		if len(u.followers) == 0 {
			bytes, err := u.cache.Get("users", u.username, "followers")
			if err != nil {
				errs <- err
				return
			}
			var v []string
			if err := json.Unmarshal(bytes, &v); err != nil {
				errs <- err
				return
			}
			u.followers = v
		}
		for _, f := range u.followers {
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
					users <- u.MakeUser(f)
				}
			}()
		}
		wg.Wait()
		close(users)
	}()

	return users, errs
}

func (u *cachedUser) Following(fOpts ...api.AllFollowingsOption) (chan User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), 100)

	go func() {
		if len(u.following) == 0 {
			bytes, err := u.cache.Get("users", u.username, "following")
			if err != nil {
				errs <- err
				return
			}
			var v []string
			if err := json.Unmarshal(bytes, &v); err != nil {
				errs <- err
				return
			}
			u.following = v
		}
		for _, f := range u.following {
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
					users <- u.MakeUser(f)
				}
			}()
		}
		wg.Wait()
		close(users)
	}()

	return users, errs
}

func (u *cachedUser) Persist(pOpts ...UserPersistOption) error {
	return nil
}
