// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=AllFollowings --outfile=allfollowingsoptions.go "offset:int" "max:int" "incl:[]string" "start:int" "threads:int" "force"

type AllFollowingsOption func(*allFollowingsOptionImpl)

type AllFollowingsOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	Force() bool
}

func AllFollowingsOffset(offset int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.offset = offset
	}
}
func AllFollowingsOffsetFlag(offset *int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.offset = *offset
	}
}

func AllFollowingsMax(max int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.max = max
	}
}
func AllFollowingsMaxFlag(max *int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.max = *max
	}
}

func AllFollowingsIncl(incl []string) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.incl = incl
	}
}
func AllFollowingsInclFlag(incl *[]string) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.incl = *incl
	}
}

func AllFollowingsStart(start int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.start = start
	}
}
func AllFollowingsStartFlag(start *int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.start = *start
	}
}

func AllFollowingsThreads(threads int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.threads = threads
	}
}
func AllFollowingsThreadsFlag(threads *int) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.threads = *threads
	}
}

func AllFollowingsForce(force bool) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.force = force
	}
}
func AllFollowingsForceFlag(force *bool) AllFollowingsOption {
	return func(opts *allFollowingsOptionImpl) {
		opts.force = *force
	}
}

type allFollowingsOptionImpl struct {
	offset  int
	max     int
	incl    []string
	start   int
	threads int
	force   bool
}

func (a *allFollowingsOptionImpl) Offset() int    { return a.offset }
func (a *allFollowingsOptionImpl) Max() int       { return a.max }
func (a *allFollowingsOptionImpl) Incl() []string { return a.incl }
func (a *allFollowingsOptionImpl) Start() int     { return a.start }
func (a *allFollowingsOptionImpl) Threads() int   { return a.threads }
func (a *allFollowingsOptionImpl) Force() bool    { return a.force }

func makeAllFollowingsOptionImpl(opts ...AllFollowingsOption) *allFollowingsOptionImpl {
	res := &allFollowingsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeAllFollowingsOptions(opts ...AllFollowingsOption) AllFollowingsOptions {
	return makeAllFollowingsOptionImpl(opts...)
}
