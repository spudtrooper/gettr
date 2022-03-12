package api

//go:generate genopts --prefix=MakeClient --outfile=api/makeclientoptions.go "debug:bool"

type MakeClientOption func(*makeClientOptionImpl)

type MakeClientOptions interface {
	Debug() bool
}

func MakeClientDebug(debug bool) MakeClientOption {
	return func(opts *makeClientOptionImpl) {
		opts.debug = debug
	}
}
func MakeClientDebugFlag(debug *bool) MakeClientOption {
	return func(opts *makeClientOptionImpl) {
		opts.debug = *debug
	}
}

type makeClientOptionImpl struct {
	debug bool
}

func (m *makeClientOptionImpl) Debug() bool { return m.debug }

func makeMakeClientOptionImpl(opts ...MakeClientOption) *makeClientOptionImpl {
	res := &makeClientOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeMakeClientOptions(opts ...MakeClientOption) MakeClientOptions {
	return makeMakeClientOptionImpl(opts...)
}
