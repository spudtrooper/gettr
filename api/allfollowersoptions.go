package api

// genopts --opt_type=AllFollowersOption --prefix=AllFollowers --outfile=api/allfollowersoptions.go 'offset:int' 'max:int' 'incl:[]string' 'start:int'

type AllFollowersOption func(*allFollowersOptionImpl)

type AllFollowersOptions interface {
	Offset() int
	Max() int
	Incl() []string
	Start() int
}

func AllFollowersOffset(offset int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.offset = offset
	}
}

func AllFollowersMax(max int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.max = max
	}
}

func AllFollowersIncl(incl []string) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.incl = incl
	}
}

func AllFollowersStart(start int) AllFollowersOption {
	return func(opts *allFollowersOptionImpl) {
		opts.start = start
	}
}

type allFollowersOptionImpl struct {
	offset int
	max    int
	incl   []string
	start  int
}

func (a *allFollowersOptionImpl) Offset() int    { return a.offset }
func (a *allFollowersOptionImpl) Max() int       { return a.max }
func (a *allFollowersOptionImpl) Incl() []string { return a.incl }
func (a *allFollowersOptionImpl) Start() int     { return a.start }

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