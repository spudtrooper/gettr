package model

// genopts --opt_type=UserPersistOption --prefix=UserPersist --outfile=model/userpersistoptions.go 'max:int' 'threads:int'

type UserPersistOption func(*userPersistOptionImpl)

type UserPersistOptions interface {
	Max() int
	Threads() int
}

func UserPersistMax(max int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.max = max
	}
}

func UserPersistThreads(threads int) UserPersistOption {
	return func(opts *userPersistOptionImpl) {
		opts.threads = threads
	}
}

type userPersistOptionImpl struct {
	max     int
	threads int
}

func (u *userPersistOptionImpl) Max() int     { return u.max }
func (u *userPersistOptionImpl) Threads() int { return u.threads }

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
