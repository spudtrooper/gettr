// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=FactoryClient --outfile=factoryclientoptions.go "debug"

type FactoryClientOption func(*factoryClientOptionImpl)

type FactoryClientOptions interface {
	Debug() bool
}

func FactoryClientDebug(debug bool) FactoryClientOption {
	return func(opts *factoryClientOptionImpl) {
		opts.debug = debug
	}
}
func FactoryClientDebugFlag(debug *bool) FactoryClientOption {
	return func(opts *factoryClientOptionImpl) {
		opts.debug = *debug
	}
}

type factoryClientOptionImpl struct {
	debug bool
}

func (f *factoryClientOptionImpl) Debug() bool { return f.debug }

func makeFactoryClientOptionImpl(opts ...FactoryClientOption) *factoryClientOptionImpl {
	res := &factoryClientOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeFactoryClientOptions(opts ...FactoryClientOption) FactoryClientOptions {
	return makeFactoryClientOptionImpl(opts...)
}
