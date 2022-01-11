package model

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/sets"
)

var (
	verboseCacheHits = flag.Bool("verbose_cache_hits", false, "log cache hits verbosely")
	verbosePersist   = flag.Bool("verbose_persist", false, "log persisting verbosely")
)

const (
	defaultThreads = 200
)

type cacheKey string

const (
	cacheKeyUserInfo     cacheKey = "userInfo"
	cacheKeySkipUserInfo cacheKey = "skipUserInfo"
	cacheKeyFollowing    cacheKey = "following"
	cacheKeyFollowers    cacheKey = "followers"
)

type User struct {
	*factory
	username  string
	userInfo  api.UserInfo
	followers []string
	following []string
}

func (u *User) MustDebugString() string {
	s, err := u.DebugString()
	check.Err(err)
	return s
}

func (u *User) DebugString() (string, error) {
	userInfo, err := u.UserInfo()
	if err != nil {
		return "", err
	}
	res := fmt.Sprintf("%s(followers: %d, following: %d)", u.Username(), userInfo.Followers(), userInfo.Following())
	return res, nil
}

func (u *User) Username() string { return u.username }

func (u *User) MarkSkipped() error {
	if err := u.cache.Set("users", u.Username(), string(cacheKeySkipUserInfo)); err != nil {
		return err
	}
	return nil
}

func (u *User) UserInfo(uOpts ...UserInfoOption) (api.UserInfo, error) {
	if u.has("users", u.username, string(cacheKeySkipUserInfo)) {
		return api.UserInfo{}, nil
	}

	if u.userInfo.Username == "" && u.has("users", u.username, string(cacheKeyUserInfo)) {
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

	if u.userInfo.Username == "" {
		uinfo, err := u.client.GetUserInfo(u.username)
		if err != nil {
			if strings.HasPrefix(err.Error(), "response error") {
				log.Printf("ignoring response error: %v", err)
				var writeSkipFile bool
				if strings.Contains(err.Error(), "user already deleted") {
					writeSkipFile = true
				} else if opts := MakeUserInfoOptions(uOpts...); opts.DontRetry() {
					writeSkipFile = true
				}
				if writeSkipFile {
					go func() {
						if err := u.MarkSkipped(); err != nil {
							log.Printf("error caching skipUserInfo for %s: %v", u.Username(), err)
						}
					}()
				}
				return api.UserInfo{}, nil
			}
			return api.UserInfo{}, err
		}
		u.userInfo = uinfo

		// Cache it.
		go func() {
			if err := u.cacheBytes(u.userInfo, cacheKeyUserInfo); err != nil {
				log.Printf("error caching userInfo for %s: %v", u.Username(), err)
			}
		}()
	}

	return u.userInfo, nil
}

func (u *User) Followers(fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	if u.has("users", u.Username(), "followers") {
		if *verboseCacheHits {
			log.Printf("cache hit for followers of %s", u.Username())
		}
		lenBefore := len(u.followers)
		users, errs := cachedFollowers(u.username, u.cache, u.factory, &u.followers, fOpts...)

		if lenBefore != len(u.followers) {
			go func() {
				if err := u.cacheBytes(u.followers, cacheKeyFollowers); err != nil {
					log.Printf("error caching followers for %s: %v", u.Username(), err)
				}
			}()
		}

		return users, errs
	}

	userInfos, errs := u.client.AllFollowersParallel(u.username, fOpts...)

	users := make(chan *User)
	usernames := make(chan string)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.userInfo = userInfo
			users <- follower
			usernames <- userInfo.Username
		}
		close(users)
		close(usernames)
	}()

	// Cache it
	go func() {
		var followers []string
		for f := range usernames {
			followers = append(followers, f)
		}
		u.followers = followers
		if err := u.cacheBytes(u.followers, cacheKeyFollowers); err != nil {
			log.Printf("error caching followers for %s: %v", u.Username(), err)
		}
	}()

	return users, errs
}

func (u *User) FollowersSync(fOpts ...api.AllFollowersOption) ([]*User, error) {
	if u.has("users", u.Username(), "followers") {
		if *verboseCacheHits {
			log.Printf("cache hit for followers of %s", u.Username())
		}
		lenBefore := len(u.followers)
		cachedFollowers(u.username, u.cache, u.factory, &u.followers, fOpts...)

		if lenBefore != len(u.followers) {
			go func() {
				if err := u.cacheBytes(u.followers, cacheKeyFollowers); err != nil {
					log.Printf("error caching followers for %s: %v", u.Username(), err)
				}
			}()
		}

		var res []*User
		for _, f := range u.followers {
			res = append(res, u.factory.MakeUser(f))
		}
		return res, nil
	}

	var res []*User
	var followers []string
	if err := u.client.AllFollowers(u.username, func(offset int, userInfos api.UserInfos) error {
		for _, ui := range userInfos {
			res = append(res, u.factory.MakeUser(ui.Username))
			followers = append(followers, ui.Username)
		}
		return nil
	}, fOpts...); err != nil {
		return nil, err
	}
	u.followers = followers

	// Cache it
	go func() {
		if err := u.cacheBytes(u.followers, cacheKeyFollowers); err != nil {
			log.Printf("error caching followers for %s: %v", u.Username(), err)
		}
	}()

	return res, nil
}

func (u *User) Following(fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	if u.has("users", u.Username(), "following") {
		if *verboseCacheHits {
			log.Printf("cache hit for followings of %s", u.Username())
		}
		lenBefore := len(u.following)
		users, errs := cachedFollowing(u.username, u.cache, u.factory, &u.following, fOpts...)

		if lenBefore != len(u.following) {
			go func() {
				if err := u.cacheBytes(u.following, cacheKeyFollowing); err != nil {
					log.Printf("error caching followings for %s: %v", u.Username(), err)
				}
			}()
		}

		return users, errs
	}

	userInfos, errs := u.client.AllFollowingsParallel(u.username, fOpts...)

	users := make(chan *User)
	usernames := make(chan string)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.userInfo = userInfo
			users <- follower
			usernames <- userInfo.Username
		}
		close(users)
		close(usernames)
	}()

	// Cache it
	go func() {
		var following []string
		for f := range usernames {
			following = append(following, f)
		}
		u.following = following
		if err := u.cacheBytes(u.following, cacheKeyFollowing); err != nil {
			log.Printf("error caching following for %s: %v", u.Username(), err)
		}
	}()

	return users, errs
}

func (u *User) FollowingSync(fOpts ...api.AllFollowingsOption) ([]*User, error) {
	if u.has("users", u.Username(), "following") {
		if *verboseCacheHits {
			log.Printf("cache hit for following of %s", u.Username())
		}
		lenBefore := len(u.following)
		cachedFollowing(u.username, u.cache, u.factory, &u.following, fOpts...)

		if lenBefore != len(u.following) {
			go func() {
				if err := u.cacheBytes(u.following, cacheKeyFollowing); err != nil {
					log.Printf("error caching following for %s: %v", u.Username(), err)
				}
			}()
		}

		var res []*User
		for _, f := range u.following {
			res = append(res, u.factory.MakeUser(f))
		}
		return res, nil
	}

	var res []*User
	var following []string
	if err := u.client.AllFollowings(u.username, func(offset int, userInfos api.UserInfos) error {
		for _, ui := range userInfos {
			res = append(res, u.factory.MakeUser(ui.Username))
			following = append(following, ui.Username)
		}
		return nil
	}, fOpts...); err != nil {
		return nil, err
	}
	u.following = following

	// Cache it
	go func() {
		if err := u.cacheBytes(u.following, cacheKeyFollowing); err != nil {
			log.Printf("error caching following for %s: %v", u.Username(), err)
		}
	}()

	return res, nil
}

func (u *User) has(parts ...string) bool {
	ok, err := u.cache.Has(parts...)
	if err != nil {
		log.Printf("has: ignoring error: %v", err)
		return false
	}
	return ok
}

func setBytes(cache Cache, val interface{}, parts ...string) error {
	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if err := cache.SetBytes(bytes, parts...); err != nil {
		return err
	}
	return nil
}

func (u *User) cacheBytes(val interface{}, part cacheKey) error {
	return setBytes(u.cache, val, "users", u.Username(), string(part))
}

var (
	persistDisallow = sets.String([]string{"hectorfbara84"})
)

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
			if err := u.cacheBytes(userInfo, cacheKeyUserInfo); err != nil {
				return err
			}
		}
		return nil
	}

	if should("users", u.username, "followers") {
		if *verbosePersist {
			log.Printf("persisting followers of %s", u.username)
		}

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
		if err := u.cacheBytes(usernames, cacheKeyFollowers); err != nil {
			return err
		}
	} else {
		if *verbosePersist {
			log.Printf("SKIP persisting followers of %s", u.username)
		}
	}

	if should("users", u.username, "following") {
		if *verbosePersist {
			log.Printf("persisting following of %s", u.username)
		}

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
		if err := u.cacheBytes(usernames, cacheKeyFollowing); err != nil {
			return err
		}
	} else {
		if *verbosePersist {
			log.Printf("SKIP persisting following of %s", u.username)
		}
	}
	if err := persistUserInfo(u); err != nil {
		return err
	}

	return nil
}

func cachedFollowers(username string, cache Cache, factory Factory, existingFollowers *[]string, fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

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

func cachedFollowing(username string, cache Cache, factory Factory, existingFollowers *[]string, fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

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
