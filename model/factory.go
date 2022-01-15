package model

import (
	"context"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/check"
)

type Factory interface {
	MakeUser(username string) *User
	Cache() Cache
	Client() *api.Client
}

type factory struct {
	cache       Cache
	client      *api.Client
	db          *DB
	userCacheMu sync.Mutex
	userCache   map[string]*User
}

func MakeFactory(cache Cache, client *api.Client) Factory {
	db, err := MakeDB(context.TODO())
	check.Err(err)
	userCache := map[string]*User{}
	return &factory{cache: cache, client: client, db: db, userCache: userCache}
}

func MakeFactoryFromFlags() (Factory, error) {
	client, err := api.MakeClientFromFlags()
	if err != nil {
		return nil, err
	}
	cache, err := MakeCacheFromFlags()
	if err != nil {
		return nil, err
	}
	factory := MakeFactory(cache, client)
	return factory, nil
}

func (f *factory) Cache() Cache        { return f.cache }
func (f *factory) Client() *api.Client { return f.client }

func (f *factory) MakeUser(username string) *User {
	f.userCacheMu.Lock()
	defer f.userCacheMu.Unlock()
	var res *User
	if user, ok := f.userCache[username]; ok && user != nil {
		res = user
	}
	if res == nil {
		res = &User{username: username, factory: f}
		f.userCache[username] = res
	}
	return res
}
