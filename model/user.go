package model

import (
	"github.com/spudtrooper/gettr/api"
)

type User interface {
	Username() string
	Followers(fOpts ...api.AllFollowersOption) (chan User, chan error)
	Following(fOpts ...api.AllFollowingsOption) (chan User, chan error)
	UserInfo() (api.UserInfo, error)
	Persist(pOpts ...UserPersistOption) error
}
