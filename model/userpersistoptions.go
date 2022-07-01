// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package model

//go:generate genopts --prefix=UserPersist --outfile=userpersistoptions.go "max:int" "threads:int" "force:bool"

type UserPersistOption func(*userPersistOptionImpl)

type UserPersistOptions interface {
	Max() int
	Threads() int
	Force() bool
}

func UserPersistMax(max int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.max = max
	}
}
func UserPersistMaxFlag(max *int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.max = *max
	}
}

func UserPersistThreads(threads int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.threads = threads
	}
}
func UserPersistThreadsFlag(threads *int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.threads = *threads
	}
}

func UserPersistForce(force bool) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.force = force
	}
}
func UserPersistForceFlag(force *bool) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.force = *force
	}
}

type userPersistOptionImpl struct {
	max     int
	threads int
	force   bool
}

func (u *userPersistOptionImpl) Max() int     { return u.max }
func (u *userPersistOptionImpl) Threads() int { return u.threads }
func (u *userPersistOptionImpl) Force() bool  { return u.force }

func makeUserPersistOptionImpl(opts ...UserPersistOption) *userPersistOptionImpl {
	res := &userPersistOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUserPersistOptions(opts ...UserPersistOption) UserPersistOptions {
	return makeUserPersistOptionImpl(opts...)
}
