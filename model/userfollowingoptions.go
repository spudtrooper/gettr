// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package model

//go:generate genopts --prefix=UserFollowing --outfile=userfollowingoptions.go "offset:int" "max:int" "incl:[]string" "start:int" "threads:int" "fromDisk" "force"

type UserFollowingOption func(*userFollowingOptionImpl)

type UserFollowingOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	FromDisk() bool
	Force() bool
}

func UserFollowingOffset(offset int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.offset = offset
	}
}
func UserFollowingOffsetFlag(offset *int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.offset = *offset
	}
}

func UserFollowingMax(max int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.max = max
	}
}
func UserFollowingMaxFlag(max *int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.max = *max
	}
}

func UserFollowingIncl(incl []string) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.incl = incl
	}
}
func UserFollowingInclFlag(incl *[]string) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.incl = *incl
	}
}

func UserFollowingStart(start int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.start = start
	}
}
func UserFollowingStartFlag(start *int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.start = *start
	}
}

func UserFollowingThreads(threads int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.threads = threads
	}
}
func UserFollowingThreadsFlag(threads *int) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.threads = *threads
	}
}

func UserFollowingFromDisk(fromDisk bool) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.fromDisk = fromDisk
	}
}
func UserFollowingFromDiskFlag(fromDisk *bool) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.fromDisk = *fromDisk
	}
}

func UserFollowingForce(force bool) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.force = force
	}
}
func UserFollowingForceFlag(force *bool) UserFollowingOption {
	return func(opts *userFollowingOptionImpl) {
		opts.force = *force
	}
}

type userFollowingOptionImpl struct {
	offset   int
	max      int
	incl     []string
	start    int
	threads  int
	fromDisk bool
	force    bool
}

func (u *userFollowingOptionImpl) Offset() int    { return u.offset }
func (u *userFollowingOptionImpl) Max() int       { return u.max }
func (u *userFollowingOptionImpl) Incl() []string { return u.incl }
func (u *userFollowingOptionImpl) Start() int     { return u.start }
func (u *userFollowingOptionImpl) Threads() int   { return u.threads }
func (u *userFollowingOptionImpl) FromDisk() bool { return u.fromDisk }
func (u *userFollowingOptionImpl) Force() bool    { return u.force }

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
