package api

// genopts --opt_type=CreatePostOption --prefix=CreatePost --outfile=api/createpostoptions.go 'images:[]string' 'debug:bool' 'previewImage:string' 'description:string'

type CreatePostOption func(*createPostOptionImpl)

type CreatePostOptions interface {
	Images() []string
	Debug() bool
	PreviewImage() string
	Description() string
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

func CreatePostPreviewImage(previewImage string) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.previewImage = previewImage
	}
}

func CreatePostDescription(description string) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.description = description
	}
}

type createPostOptionImpl struct {
	images       []string
	debug        bool
	previewImage string
	description  string
}

func (c *createPostOptionImpl) Images() []string     { return c.images }
func (c *createPostOptionImpl) Debug() bool          { return c.debug }
func (c *createPostOptionImpl) PreviewImage() string { return c.previewImage }
func (c *createPostOptionImpl) Description() string  { return c.description }

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
