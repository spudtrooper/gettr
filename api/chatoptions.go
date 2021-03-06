// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Chat --outfile=chatoptions.go "debug"

type ChatOption func(*chatOptionImpl)

type ChatOptions interface {
	Debug() bool
}

func ChatDebug(debug bool) ChatOption {
	return func(opts *chatOptionImpl) {
		opts.debug = debug
	}
}
func ChatDebugFlag(debug *bool) ChatOption {
	return func(opts *chatOptionImpl) {
		opts.debug = *debug
	}
}

type chatOptionImpl struct {
	debug bool
}

func (c *chatOptionImpl) Debug() bool { return c.debug }

func makeChatOptionImpl(opts ...ChatOption) *chatOptionImpl {
	res := &chatOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeChatOptions(opts ...ChatOption) ChatOptions {
	return makeChatOptionImpl(opts...)
}
