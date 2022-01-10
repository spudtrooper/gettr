package model

import (
	"encoding/json"
	"flag"
	"log"
	"strings"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/sets"
)

var (
	verboseCacheHits = flag.Bool("verbose_cache_hits", false, "log cache hits verbosely")
)

type User struct {
	*factory
	username string
	userInfo api.UserInfo
}

func (u *User) Username() string { return u.username }

func (u *User) UserInfo() (api.UserInfo, error) {
	if u.userInfo.Username == "" {
		has, err := u.cache.Has("users", u.username, "userInfo")
		if err != nil {
			return api.UserInfo{}, err
		}
		if has {
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
	}
	if u.userInfo.Username == "" {
		uinfo, err := u.client.GetUserInfo(u.username)
		if err != nil {
			if strings.HasPrefix(err.Error(), "response error") {
				log.Printf("ignoring response error: %v", err)
				return api.UserInfo{}, nil
			}
			return api.UserInfo{}, err
		}
		u.userInfo = uinfo

		// Cache it.
		go func() {
			if err := u.setBytes(u.userInfo, "users", u.Username(), "userInfo"); err != nil {
				log.Printf("error caching userInfo for %s: %v", u.Username(), err)
			}
		}()
	}
	return u.userInfo, nil
}

func (u *User) Followers(fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	cached, err := u.cache.Has("users", u.Username(), "followers")
	if err != nil {
		log.Printf("error checking for %s followers: %v", u.Username(), err)
		cached = false
	}
	if cached {
		if *verboseCacheHits {
			log.Printf("cache hit for followings of %s", u.Username())
		}
		var followers []string
		users, errs := cachedFollowers(u.username, u.cache, u.factory, &followers, fOpts...)

		// Cache it.
		go func() {
			if err := u.setBytes(followers, "users", u.Username(), "followers"); err != nil {
				log.Printf("error caching followers for %s: %v", u.Username(), err)
			}
		}()

		return users, errs
	}

	userInfos, errs := u.client.AllFollowersParallel(u.username, fOpts...)

	users := make(chan *User)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.userInfo = userInfo
			users <- follower
		}
		close(users)
	}()

	return users, errs
}

func (u *User) Following(fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	cached, err := u.cache.Has("users", u.Username(), "following")
	if err != nil {
		log.Printf("error checking for %s followings: %v", u.Username(), err)
		cached = false
	}
	if cached {
		if *verboseCacheHits {
			log.Printf("cache hit for followings of %s", u.Username())
		}
		var following []string
		users, errs := cachedFollowing(u.username, u.cache, u.factory, &following, fOpts...)

		// Cache it.
		go func() {
			if err := u.setBytes(following, "users", u.Username(), "following"); err != nil {
				log.Printf("error caching followings for %s: %v", u.Username(), err)
			}
		}()

		return users, errs
	}

	userInfos, errs := u.client.AllFollowingsParallel(u.username, fOpts...)

	users := make(chan *User)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.userInfo = userInfo
			users <- follower
		}
		close(users)
	}()

	return users, errs
}

func (u *User) has(parts ...string) bool {
	ok, err := u.cache.Has(parts...)
	return !ok || err != nil
}

var (
	persistDisallow = sets.String([]string{"hectorfbara84"})
)

func (u *User) setBytes(val interface{}, parts ...string) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if err := u.cache.SetBytes(bytes, parts...); err != nil {
		return err
	}
	return nil
}

func (u *User) Persist(pOpts ...UserPersistOption) error {
	if persistDisallow[u.Username()] {
		return nil
	}

	opts := MakeUserPersistOptions(pOpts...)

	should := func(parts ...string) bool {
		return opts.Force() || !u.has(parts...)
	}

	persistUserInfo := func(u *User) error {
		if should("users", u.Username(), "userInfo") {
			userInfo, err := u.UserInfo()
			if err != nil {
				return err
			}
			if err := u.setBytes(userInfo, "users", u.Username(), "userInfo"); err != nil {
				return err
			}
		}
		return nil
	}

	if should("users", u.username, "followers") {
		log.Printf("persisting followers of %s", u.username)

		followers := make(chan *User)
		go func() {
			users, _ := u.Followers(api.AllFollowersMax(opts.Max()), api.AllFollowersThreads(opts.Threads()))
			for u := range users {
				followers <- u
			}
			close(followers)
		}()

		var usernames []string
		for u := range followers {
			usernames = append(usernames, u.Username())
			if err := persistUserInfo(u); err != nil {
				return err
			}
		}
		log.Printf("persisted %d followers of %s", len(usernames), u.username)
		if err := u.setBytes(usernames, "users", u.username, "followers"); err != nil {
			return err
		}
	} else {
		log.Printf("SKIP persisting followers of %s", u.username)
	}

	if should("users", u.username, "following") {
		log.Printf("persisting following of %s", u.username)

		following := make(chan *User)
		go func() {
			users, _ := u.Following(api.AllFollowingsMax(opts.Max()), api.AllFollowingsThreads(opts.Threads()))
			for u := range users {
				following <- u
			}
			close(following)
		}()

		var usernames []string
		for u := range following {
			usernames = append(usernames, u.Username())
			if err := persistUserInfo(u); err != nil {
				return err
			}
		}
		log.Printf("persisted %d following of %s", len(usernames), u.username)
		if err := u.setBytes(usernames, "users", u.username, "following"); err != nil {
			return err
		}
	} else {
		log.Printf("SKIP persisting following of %s", u.username)
	}
	if err := persistUserInfo(u); err != nil {
		return err
	}

	return nil
}
