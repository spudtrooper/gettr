// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=AllFollowers --outfile=allfollowersoptions.go "offset:int" "max:int" "incl:[]string" "start:int" "threads:int" "force"

type AllFollowersOption func(*allFollowersOptionImpl)

type AllFollowersOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	Force() bool
}

func AllFollowersOffset(offset int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.offset = offset
	}
}
func AllFollowersOffsetFlag(offset *int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.offset = *offset
	}
}

func AllFollowersMax(max int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.max = max
	}
}
func AllFollowersMaxFlag(max *int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.max = *max
	}
}

func AllFollowersIncl(incl []string) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.incl = incl
	}
}
func AllFollowersInclFlag(incl *[]string) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.incl = *incl
	}
}

func AllFollowersStart(start int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.start = start
	}
}
func AllFollowersStartFlag(start *int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.start = *start
	}
}

func AllFollowersThreads(threads int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.threads = threads
	}
}
func AllFollowersThreadsFlag(threads *int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.threads = *threads
	}
}

func AllFollowersForce(force bool) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.force = force
	}
}
func AllFollowersForceFlag(force *bool) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.force = *force
	}
}

type allFollowersOptionImpl struct {
	offset  int
	max     int
	incl    []string
	start   int
	threads int
	force   bool
}

func (a *allFollowersOptionImpl) Offset() int    { return a.offset }
func (a *allFollowersOptionImpl) Max() int       { return a.max }
func (a *allFollowersOptionImpl) Incl() []string { return a.incl }
func (a *allFollowersOptionImpl) Start() int     { return a.start }
func (a *allFollowersOptionImpl) Threads() int   { return a.threads }
func (a *allFollowersOptionImpl) Force() bool    { return a.force }

func makeAllFollowersOptionImpl(opts ...AllFollowersOption) *allFollowersOptionImpl {
	res := &allFollowersOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeAllFollowersOptions(opts ...AllFollowersOption) AllFollowersOptions {
	return makeAllFollowersOptionImpl(opts...)
}
