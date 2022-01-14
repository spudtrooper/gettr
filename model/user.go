package model

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/sets"
	"github.com/thomaso-mirodin/intmath/intgr"
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
	cacheKeyUserInfo          cacheKey = "userInfo"
	cacheKeySkipUserInfo      cacheKey = "skipUserInfo"
	cacheKeyFollowers         cacheKey = "followers"
	cacheKeyFollowersByOffset cacheKey = "followersByOffset"
	cacheKeyFollowersDone     cacheKey = "followersDone"
	cacheKeyFollowing         cacheKey = "following"
	cacheKeyFollowingByOffset cacheKey = "followingByOffset"
	cacheKeyFollowingDone     cacheKey = "followingDone"
)

type User struct {
	*factory
	username               string
	userInfo               api.UserInfo
	followers              []string
	following              []string
	cacheFollowersInMemory bool
	cacheFollowingInMemory bool
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
	if u.has(cacheKeySkipUserInfo) {
		return api.UserInfo{}, nil
	}

	if u.userInfo.Username == "" && u.has(cacheKeyUserInfo) {
		bytes, err := u.cache.GetBytes("users", u.username, "userInfo")
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
	// Followers are either in "followers" (legacy) or sharded into "followersByOffset"
	if u.has(cacheKeyFollowers) {
		if *verboseCacheHits {
			log.Printf("cache hit for followers of %s", u.Username())
		}
		var arr *[]string
		if u.cacheFollowersInMemory {
			arr = &u.followers
		}
		return cachedFollowers(u.username, u.cache, u.factory, arr, fOpts...)
	} else if u.has(cacheKeyFollowersDone) &&
		u.has(cacheKeyFollowersByOffset) {
		// If we've completely read users and sharded them, read from the sharded directory.
		if *verboseCacheHits {
			log.Printf("cache hit for followersByOffset and followersDone of %s", u.Username())
		}
		var arr *[]string
		if u.cacheFollowersInMemory {
			arr = &u.followers
		}
		return cachedFollowersByOffset(u.username, u.cache, u.factory, arr, fOpts...)
	} else if u.has(cacheKeyFollowersByOffset) {
		// In the case we have partially read the followers, populate the in-memory cache.
		if *verboseCacheHits {
			log.Printf("cache hit for followersByOffset of %s", u.Username())
		}

		v, err := u.cache.GetAllStrings("users", u.Username(), string(cacheKeyFollowersByOffset))
		if err != nil {
			log.Printf("ignoring GetAllStrings for %s", u.Username())
		} else {
			if u.cacheFollowersInMemory {
				u.followers = v
			}
		}
	}

	lastOffset := -1
	{
		if u.has(cacheKeyFollowersByOffset) {
			keys, err := u.cache.FindKeys("users", u.Username(), string(cacheKeyFollowersByOffset))
			if err != nil {
				log.Printf("ignoring FindLargestKey: %v", err)
			} else {
				for _, k := range keys {
					n, err := strconv.Atoi(k)
					if err != nil {
						log.Printf("ignoring Atoi: %v", err)
						continue
					}
					lastOffset = intgr.Max(lastOffset, n)
				}
			}
		}
	}
	log.Printf("have last followers offset: %d", lastOffset)
	if lastOffset > 0 {
		fOpts = append(fOpts, api.AllFollowersStart(lastOffset))
	}

	userInfos, userNamesToCache, errs := u.client.AllFollowersParallel(u.username, fOpts...)

	users := make(chan *User)
	usernames := make(chan string)

	go func() {
		// Transfer the partially-populated in-memory cache to the result.
		if u.cacheFollowersInMemory {
			for _, f := range u.followers {
				follower := u.MakeUser(f)
				users <- follower
			}
		}
		// Transfer the newly-read users to the result.
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.userInfo = userInfo
			users <- follower
			usernames <- userInfo.Username
		}
		close(users)
		close(usernames)
	}()

	go func() {
		// Cache by shard.
		for so := range userNamesToCache {
			if err := u.cacheOffsetStrings(so.Strings, cacheKeyFollowersByOffset, so.Offset); err != nil {
				log.Printf("cacheOffsetStrings: error caching followers for %s, offset=%d: %v", u.Username(), so.Offset, err)
			}
		}
		// Mark that we are complete.
		if err := u.cacheBytes("", cacheKeyFollowersDone); err != nil {
			log.Printf("cacheBytes: error caching cacheKeyFollowersDone for %s: %v", u.Username(), err)
		}
	}()

	// Cache all newly-read users in memory.
	go func() {
		var followers []string
		for f := range usernames {
			if u.cacheFollowersInMemory {
				followers = append(followers, f)
			}
		}
		if u.cacheFollowersInMemory {
			u.followers = append(u.followers, followers...)
		}
	}()

	return users, errs
}

func (u *User) FollowersSync(fOpts ...api.AllFollowersOption) ([]*User, error) {
	if u.has(cacheKeyFollowers) {
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
	// Following are either in "following" (legacy) or sharded into "followingByOffset"
	if u.has(cacheKeyFollowing) {
		if *verboseCacheHits {
			log.Printf("cache hit for following of %s", u.Username())
		}
		var arr *[]string
		if u.cacheFollowingInMemory {
			arr = &u.following
		}
		return cachedFollowing(u.username, u.cache, u.factory, arr, fOpts...)
	} else if u.has(cacheKeyFollowingDone) && u.has(cacheKeyFollowingByOffset) {
		// If we've completely read users and sharded them, read from the sharded directory.
		if *verboseCacheHits {
			log.Printf("cache hit for followingByOffset and followingDone of %s", u.Username())
		}
		var arr *[]string
		if u.cacheFollowingInMemory {
			arr = &u.following
		}
		return cachedFollowingByOffset(u.username, u.cache, u.factory, arr, fOpts...)
	} else if u.has(cacheKeyFollowingByOffset) {
		// In the case we have partially read the following, populate the in-memory cache.
		if *verboseCacheHits {
			log.Printf("cache hit for followingByOffset of %s", u.Username())
		}

		v, err := u.cache.GetAllStrings("users", u.Username(), string(cacheKeyFollowingByOffset))
		if err != nil {
			log.Printf("ignoring GetAllStrings for %s", u.Username())
		} else {
			if u.cacheFollowingInMemory {
				u.following = v
			}
		}
	}

	lastOffset := -1
	{
		if u.has(cacheKeyFollowingByOffset) {
			keys, err := u.cache.FindKeys("users", u.Username(), string(cacheKeyFollowingByOffset))
			if err != nil {
				log.Printf("ignoring FindLargestKey: %v", err)
			} else {
				for _, k := range keys {
					n, err := strconv.Atoi(k)
					if err != nil {
						log.Printf("ignoring Atoi: %v", err)
						continue
					}
					lastOffset = intgr.Max(lastOffset, n)
				}
			}
		}
	}
	log.Printf("have last following offset: %d", lastOffset)
	if lastOffset > 0 {
		fOpts = append(fOpts, api.AllFollowingsStart(lastOffset))
	}

	userInfos, userNamesToCache, errs := u.client.AllFollowingParallel(u.username, fOpts...)

	users := make(chan *User)
	usernames := make(chan string)

	go func() {
		// Transfer the partially-populated in-memory cache to the result.
		if u.cacheFollowingInMemory {
			for _, f := range u.following {
				following := u.MakeUser(f)
				users <- following
			}
		}
		// Transfer the newly-read users to the result.
		for userInfo := range userInfos {
			following := u.MakeUser(userInfo.Username)
			following.userInfo = userInfo
			users <- following
			usernames <- userInfo.Username
		}
		close(users)
		close(usernames)
	}()

	go func() {
		// Cache by shard.
		for so := range userNamesToCache {
			if err := u.cacheOffsetStrings(so.Strings, cacheKeyFollowingByOffset, so.Offset); err != nil {
				log.Printf("cacheOffsetStrings: error caching cacheKeyFollowingByOffset for %s, offset=%d: %v", u.Username(), so.Offset, err)
			}
		}
		// Mark that we are complete.
		if err := u.cacheBytes("", cacheKeyFollowingDone); err != nil {
			log.Printf("cacheBytes: error caching cacheKeyFollowingDone for %s: %v", u.Username(), err)
		}
	}()

	// Cache all newly-read users in memory.
	go func() {
		var following []string
		for f := range usernames {
			if u.cacheFollowingInMemory {
				following = append(following, f)
			}
		}
		if u.cacheFollowingInMemory {
			u.following = append(u.following, following...)
		}
	}()

	return users, errs
}

func (u *User) FollowingSync(fOpts ...api.AllFollowingsOption) ([]*User, error) {
	if u.has(cacheKeyFollowing) {
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

func (u *User) has(key cacheKey) bool {
	ok, err := u.cache.Has("user", u.Username(), string(key))
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

func (u *User) cacheOffsetStrings(val []string, part cacheKey, offset int) error {
	return setBytes(u.cache, val, "users", u.Username(), string(part), fmt.Sprintf("%d", offset))
}

var (
	persistDisallow = sets.String([]string{"hectorfbara84"})
)

func (u *User) Persist(pOpts ...UserPersistOption) error {
	if persistDisallow[u.Username()] {
		return nil
	}

	opts := MakeUserPersistOptions(pOpts...)

	should := func(key cacheKey) bool {
		return opts.Force() || !u.has(key)
	}

	persistUserInfo := func(u *User) error {
		if should(cacheKeyUserInfo) {
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

	if should(cacheKeyFollowers) {
		if *verbosePersist {
			log.Printf("persisting followers of %s", u.username)
		}

		followers := make(chan *User)
		go func() {
			users, errs := u.Followers(api.AllFollowersMax(opts.Max()), api.AllFollowersThreads(opts.Threads()))
			for u := range users {
				followers <- u
			}
			for e := range errs {
				log.Printf("error: %v", e)
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

	if should(cacheKeyFollowing) {
		if *verbosePersist {
			log.Printf("persisting following of %s", u.username)
		}

		following := make(chan *User)
		go func() {
			users, errs := u.Following(api.AllFollowingsMax(opts.Max()), api.AllFollowingsThreads(opts.Threads()))
			for u := range users {
				following <- u
			}
			for e := range errs {
				log.Printf("error: %v", e)
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
		var existing []string
		if existingFollowers != nil {
			existing = *existingFollowers
		}
		if len(existing) == 0 {
			bytes, err := cache.GetBytes("users", username, "followers")
			if err != nil {
				errs <- err
				return
			}
			var v []string
			if err := json.Unmarshal(bytes, &v); err != nil {
				errs <- err
				return
			}
			existing = v
		}
		for _, f := range existing {
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

func cachedFollowersByOffset(username string, cache Cache, factory Factory, existingFollowers *[]string, fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if existingFollowers != nil {
			existing = *existingFollowers
		}
		if len(existing) == 0 {
			v, err := cache.GetAllStrings("users", username, string(cacheKeyFollowersByOffset))
			if err != nil {
				errs <- err
				return
			}
			existing = v
		}
		for _, f := range existing {
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
			bytes, err := cache.GetBytes("users", username, "following")
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

func cachedFollowingByOffset(username string, cache Cache, factory Factory, existingFollowing *[]string, fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if existingFollowing != nil {
			existing = *existingFollowing
		}
		if len(existing) == 0 {
			v, err := cache.GetAllStrings("users", username, string(cacheKeyFollowingByOffset))
			if err != nil {
				errs <- err
				return
			}
			existing = v
		}
		for _, f := range existing {
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
