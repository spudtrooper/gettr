// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=SharePost --outfile=sharepostoptions.go "debug"

type SharePostOption func(*sharePostOptionImpl)

type SharePostOptions interface {
	Debug() bool
}

func SharePostDebug(debug bool) SharePostOption {
	return func(opts *sharePostOptionImpl) {
		opts.debug = debug
	}
}
func SharePostDebugFlag(debug *bool) SharePostOption {
	return func(opts *sharePostOptionImpl) {
		opts.debug = *debug
	}
}

type sharePostOptionImpl struct {
	debug bool
}

func (s *sharePostOptionImpl) Debug() bool { return s.debug }

func makeSharePostOptionImpl(opts ...SharePostOption) *sharePostOptionImpl {
	res := &sharePostOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeSharePostOptions(opts ...SharePostOption) SharePostOptions {
	return makeSharePostOptionImpl(opts...)
}
