package model

// genopts --opt_type=UserFollowersOption --prefix=UserFollowers --outfile=model/userfollowersoptions.go 'offset:int' 'max:int' 'incl:[]string' 'start:int' 'threads:int' 'fromDisk'

type UserFollowersOption func(*userFollowersOptionImpl)

type UserFollowersOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	FromDisk() bool
}

func UserFollowersOffset(offset int) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.offset = offset
	}
}

func UserFollowersMax(max int) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.max = max
	}
}

func UserFollowersIncl(incl []string) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.incl = incl
	}
}

func UserFollowersStart(start int) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.start = start
	}
}

func UserFollowersThreads(threads int) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.threads = threads
	}
}

func UserFollowersFromDisk(fromDisk bool) UserFollowersOption {
	return func(opts *userFollowersOptionImpl) {
		opts.fromDisk = fromDisk
	}
}

type userFollowersOptionImpl struct {
	offset   int
	max      int
	incl     []string
	start    int
	threads  int
	fromDisk bool
}

func (u *userFollowersOptionImpl) Offset() int    { return u.offset }
func (u *userFollowersOptionImpl) Max() int       { return u.max }
func (u *userFollowersOptionImpl) Incl() []string { return u.incl }
func (u *userFollowersOptionImpl) Start() int     { return u.start }
func (u *userFollowersOptionImpl) Threads() int   { return u.threads }
func (u *userFollowersOptionImpl) FromDisk() bool { return u.fromDisk }

func makeUserFollowersOptionImpl(opts ...UserFollowersOption) *userFollowersOptionImpl {
	res := &userFollowersOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUserFollowersOptions(opts ...UserFollowersOption) UserFollowersOptions {
	return makeUserFollowersOptionImpl(opts...)
}
