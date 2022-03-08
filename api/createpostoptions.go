package api

//go:generate genopts --opt_type=CreatePostOption --prefix=CreatePost --outfile=api/createpostoptions.go "images:[]string" "debug:bool" "previewImage:string" "description:string" "title:string" "previewSource:string"

type CreatePostOption func(*createPostOptionImpl)

type CreatePostOptions interface {
	Images() []string
	Debug() bool
	PreviewImage() string
	Description() string
	Title() string
	PreviewSource() string
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

func CreatePostTitle(title string) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.title = title
	}
}

func CreatePostPreviewSource(previewSource string) CreatePostOption {
	return func(opts *createPostOptionImpl) {
		opts.previewSource = previewSource
	}
}

type createPostOptionImpl struct {
	images        []string
	debug         bool
	previewImage  string
	description   string
	title         string
	previewSource string
}

func (c *createPostOptionImpl) Images() []string      { return c.images }
func (c *createPostOptionImpl) Debug() bool           { return c.debug }
func (c *createPostOptionImpl) PreviewImage() string  { return c.previewImage }
func (c *createPostOptionImpl) Description() string   { return c.description }
func (c *createPostOptionImpl) Title() string         { return c.title }
func (c *createPostOptionImpl) PreviewSource() string { return c.previewSource }

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
