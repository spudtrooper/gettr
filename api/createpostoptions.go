package api

// genopts --opt_type=CreatePostOption --prefix=CreatePost --outfile=api/createpostoptions.go 'images:[]string' 'debug:bool'

type CreatePostOption func(*createPostOptionImpl)

type CreatePostOptions interface {
	Images() []string
	Debug() bool
}

func CreatePostImages(images []string) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.images = images
	}
}

func CreatePostDebug(debug bool) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.debug = debug
	}
}

type createPostOptionImpl struct {
	images []string
	debug  bool
}

func (c *createPostOptionImpl) Images() []string { return c.images }
func (c *createPostOptionImpl) Debug() bool      { return c.debug }

func makeCreatePostOptionImpl(opts ...CreatePostOption) *createPostOptionImpl {
	res := &createPostOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeCreatePostOptions(opts ...CreatePostOption) CreatePostOptions {
	return makeCreatePostOptionImpl(opts...)
}
