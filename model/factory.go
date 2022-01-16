package model

import (
	"context"
	"sync"

	"github.com/spudtrooper/gettr/api"
	"github.com/spudtrooper/goutil/flags"
)

var (
	userInfoUsingDiskCache  = flags.Bool("user_info_using_disk_cache", "use disk cache for userInfo")
	followersUsingDiskCache = flags.Bool("followers_using_disk_cache", "use disk cache for followers")
	followingUsingDiskCache = flags.Bool("following_using_disk_cache", "use disk cache for followers")
	verboseCacheHits        = flags.Bool("verbose_cache_hits", "log cache hits verbosely")
	verbosePersist          = flags.Bool("verbose_persist", "log persisting verbosely")
)

type Factory interface {
	MakeUser(username string) *User
	Cache() Cache
	Client() *api.Extended
}

type factoryOptions struct {
	userInfoUsingDiskCache  bool
	followersUsingDiskCache bool
	followingUsingDiskCache bool
	verboseCacheHits        bool
	verbosePersist          bool
}

type factory struct {
	cache       Cache
	client      *api.Extended
	db          *DB
	userCacheMu sync.Mutex
	userCache   map[string]*User
	factoryOptions
}

func (f *factory) opts() factoryOptions { return f.factoryOptions }

func MakeFactory(ctx context.Context, cache Cache, client *api.Core) (Factory, error) {
	db, err := MakeDB(ctx)
	if err != nil {
		return nil, err
	}
	userCache := map[string]*User{}
	res := &factory{
		cache:     cache,
		client:    api.MakeExtended(client),
		db:        db,
		userCache: userCache,
		factoryOptions: factoryOptions{
			userInfoUsingDiskCache:  *userInfoUsingDiskCache,
			followersUsingDiskCache: *followersUsingDiskCache,
			followingUsingDiskCache: *followingUsingDiskCache,
			verboseCacheHits:        *verboseCacheHits,
			verbosePersist:          *verbosePersist,
		},
	}
	return res, nil
}

func MakeFactoryFromFlags(ctx context.Context) (Factory, error) {
	client, err := api.MakeClientFromFlags()
	if err != nil {
		return nil, err
	}
	cache, err := MakeCacheFromFlags()
	if err != nil {
		return nil, err
	}
	factory, err := MakeFactory(ctx, cache, client)
	if err != nil {
		return nil, err
	}
	return factory, nil
}

func (f *factory) Cache() Cache          { return f.cache }
func (f *factory) Client() *api.Extended { return f.client }

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
