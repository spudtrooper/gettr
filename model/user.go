package model

import "github.com/spudtrooper/gettr/api"

type User interface {
	Username() string
	Followers(func(u User) error) error
	Following(func(u User) error) error
	UserInfo() (api.UserInfo, error)
}

type user struct {
	username string
	cache    Cache
	client   *api.Client
	userInfo api.UserInfo
}

func MakeUser(username string, cache Cache, client *api.Client) User {
	return wrapUser(username, cache, client)
}

func wrapUser(username string, cache Cache, client *api.Client) *user {
	return &user{
		username: username,
		cache:    cache,
		client:   client,
	}
}

func (u *user) Username() string {
	return u.username
}

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

func (u *user) Followers(process func(u User) error) error {
	if err := u.client.AllFollowers(u.username, func(offset int, us api.UserInfos) error {
		for _, userInfo := range us {
			follower := wrapUser(userInfo.Username, u.cache, u.client)
			follower.userInfo = userInfo
			if err := process(follower); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (u *user) Following(process func(u User) error) error {
	if err := u.client.AllFollowings(u.username, func(offset int, us api.UserInfos) error {
		for _, userInfo := range us {
			follower := wrapUser(userInfo.Username, u.cache, u.client)
			follower.userInfo = userInfo
			if err := process(follower); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
