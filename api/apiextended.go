package api

import (
	"math"
	"sync"

	"github.com/spudtrooper/goutil/or"
)

// type Extended has functions the use the Core API to produce multiple results.
type Extended struct {
	*Core
}

func MakeExtended(c *Core) *Extended {
	return &Extended{c}
}

func (c *Extended) GetAllFollowings(username string, fOpts ...FollowingsOption) (UserInfos, error) {
	opts := MakeFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	var res UserInfos
	for offset := 0; ; offset += max {
		followings, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
		if err != nil {
			return nil, err
		}
		if len(followings) == 0 {
			break
		}
		res = append(res, followings...)
	}
	return res, nil
}

func (c *Extended) AllFollowings(username string, f func(offset int, us UserInfos) error, fOpts ...AllFollowingsOption) error {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	for offset := start; ; offset += max {
		followings, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(offset, followings); err != nil {
			return err
		}
	}
	return nil
}

func (c *Extended) AllFollowingsParallel(username string, fOpts ...AllFollowingsOption) (chan UserInfo, chan error) {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	userInfos := make(chan UserInfo)
	offsets := make(chan int)
	errs := make(chan error)

	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for offset := range offsets {
					fs, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					for _, u := range fs {
						userInfos <- u
					}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(errs)
	}()

	return userInfos, errs
}

func (c *Extended) AllFollowersParallel(username string, fOpts ...AllFollowersOption) (chan UserInfo, chan OffsetStrings, chan error) {
	opts := MakeAllFollowersOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	offsets := make(chan int)
	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	userInfos := make(chan UserInfo)
	userNames := make(chan OffsetStrings)
	errs := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col := makeClientStatsCollector("AllFollowersParallel")
				for offset := range offsets {
					fs, err := c.GetFollowers(username, FollowersOffset(offset), FollowersMax(max))
					if *clientStats {
						col.RecordAndPrint()
					}
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					var us []string
					for _, u := range fs {
						userInfos <- u
						us = append(us, u.Username)
					}
					userNames <- OffsetStrings{Strings: us, Offset: offset}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(userNames)
		close(errs)
	}()

	return userInfos, userNames, errs
}

func (c *Extended) AllFollowers(username string, f func(offset int, userInfos UserInfos) error, fOpts ...AllFollowersOption) error {
	opts := MakeAllFollowersOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	for offset := start; ; offset += max {
		followings, err := c.GetFollowers(username, FollowersOffset(offset), FollowersMax(max))
		if err != nil {
			return err
		}
		if len(followings) == 0 {
			break
		}
		if err := f(offset, followings); err != nil {
			return err
		}
	}
	return nil
}

func (c *Extended) AllFollowingParallel(username string, fOpts ...AllFollowingsOption) (chan UserInfo, chan OffsetStrings, chan error) {
	opts := MakeAllFollowingsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	userInfos := make(chan UserInfo)
	userNames := make(chan OffsetStrings)
	offsets := make(chan int)
	errs := make(chan error)

	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col := makeClientStatsCollector("AllFollowingParallel")
				for offset := range offsets {
					fs, err := c.GetFollowings(username, FollowingsOffset(offset), FollowingsMax(max))
					if *clientStats {
						col.RecordAndPrint()
					}
					if err != nil {
						errs <- err
						break
					}
					if len(fs) == 0 {
						break
					}
					var us []string
					for _, u := range fs {
						userInfos <- u
						us = append(us, u.Username)
					}
					userNames <- OffsetStrings{Strings: us, Offset: offset}
				}
			}()
		}
		wg.Wait()
		close(userInfos)
		close(userNames)
		close(errs)
	}()

	return userInfos, userNames, errs
}

func (c *Extended) AllPosts(username string, fOpts ...AllPostsOption) (chan OffsetPosts, chan error) {
	opts := MakeAllPostsOptions(fOpts...)
	max := or.Int(opts.Max(), defaultMax)
	start := or.Int(opts.Start(), defaultStart)
	threads := or.Int(opts.Threads(), defaultThreads)

	offsetPosts := make(chan OffsetPosts)
	offsets := make(chan int)
	errs := make(chan error)

	go func() {
		for offset := start; offset < math.MaxInt; offset += max {
			offsets <- offset
		}
		close(offsets)
	}()

	go func() {
		var wg sync.WaitGroup
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				col := makeClientStatsCollector("AllPosts")
				for offset := range offsets {
					posts, err := c.GetPosts(username, PostsOffset(offset), PostsMax(max))
					if *clientStats {
						col.RecordAndPrint()
					}
					if err != nil {
						errs <- err
						break
					}
					if len(posts) == 0 {
						break
					}
					offsetPosts <- OffsetPosts{Posts: posts, Offset: offset}
				}
			}()
		}
		wg.Wait()
		close(offsetPosts)
		close(errs)
	}()

	return offsetPosts, errs
}
