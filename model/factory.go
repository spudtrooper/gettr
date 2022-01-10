package model

import (
	"github.com/spudtrooper/gettr/api"
)

type Factory interface {
	MakeUser(username string) *User
}

type factory struct {
	cache  Cache
	client *api.Client
}

func MakeFactory(cache Cache, client *api.Client) Factory {
	return &factory{cache, client}
}

func (f *factory) MakeUser(username string) *User {
	return &User{username: username, factory: f}
}
