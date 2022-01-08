package model

import (
	"encoding/json"

	"github.com/spudtrooper/gettr/api"
)

type User interface {
	Username() string
	Followers(fOpts ...api.AllFollowersOption) (chan User, chan error)
	Following(fOpts ...api.AllFollowingsOption) (chan User, chan error)
	UserInfo() (api.UserInfo, error)
	Persist(pOpts ...UserPersistOption) error
}

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
	users := make(chan User)

	userInfos, errs := u.client.AllFollowersParallel(u.username, fOpts...)

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
	users := make(chan User)

	userInfos, errs := u.client.AllFollowingsParallel(u.username, fOpts...)

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

func (u *user) Persist(pOpts ...UserPersistOption) error {
	opts := MakeUserPersistOptions(pOpts...)

	followers := make(chan User)
	following := make(chan User)

	cache := u.cache

	setBytes := func(val interface{}, parts ...string) error {
		bytes, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if err := cache.SetBytes(bytes, parts...); err != nil {
			return err
		}
		return nil
	}

	go func() {
		users, _ := u.Followers(api.AllFollowersMax(opts.Max()), api.AllFollowersMax(opts.Threads()))
		for u := range users {
			followers <- u
		}
		close(followers)
	}()
	go func() {
		users, _ := u.Following(api.AllFollowingsMax(opts.Max()), api.AllFollowingsMax(opts.Threads()))
		for u := range users {
			following <- u
		}
		close(following)
	}()

	{
		var arr []string
		for u := range followers {
			arr = append(arr, u.Username())
			userInfo, err := u.UserInfo()
			if err != nil {
				return err
			}
			if err := setBytes(userInfo, "users", u.Username(), "userInfo"); err != nil {
				return err
			}
		}
		if err := setBytes(arr, "users", u.username, "followers"); err != nil {
			return err
		}
	}
	{
		var arr []string
		for u := range following {
			arr = append(arr, u.Username())
			userInfo, err := u.UserInfo()
			if err != nil {
				return err
			}
			if err := setBytes(userInfo, "users", u.Username(), "userInfo"); err != nil {
				return err
			}
		}
		if err := setBytes(arr, "users", u.username, "following"); err != nil {
			return err
		}
	}
	if err := setBytes(u.userInfo, "users", u.username, "userInfo"); err != nil {
		return err
	}

	return nil
}
