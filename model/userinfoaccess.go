package model

import (
	"time"

	"github.com/pkg/errors"
)

func (u *User) BGImg(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("BGImg: no userInfo for: %s", u.username)
	}
	return ui.BGImg, nil
}

func (u *User) Desc(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Desc: no userInfo for: %s", u.username)
	}
	return ui.Desc, nil
}

func (u *User) ICO(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("ICO: no userInfo for: %s", u.username)
	}
	return ui.ICO, nil
}

func (u *User) Infl(uOpts ...UserInfoOption) (int, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return 0, err
	}
	if ui.Username == "" {
		return 0, errors.Errorf("Infl: no userInfo for: %s", u.username)
	}
	return ui.Infl, nil
}

func (u *User) Lang(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Lang: no userInfo for: %s", u.username)
	}
	return ui.Lang, nil
}

func (u *User) OUsername(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("OUsername: no userInfo for: %s", u.username)
	}
	return ui.OUsername, nil
}

func (u *User) Website(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Website: no userInfo for: %s", u.username)
	}
	return ui.Website, nil
}

func (u *User) TweetFollowing(uOpts ...UserInfoOption) (int, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return 0, err
	}
	if ui.Username == "" {
		return 0, errors.Errorf("TweetFollowing: no userInfo for: %s", u.username)
	}
	return ui.TwitterFollowing(), nil
}

func (u *User) TweetFollowers(uOpts ...UserInfoOption) (int, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return 0, err
	}
	if ui.Username == "" {
		return 0, errors.Errorf("TweetFollowers: no userInfo for: %s", u.username)
	}
	return ui.TwitterFollowers(), nil
}

func (u *User) GetFollowers(uOpts ...UserInfoOption) (int, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return 0, err
	}
	if ui.Username == "" {
		return 0, errors.Errorf("GetFollowers: no userInfo for: %s", u.username)
	}
	return ui.Followers(), nil
}

func (u *User) GetFollowing(uOpts ...UserInfoOption) (int, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return 0, err
	}
	if ui.Username == "" {
		return 0, errors.Errorf("GetFollowing: no userInfo for: %s", u.username)
	}
	return ui.Following(), nil
}

func (u *User) CDate(uOpts ...UserInfoOption) (time.Time, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return time.Time{}, err
	}
	if ui.Username == "" {
		return time.Time{}, errors.Errorf("CDate: no userInfo for: %s", u.username)
	}
	return ui.CDate.Time()
}

func (u *User) UDate(uOpts ...UserInfoOption) (time.Time, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return time.Time{}, err
	}
	if ui.Username == "" {
		return time.Time{}, errors.Errorf("UDate: no userInfo for: %s", u.username)
	}
	return ui.UDate.Time()
}

func (u *User) Type(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Type: no userInfo for: %s", u.username)
	}
	return ui.Type, nil
}

func (u *User) ID(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("ID: no userInfo for: %s", u.username)
	}
	return ui.ID, nil
}

func (u *User) Nickname(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Nickname: no userInfo for: %s", u.username)
	}
	return ui.Nickname, nil
}

func (u *User) Status(uOpts ...UserInfoOption) (string, error) {
	ui, err := u.UserInfo(uOpts...)
	if err != nil {
		return "", err
	}
	if ui.Username == "" {
		return "", errors.Errorf("Status: no userInfo for: %s", u.username)
	}
	return ui.Status, nil
}
