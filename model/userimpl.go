package model

import (
	"encoding/json"
	"log"

	"github.com/spudtrooper/gettr/api"
)

type user struct {
	*factory
	username string
	userInfo api.UserInfo
}

func (u *user) Username() string { return u.username }

func (u *user) UserInfo() (api.UserInfo, error) {
	if u.userInfo.Username == "" {
		uinfo, err := u.client.GetUserInfo(u.username)
		if err != nil {
			return api.UserInfo{}, err
		}
		u.userInfo = uinfo
	}
	return u.userInfo, nil
}

func (u *user) Followers(fOpts ...api.AllFollowersOption) (chan User, chan error) {

	userInfos, errs := u.client.AllFollowersParallel(u.username, fOpts...)

	users := make(chan User)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.(*user).userInfo = userInfo
			users <- follower
		}
		close(users)
	}()

	return users, errs
}

func (u *user) Following(fOpts ...api.AllFollowingsOption) (chan User, chan error) {
	userInfos, errs := u.client.AllFollowingsParallel(u.username, fOpts...)

	users := make(chan User)

	go func() {
		for userInfo := range userInfos {
			follower := u.MakeUser(userInfo.Username)
			follower.(*user).userInfo = userInfo
			users <- follower
		}
		close(users)
	}()

	return users, errs
}

func (u *user) has(parts ...string) bool {
	ok, err := u.cache.Has(parts...)
	return !ok || err != nil
}

func (u *user) Persist(pOpts ...UserPersistOption) error {
	opts := MakeUserPersistOptions(pOpts...)

	setBytes := func(val interface{}, parts ...string) error {
		bytes, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if err := u.cache.SetBytes(bytes, parts...); err != nil {
			return err
		}
		return nil
	}

	should := func(parts ...string) bool {
		return opts.Force() || !u.has(parts...)
	}

	if should("users", u.username, "followers") {
		log.Printf("persisting followers of %s", u.username)

		followers := make(chan User)
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
			userInfo, err := u.UserInfo()
			if err != nil {
				return err
			}
			if err := setBytes(userInfo, "users", u.Username(), "userInfo"); err != nil {
				return err
			}
		}
		log.Printf("persisted %d followers of %s", len(usernames), u.username)
		if err := setBytes(usernames, "users", u.username, "followers"); err != nil {
			return err
		}
	}

	if should("users", u.username, "following") {
		log.Printf("persisting following of %s", u.username)

		following := make(chan User)
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
			userInfo, err := u.UserInfo()
			if err != nil {
				return err
			}
			if err := setBytes(userInfo, "users", u.Username(), "userInfo"); err != nil {
				return err
			}
		}
		log.Printf("persisted %d following of %s", len(usernames), u.username)
		if err := setBytes(usernames, "users", u.username, "following"); err != nil {
			return err
		}
	}
	if should("users", u.username, "userInfo") {
		log.Printf("persisting userInfo of %s", u.username)
		if err := setBytes(u.userInfo, "users", u.username, "userInfo"); err != nil {
			return err
		}
	}

	return nil
}
