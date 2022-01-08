package model

import (
	"github.com/spudtrooper/gettr/api"
)

type Factory interface {
	MakeUser(username string) User
	MakeCachedUser(username string) User
}

type factory struct {
	cache  Cache
	client *api.Client
}

func MakeFactory(cache Cache, client *api.Client) Factory {
	return &factory{cache, client}
}

func (f *factory) MakeUser(username string) User {
	return &user{username: username, factory: f}
}

func (f *factory) MakeCachedUser(username string) User {
	return &cachedUser{username: username, factory: f}
}
