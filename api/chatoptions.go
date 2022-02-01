package api

// genopts --opt_type=ChatOption --prefix=Chat --outfile=api/chatoptions.go 'debug'

type ChatOption func(*chatOptionImpl)

type ChatOptions interface {
	Debug() bool
}

func ChatDebug(debug bool) ChatOption {
	return func(opts *chatOptionImpl) {
		opts.debug = debug
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
