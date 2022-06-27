// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Reply --outfile=replyoptions.go "images:[]string" "debug:bool" "previewImage:string" "description:string" "title:string" "previewSource:string"

type ReplyOption func(*replyOptionImpl)

type ReplyOptions interface {
	Images() []string
	Debug() bool
	PreviewImage() string
	Description() string
	Title() string
	PreviewSource() string
}

func ReplyImages(images []string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.images = images
	}
}
func ReplyImagesFlag(images *[]string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.images = *images
	}
}

func ReplyDebug(debug bool) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.debug = debug
	}
}
func ReplyDebugFlag(debug *bool) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.debug = *debug
	}
}

func ReplyPreviewImage(previewImage string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.previewImage = previewImage
	}
}
func ReplyPreviewImageFlag(previewImage *string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.previewImage = *previewImage
	}
}

func ReplyDescription(description string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.description = description
	}
}
func ReplyDescriptionFlag(description *string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.description = *description
	}
}

func ReplyTitle(title string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.title = title
	}
}
func ReplyTitleFlag(title *string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.title = *title
	}
}

func ReplyPreviewSource(previewSource string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.previewSource = previewSource
	}
}
func ReplyPreviewSourceFlag(previewSource *string) ReplyOption {
	return func(opts *replyOptionImpl) {
		opts.previewSource = *previewSource
	}
}

type replyOptionImpl struct {
	images        []string
	debug         bool
	previewImage  string
	description   string
	title         string
	previewSource string
}

func (r *replyOptionImpl) Images() []string      { return r.images }
func (r *replyOptionImpl) Debug() bool           { return r.debug }
func (r *replyOptionImpl) PreviewImage() string  { return r.previewImage }
func (r *replyOptionImpl) Description() string   { return r.description }
func (r *replyOptionImpl) Title() string         { return r.title }
func (r *replyOptionImpl) PreviewSource() string { return r.previewSource }

func makeReplyOptionImpl(opts ...ReplyOption) *replyOptionImpl {
	res := &replyOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeReplyOptions(opts ...ReplyOption) ReplyOptions {
	return makeReplyOptionImpl(opts...)
}
