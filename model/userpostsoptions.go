// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package model

//go:generate genopts --prefix=UserPosts --outfile=userpostsoptions.go "offset:int" "max:int" "incl:[]string" "start:int" "threads:int" "fromDisk" "force"

type UserPostsOption func(*userPostsOptionImpl)

type UserPostsOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	FromDisk() bool
	Force() bool
}

func UserPostsOffset(offset int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.offset = offset
	}
}
func UserPostsOffsetFlag(offset *int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.offset = *offset
	}
}

func UserPostsMax(max int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.max = max
	}
}
func UserPostsMaxFlag(max *int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.max = *max
	}
}

func UserPostsIncl(incl []string) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.incl = incl
	}
}
func UserPostsInclFlag(incl *[]string) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.incl = *incl
	}
}

func UserPostsStart(start int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.start = start
	}
}
func UserPostsStartFlag(start *int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.start = *start
	}
}

func UserPostsThreads(threads int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.threads = threads
	}
}
func UserPostsThreadsFlag(threads *int) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.threads = *threads
	}
}

func UserPostsFromDisk(fromDisk bool) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.fromDisk = fromDisk
	}
}
func UserPostsFromDiskFlag(fromDisk *bool) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.fromDisk = *fromDisk
	}
}

func UserPostsForce(force bool) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.force = force
	}
}
func UserPostsForceFlag(force *bool) UserPostsOption {
	return func(opts *userPostsOptionImpl) {
		opts.force = *force
	}
}

type userPostsOptionImpl struct {
	offset   int
	max      int
	incl     []string
	start    int
	threads  int
	fromDisk bool
	force    bool
}

func (u *userPostsOptionImpl) Offset() int    { return u.offset }
func (u *userPostsOptionImpl) Max() int       { return u.max }
func (u *userPostsOptionImpl) Incl() []string { return u.incl }
func (u *userPostsOptionImpl) Start() int     { return u.start }
func (u *userPostsOptionImpl) Threads() int   { return u.threads }
func (u *userPostsOptionImpl) FromDisk() bool { return u.fromDisk }
func (u *userPostsOptionImpl) Force() bool    { return u.force }

func makeUserPostsOptionImpl(opts ...UserPostsOption) *userPostsOptionImpl {
	res := &userPostsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUserPostsOptions(opts ...UserPostsOption) UserPostsOptions {
	return makeUserPostsOptionImpl(opts...)
}
