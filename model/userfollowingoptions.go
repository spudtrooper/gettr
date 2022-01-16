package model

// genopts --opt_type=UserFollowingOption --prefix=UserFollowing --outfile=model/userfollowingoptions.go 'offset:int' 'max:int' 'incl:[]string' 'start:int' 'threads:int' 'fromDisk'

type UserFollowingOption func(*userFollowingOptionImpl)

type UserFollowingOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	FromDisk() bool
}

func UserFollowingOffset(offset int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.offset = offset
	}
}

func UserFollowingMax(max int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.max = max
	}
}

func UserFollowingIncl(incl []string) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.incl = incl
	}
}

func UserFollowingStart(start int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.start = start
	}
}

func UserFollowingThreads(threads int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.threads = threads
	}
}

func UserFollowingFromDisk(fromDisk bool) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.fromDisk = fromDisk
	}
}

type userFollowingOptionImpl struct {
	offset   int
	max      int
	incl     []string
	start    int
	threads  int
	fromDisk bool
}

func (u *userFollowingOptionImpl) Offset() int    { return u.offset }
func (u *userFollowingOptionImpl) Max() int       { return u.max }
func (u *userFollowingOptionImpl) Incl() []string { return u.incl }
func (u *userFollowingOptionImpl) Start() int     { return u.start }
func (u *userFollowingOptionImpl) Threads() int   { return u.threads }
func (u *userFollowingOptionImpl) FromDisk() bool { return u.fromDisk }

func makeUserFollowingOptionImpl(opts ...UserFollowingOption) *userFollowingOptionImpl {
	res := &userFollowingOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUserFollowingOptions(opts ...UserFollowingOption) UserFollowingOptions {
	return makeUserFollowingOptionImpl(opts...)
}
