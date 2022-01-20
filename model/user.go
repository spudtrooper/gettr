package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/gettr/log"
	"github.com/spudtrooper/gettr/todo"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/must"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/parallel"
	"github.com/spudtrooper/goutil/sets"
	"github.com/thomaso-mirodin/intmath/intgr"
)

const (
	defaultThreads = 200
)

type User struct {
	*factory
	username               string
	userInfo               api.UserInfo
	followers              []string
	following              []string
	cacheFollowersInMemory bool // currently always false
	cacheFollowingInMemory bool // currently always false
	checkSkip              bool // currently always false (this is a complete waste)
	cachePostsInMemory     bool // currently always false
}

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

func (u *User) MustDebugString(ctx context.Context) string {
	s, err := u.DebugString(ctx)
	check.Err(err)
	return s
}

func (u *User) DebugString(ctx context.Context) (string, error) {
	userInfo, err := u.UserInfo(ctx)
	if err != nil {
		return "", err
	}
	res := fmt.Sprintf("%s(followers: %d, following: %d)", u.Username(), userInfo.Followers(), userInfo.Following())
	return res, nil
}

func (u *User) Username() string { return u.username }

func (u *User) MarkSkipped() error {
	if err := u.cache.Set(u.cacheParts(cacheKeySkipUserInfo)...); err != nil {
		return err
	}
	return nil
}

func (u *User) UserInfo(ctx context.Context, uOpts ...UserInfoOption) (api.UserInfo, error) {
	if u.opts().userInfoUsingDiskCache {
		return u.userInfoUsingDiskCache(ctx, uOpts...)
	}
	return u.userInfoUsingDB(ctx, uOpts...)
}

func (u *User) userInfoUsingDB(ctx context.Context, uOpts ...UserInfoOption) (api.UserInfo, error) {
	if u.checkSkip {
		userOptions, err := u.db.GetUserOptions(ctx, u.username)
		if err == nil && userOptions != nil && userOptions.Skip {
			return api.UserInfo{}, nil
		}
	}

	if u.userInfo.OUsername == "" {
		userInfo, err := u.db.GetUserInfo(ctx, u.username)
		if err != nil && !noUsers(err) {
			return api.UserInfo{}, err
		}
		if userInfo != nil {
			u.userInfo = *userInfo
		}
	}

	if u.userInfo.OUsername == "" {
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
						if err := u.db.SetUserSkip(ctx, u.Username(), true); err != nil {
							log.Printf("SetUserSkip for %s: %v", u.Username(), err)
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
			if err := u.db.SetUserInfo(ctx, u.Username(), u.userInfo); err != nil {
				log.Printf("SetUserInfo for %s: %v", u.Username(), err)
			}
		}()
	}

	return u.userInfo, nil
}

func (u *User) userInfoUsingDiskCache(ctx context.Context, uOpts ...UserInfoOption) (api.UserInfo, error) {
	if u.has(cacheKeySkipUserInfo) {
		return api.UserInfo{}, nil
	}

	if u.userInfo.Username == "" && u.has(cacheKeyUserInfo) {
		bytes, err := u.cache.GetBytes(u.cacheParts(cacheKeyUserInfo)...)
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

func (u *User) Followers(ctx context.Context, fOpts ...UserFollowersOption) (chan *User, chan error) {
	opts := MakeUserFollowersOptions(fOpts...)
	convertOpts := func(opts UserFollowersOptions) []api.AllFollowersOption {
		return []api.AllFollowersOption{
			api.AllFollowersIncl(opts.Incl()),
			api.AllFollowersForce(opts.Force()),
			api.AllFollowersMax(opts.Max()),
			api.AllFollowersOffset(opts.Offset()),
			api.AllFollowersStart(opts.Start()),
			api.AllFollowersThreads(opts.Threads()),
		}
	}
	if opts.FromDisk() || u.opts().followersUsingDiskCache {
		return u.followersUsingDiskCache(ctx, convertOpts(opts)...)
	}
	return u.followersUsingDB(ctx, convertOpts(opts)...)
}

func (u *User) followersUsingDB(ctx context.Context, fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)
	threads := or.Int(opts.Threads(), defaultThreads)

	if !opts.Force() {
		done, err := u.db.GetUserFollowersDone(ctx, u.Username())
		check.Err(err)
		log.Printf("followersUsingDB: done=%t", done)
		if done {
			users := make(chan *User)

			followers, errors, err := u.db.GetFollowers(ctx, u.Username())
			check.Err(err)

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

			return users, errors
		}
	}

	users := make(chan *User)
	usernames := make(chan string)

	if opts.Force() {
		log.Printf("deleting followers of %s", u.username)
		if err := u.db.DeleteFollowers(ctx, u.username); err != nil {
			todo.SkipErr("deleteFollowers", err)
		}
	} else {
		lastOffset, err := u.db.GetUserMaxFollowerOffset(ctx, u.Username())
		check.Err(err)
		log.Printf("have last followers offset: %d", lastOffset)
		if lastOffset > 0 {
			fOpts = append(fOpts, api.AllFollowersStart(lastOffset))
		}
	}

	userInfos, userNamesToCache, errs := u.client.AllFollowersParallel(u.username, fOpts...)

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
			if err := u.db.SetFollowers(ctx, u.Username(), so.Offset, so.Strings); err != nil {
				log.Printf("SetFollowers: error caching followers for %s, offset=%d: %v", u.Username(), so.Offset, err)
			}
		}
		// Mark that we are complete.
		if err := u.db.SetUserFollowersDone(ctx, u.Username(), true); err != nil {
			log.Printf("SetUserFollowersDone: error caching followers for %s: %v", u.Username(), err)
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

func (u *User) followersUsingDiskCache(ctx context.Context, fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	// // Followers are either in "followers" (legacy) or sharded into "followersByOffset"
	// if u.has(cacheKeyFollowers) {
	// 	if u.opts().verboseCacheHits {
	// 		log.Printf("cache hit for followers of %s", u.Username())
	// 	}
	// 	return u.cachedFollowers(fOpts...)
	// } else
	if u.has(cacheKeyFollowersDone) && u.has(cacheKeyFollowersByOffset) {
		// If we've completely read users and sharded them, read from the sharded directory.
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for followersByOffset and followersDone of %s", u.Username())
		}
		return u.cachedFollowersByOffset(fOpts...)
	} else if u.has(cacheKeyFollowersByOffset) {
		// In the case we have partially read the followers, populate the in-memory cache.
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for followersByOffset of %s", u.Username())
		}

		v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowersByOffset)...)
		if err != nil {
			log.Printf("ignoring GetAllStrings for %s", u.Username())
		} else {
			if u.cacheFollowersInMemory {
				u.followers = v.Strings()
			}
		}
	}

	lastOffset := -1
	{
		if u.has(cacheKeyFollowersByOffset) {
			keys, err := u.cache.FindKeys(u.cacheParts(cacheKeyFollowersByOffset)...)
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

	userInfos, userNamesToCache, followersErrs := u.client.AllFollowersParallel(u.username, fOpts...)

	users := make(chan *User)
	usernames := make(chan string)
	errs := make(chan error)

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
		for e := range followersErrs {
			errs <- e
		}

		close(users)
		close(usernames)
		close(errs)
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

	go func() {
		// Cache all newly-read users in memory.
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
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for followers of %s", u.Username())
		}
		lenBefore := len(u.followers)
		u.cachedFollowers(fOpts...)

		if lenBefore != len(u.followers) {
			go func() {
				if err := u.cacheBytes(u.followers, cacheKeyFollowers); err != nil {
					log.Printf("error caching followers for %s: %v", u.Username(), err)
				}
			}()
		}

		var res []*User
		for _, f := range u.followers {
			res = append(res, u.MakeUser(f))
		}
		return res, nil
	}

	var res []*User
	var followers []string
	if err := u.client.AllFollowers(u.username, func(offset int, userInfos api.UserInfos) error {
		for _, ui := range userInfos {
			res = append(res, u.MakeUser(ui.Username))
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

func (u *User) Following(ctx context.Context, fOpts ...UserFollowingOption) (chan *User, chan error) {
	opts := MakeUserFollowingOptions(fOpts...)
	convertOpts := func(opts UserFollowingOptions) []api.AllFollowingsOption {
		return []api.AllFollowingsOption{
			api.AllFollowingsIncl(opts.Incl()),
			api.AllFollowingsForce(opts.Force()),
			api.AllFollowingsMax(opts.Max()),
			api.AllFollowingsOffset(opts.Offset()),
			api.AllFollowingsStart(opts.Start()),
			api.AllFollowingsThreads(opts.Threads()),
		}
	}
	if opts.FromDisk() || u.opts().followingUsingDiskCache {
		return u.followingUsingDiskCache(ctx, convertOpts(opts)...)
	}
	return u.followingUsingDB(ctx, convertOpts(opts)...)
}

func (u *User) followingUsingDB(ctx context.Context, fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)
	threads := or.Int(opts.Threads(), defaultThreads)

	done, err := u.db.GetUserFollowingDone(ctx, u.Username())
	check.Err(err)
	if done {
		users := make(chan *User)

		followers, errors, err := u.db.GetFollowing(ctx, u.Username())
		check.Err(err)

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

		return users, errors
	}

	if opts.Force() {
		log.Printf("deleting following of %s", u.username)
		if err := u.db.DeleteFollowing(ctx, u.username); err != nil {
			todo.SkipErr("DeleteFollowing", err)
		}
	} else {
		lastOffset, err := u.db.GetUserMaxFollowingOffset(ctx, u.Username())
		check.Err(err)
		log.Printf("have last following offset: %d", lastOffset)
		if lastOffset > 0 {
			fOpts = append(fOpts, api.AllFollowingsStart(lastOffset))
		}
	}

	userInfos, userNamesToCache, errs := u.client.AllFollowingParallel(u.username, fOpts...)

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
			if err := u.db.SetFollowing(ctx, u.Username(), so.Offset, so.Strings); err != nil {
				log.Printf("SetFollowing: error caching following for %s, offset=%d: %v", u.Username(), so.Offset, err)
			}
		}
		// Mark that we are complete.
		if err := u.db.SetUserFollowingDone(ctx, u.Username(), true); err != nil {
			log.Printf("SetUserFollowingDone: error caching following for %s: %v", u.Username(), err)
		}
	}()

	// Cache all newly-read users in memory.
	go func() {
		var following []string
		for f := range usernames {
			if u.cacheFollowersInMemory {
				following = append(following, f)
			}
		}
		if u.cacheFollowersInMemory {
			u.following = append(u.following, following...)
		}
	}()

	return users, errs
}

func (u *User) followingUsingDiskCache(ctx context.Context, fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	// Following are either in "following" (legacy) or sharded into "followingByOffset"
	if u.has(cacheKeyFollowing) {
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for following of %s", u.Username())
		}
		return u.cachedFollowing(fOpts...)
	} else if u.has(cacheKeyFollowingDone) && u.has(cacheKeyFollowingByOffset) {
		// If we've completely read users and sharded them, read from the sharded directory.
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for followingByOffset and followingDone of %s", u.Username())
		}
		return u.cachedFollowingByOffset(fOpts...)
	} else if u.has(cacheKeyFollowingByOffset) {
		// In the case we have partially read the following, populate the in-memory cache.
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for followingByOffset of %s", u.Username())
		}

		v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowingByOffset)...)
		if err != nil {
			log.Printf("ignoring GetAllStrings for %s", u.Username())
		} else {
			if u.cacheFollowingInMemory {
				u.following = v.Strings()
			}
		}
	}

	lastOffset := -1
	{
		if u.has(cacheKeyFollowingByOffset) {
			keys, err := u.cache.FindKeys(u.cacheParts(cacheKeyFollowingByOffset)...)
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

	userInfos, userNamesToCache, userErrors := u.client.AllFollowingParallel(u.username, fOpts...)

	users := make(chan *User)
	errs := make(chan error)
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
		for e := range userErrors {
			errs <- e
		}
		close(users)
		close(errs)
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
		if u.opts().verboseCacheHits {
			log.Printf("cache hit for following of %s", u.Username())
		}
		lenBefore := len(u.following)
		u.cachedFollowing(fOpts...)

		if lenBefore != len(u.following) {
			go func() {
				if err := u.cacheBytes(u.following, cacheKeyFollowing); err != nil {
					log.Printf("error caching following for %s: %v", u.Username(), err)
				}
			}()
		}

		var res []*User
		for _, f := range u.following {
			res = append(res, u.MakeUser(f))
		}
		return res, nil
	}

	var res []*User
	var following []string
	if err := u.client.AllFollowings(u.username, func(offset int, userInfos api.UserInfos) error {
		for _, ui := range userInfos {
			res = append(res, u.MakeUser(ui.Username))
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
	ok, err := u.cache.Has(u.cacheParts(key)...)
	if err != nil {
		log.Printf("has: ignoring error: %v", err)
		return false
	}
	return ok
}

func (u *User) cacheParts(key cacheKey) []string {
	return []string{"users", u.Username(), string(key)}
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
	return setBytes(u.cache, val, u.cacheParts(part)...)
}

func (u *User) cacheOffsetStrings(val []string, part cacheKey, offset int) error {
	parts := u.cacheParts(part)
	parts = append(parts, fmt.Sprintf("%d", offset))
	return setBytes(u.cache, val, parts...)
}

var (
	persistDisallow = sets.String([]string{"hectorfbara84"})
)

func (u *User) Persist(ctx context.Context, pOpts ...UserPersistOption) error {
	if persistDisallow[u.Username()] {
		return nil
	}

	opts := MakeUserPersistOptions(pOpts...)

	should := func(key cacheKey) bool {
		return opts.Force() || !u.has(key)
	}

	persistUserInfo := func(u *User) error {
		if should(cacheKeyUserInfo) {
			userInfo, err := u.UserInfo(ctx)
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
		if u.opts().verbosePersist {
			log.Printf("persisting followers of %s", u.username)
		}

		followers := make(chan *User)
		go func() {
			users, errs := u.Followers(ctx, UserFollowersMax(opts.Max()), UserFollowersThreads(opts.Threads()))
			parallel.WaitFor(func() {
				for u := range users {
					followers <- u
				}
			}, func() {
				for e := range errs {
					log.Printf("error: %v", e)
				}
			})
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
		if u.opts().verbosePersist {
			log.Printf("SKIP persisting followers of %s", u.username)
		}
	}

	if should(cacheKeyFollowing) {
		if u.opts().verbosePersist {
			log.Printf("persisting following of %s", u.username)
		}

		following := make(chan *User)
		go func() {
			users, errs := u.Following(ctx, UserFollowingMax(opts.Max()), UserFollowingThreads(opts.Threads()))
			parallel.WaitFor(func() {
				for u := range users {
					following <- u
				}
			}, func() {
				for e := range errs {
					log.Printf("error: %v", e)
				}
			})
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
		if u.opts().verbosePersist {
			log.Printf("SKIP persisting following of %s", u.username)
		}
	}
	if err := persistUserInfo(u); err != nil {
		return err
	}

	return nil
}

func (u *User) cachedFollowers(fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if u.cacheFollowersInMemory {
			existing = u.followers
		}
		if len(existing) == 0 {
			v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowersByOffset)...)
			if err != nil {
				errs <- err
				return
			}
			existing = v.Strings()
		}
		for _, f := range existing {
			followers <- f
		}
		if u.cacheFollowersInMemory {
			u.followers = existing
		}
		close(followers)
		close(errs)
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

func (u *User) cachedFollowersByOffset(fOpts ...api.AllFollowersOption) (chan *User, chan error) {
	opts := api.MakeAllFollowersOptions(fOpts...)

	followers := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if u.cacheFollowersInMemory {
			existing = u.followers
		}
		if len(existing) == 0 {
			v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowersByOffset)...)
			if err != nil {
				errs <- err
				return
			}
			// TODO: Transfer shard
			existing = v.Strings()
		}
		for _, f := range existing {
			followers <- f
		}
		if u.cacheFollowersInMemory {
			u.followers = existing
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

func (u *User) cachedFollowing(fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if u.cacheFollowingInMemory {
			existing = u.following
		}
		if len(existing) == 0 {
			v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowingByOffset)...)
			if err != nil {
				errs <- err
				return
			}
			existing = v.Strings()
		}
		for _, f := range existing {
			following <- f
		}
		if u.cacheFollowingInMemory {
			u.following = existing
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

func (u *User) cachedFollowingByOffset(fOpts ...api.AllFollowingsOption) (chan *User, chan error) {
	opts := api.MakeAllFollowingsOptions(fOpts...)

	following := make(chan string)
	users := make(chan *User)
	errs := make(chan error)
	threads := or.Int(opts.Threads(), defaultThreads)

	go func() {
		var existing []string
		if u.cacheFollowingInMemory {
			existing = u.following
		}
		if len(existing) == 0 {
			v, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowingByOffset)...)
			if err != nil {
				errs <- err
				return
			}
			existing = v.Strings()
		}
		for _, f := range existing {
			following <- f
		}
		if u.cacheFollowingInMemory {
			u.following = existing
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

func (u *User) PersistInDB(ctx context.Context, pOpts ...PersistInDBOption) error {
	opts := MakePersistInDBOptions(pOpts...)
	threads := or.Int(opts.Threads(), defaultThreads)

	if u.has(cacheKeyUserInfo) {
		log.Printf("transering userInfo")
		userInfo, err := u.userInfoUsingDiskCache(ctx)
		if err != nil {
			return err
		}
		if err := u.db.SetUserInfo(ctx, u.Username(), userInfo); err != nil {
			return err
		}
	}

	type usersAndOffset struct {
		users  []string
		offset string
	}

	createUsersByOffset := func(ss ShardedStrings) chan usersAndOffset {
		usersByOffsets := map[string][]string{}
		for _, x := range ss {
			user, offset := x.Val, x.Dir
			users := usersByOffsets[offset]
			users = append(users, user)
			usersByOffsets[offset] = users
		}
		ch := make(chan usersAndOffset)
		go func() {
			for offset, users := range usersByOffsets {
				ch <- usersAndOffset{offset: offset, users: users}
			}
			close(ch)
		}()
		return ch
	}

	var wg sync.WaitGroup

	if u.has(cacheKeyFollowersByOffset) {
		log.Printf("transering followers with %d threads", threads)
		ss, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowersByOffset)...)
		if err != nil {
			return err
		}
		ch := createUsersByOffset(ss)
		go func() {
			for i := 0; i < threads; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for uao := range ch {
						if err := u.db.SetFollowers(ctx, u.Username(), must.Atoi(uao.offset), uao.users); err != nil {
							log.Printf("TODO SetFollowers; %v", err)
						}
					}
				}()
			}
		}()
	}

	if u.has(cacheKeyFollowingByOffset) {
		log.Printf("transering following with %d threads", threads)
		ss, err := u.cache.GetAllStrings(u.cacheParts(cacheKeyFollowingByOffset)...)
		if err != nil {
			return err
		}
		ch := createUsersByOffset(ss)
		go func() {
			for i := 0; i < threads; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for uao := range ch {
						if err := u.db.SetFollowing(ctx, u.Username(), must.Atoi(uao.offset), uao.users); err != nil {
							log.Printf("TODO SetFollowing; %v", err)
						}
					}
				}()
			}
		}()
	}

	wg.Wait()

	return nil
}

// func (u *User) Posts(ctx context.Context, fOpts ...UserPostsOption) (chan *User, chan error) {
// 	opts := MakeUserPostsOptions(fOpts...)
// 	convertOpts := func(opts UserPostsOptions) []api.AllPostsOption {
// 		return []api.AllPostsOption{
// 			api.AllPostsIncl(opts.Incl()),
// 			api.AllPostsForce(opts.Force()),
// 			api.AllPostsMax(opts.Max()),
// 			api.AllPostsOffset(opts.Offset()),
// 			api.AllPostsStart(opts.Start()),
// 			api.AllPostsThreads(opts.Threads()),
// 		}
// 	}
// 	return u.postsUsingDB(ctx, convertOpts(opts)...)
// }

// func (u *User) postsUsingDB(ctx context.Context, fOpts ...api.AllPostsOption) (chan *User, chan error) {
// 	opts := api.MakeAllPostsOptions(fOpts...)
// 	threads := or.Int(opts.Threads(), defaultThreads)

// 	users := make(chan *User)
// 	usernames := make(chan string)

// 	posts, errs := u.client.AllPosts(u.username, fOpts...)

// 	go func() {
// 		// Transfer the partially-populated in-memory cache to the result.
// 		if u.cacheFollowersInMemory {
// 			for _, f := range u.followers {
// 				follower := u.MakeUser(f)
// 				users <- follower
// 			}
// 		}
// 		// Transfer the newly-read users to the result.
// 		for postInfo := range posts {
// 			follower := u.MakeUser(postInfo.Username)
// 			follower.userInfo = postInfo
// 			users <- follower
// 			usernames <- postInfo.Username
// 		}
// 		close(users)
// 		close(usernames)
// 	}()

// 	go func() {

// 		// Cache by shard.
// 		for so := range userNamesToCache {
// 			if err := u.db.SetFollowers(ctx, u.Username(), so.Offset, so.Strings); err != nil {
// 				log.Printf("SetFollowers: error caching followers for %s, offset=%d: %v", u.Username(), so.Offset, err)
// 			}
// 		}
// 		// Mark that we are complete.
// 		if err := u.db.SetUserFollowersDone(ctx, u.Username(), true); err != nil {
// 			log.Printf("SetUserFollowersDone: error caching followers for %s: %v", u.Username(), err)
// 		}
// 	}()

// 	// // Cache all newly-read users in memory.
// 	// go func() {
// 	// 	var posts []api.PostInfo
// 	// 	for f := range usernames {
// 	// 		if u.cachePostsInMemory {
// 	// 			followers = append(followers, f)
// 	// 		}
// 	// 	}
// 	// 	if u.cachePostsInMemory {
// 	// 		u.posts = append(u.posts, posts...)
// 	// 	}
// 	// }()

// 	return users, errs
// }
