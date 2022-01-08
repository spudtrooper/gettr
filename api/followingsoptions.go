package api

// genopts --opt_type=FollowingsOption --prefix=Followings --outfile=api/followingsoptions.go 'offset:int' 'max:int' 'incl:[]string'

type FollowingsOption func(*followingsOptionImpl)

type FollowingsOptions interface {
	Offset() int
	Max() int
	Incl() []string
}

func FollowingsOffset(offset int) FollowingsOption {
	return func(opts *followingsOptionImpl) {
		opts.offset = offset
	}
}

func FollowingsMax(max int) FollowingsOption {
	return func(opts *followingsOptionImpl) {
		opts.max = max
	}
}

func FollowingsIncl(incl []string) FollowingsOption {
	return func(opts *followingsOptionImpl) {
		opts.incl = incl
	}
}

type followingsOptionImpl struct {
	offset int
	max    int
	incl   []string
}

func (f *followingsOptionImpl) Offset() int    { return f.offset }
func (f *followingsOptionImpl) Max() int       { return f.max }
func (f *followingsOptionImpl) Incl() []string { return f.incl }

func makeFollowingsOptionImpl(opts ...FollowingsOption) *followingsOptionImpl {
	res := &followingsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeFollowingsOptions(opts ...FollowingsOption) FollowingsOptions {
	return makeFollowingsOptionImpl(opts...)
}
