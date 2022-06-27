// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=AllPosts --outfile=allpostsoptions.go "offset:int" "max:int" "incl:[]string" "start:int" "threads:int" "force"

type AllPostsOption func(*allPostsOptionImpl)

type AllPostsOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
	Threads() int
	Force() bool
}

func AllPostsOffset(offset int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.offset = offset
	}
}
func AllPostsOffsetFlag(offset *int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.offset = *offset
	}
}

func AllPostsMax(max int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.max = max
	}
}
func AllPostsMaxFlag(max *int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.max = *max
	}
}

func AllPostsIncl(incl []string) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.incl = incl
	}
}
func AllPostsInclFlag(incl *[]string) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.incl = *incl
	}
}

func AllPostsStart(start int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.start = start
	}
}
func AllPostsStartFlag(start *int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.start = *start
	}
}

func AllPostsThreads(threads int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.threads = threads
	}
}
func AllPostsThreadsFlag(threads *int) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.threads = *threads
	}
}

func AllPostsForce(force bool) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.force = force
	}
}
func AllPostsForceFlag(force *bool) AllPostsOption {
	return func(opts *allPostsOptionImpl) {
		opts.force = *force
	}
}

type allPostsOptionImpl struct {
	offset  int
	max     int
	incl    []string
	start   int
	threads int
	force   bool
}

func (a *allPostsOptionImpl) Offset() int    { return a.offset }
func (a *allPostsOptionImpl) Max() int       { return a.max }
func (a *allPostsOptionImpl) Incl() []string { return a.incl }
func (a *allPostsOptionImpl) Start() int     { return a.start }
func (a *allPostsOptionImpl) Threads() int   { return a.threads }
func (a *allPostsOptionImpl) Force() bool    { return a.force }

func makeAllPostsOptionImpl(opts ...AllPostsOption) *allPostsOptionImpl {
	res := &allPostsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeAllPostsOptions(opts ...AllPostsOption) AllPostsOptions {
	return makeAllPostsOptionImpl(opts...)
}
