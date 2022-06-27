// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Followers --outfile=followersoptions.go "offset:int" "max:int" "incl:[]string"

type FollowersOption func(*followersOptionImpl)

type FollowersOptions interface {
	Offset() int
	Max() int
	Incl() []string
}

func FollowersOffset(offset int) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.offset = offset
	}
}
func FollowersOffsetFlag(offset *int) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.offset = *offset
	}
}

func FollowersMax(max int) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.max = max
	}
}
func FollowersMaxFlag(max *int) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.max = *max
	}
}

func FollowersIncl(incl []string) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.incl = incl
	}
}
func FollowersInclFlag(incl *[]string) FollowersOption {
	return func(opts *followersOptionImpl) {
		opts.incl = *incl
	}
}

type followersOptionImpl struct {
	offset int
	max    int
	incl   []string
}

func (f *followersOptionImpl) Offset() int    { return f.offset }
func (f *followersOptionImpl) Max() int       { return f.max }
func (f *followersOptionImpl) Incl() []string { return f.incl }

func makeFollowersOptionImpl(opts ...FollowersOption) *followersOptionImpl {
	res := &followersOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeFollowersOptions(opts ...FollowersOption) FollowersOptions {
	return makeFollowersOptionImpl(opts...)
}
