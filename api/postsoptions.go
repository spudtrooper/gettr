package api

//go:generate genopts --prefix=Posts --outfile=api/postsoptions.go "offset:int" "max:int" "dir:string" "incl:[]string" "fp:string"

type PostsOption func(*postsOptionImpl)

type PostsOptions interface {
	Offset() int
	Max() int
	Dir() string
	Incl() []string
	Fp() string
}

func PostsOffset(offset int) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.offset = offset
	}
}
func PostsOffsetFlag(offset *int) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.offset = *offset
	}
}

func PostsMax(max int) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.max = max
	}
}
func PostsMaxFlag(max *int) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.max = *max
	}
}

func PostsDir(dir string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.dir = dir
	}
}
func PostsDirFlag(dir *string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.dir = *dir
	}
}

func PostsIncl(incl []string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.incl = incl
	}
}
func PostsInclFlag(incl *[]string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.incl = *incl
	}
}

func PostsFp(fp string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.fp = fp
	}
}
func PostsFpFlag(fp *string) PostsOption {
	return func(opts *postsOptionImpl) {
		opts.fp = *fp
	}
}

type postsOptionImpl struct {
	offset int
	max    int
	dir    string
	incl   []string
	fp     string
}

func (p *postsOptionImpl) Offset() int    { return p.offset }
func (p *postsOptionImpl) Max() int       { return p.max }
func (p *postsOptionImpl) Dir() string    { return p.dir }
func (p *postsOptionImpl) Incl() []string { return p.incl }
func (p *postsOptionImpl) Fp() string     { return p.fp }

func makePostsOptionImpl(opts ...PostsOption) *postsOptionImpl {
	res := &postsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakePostsOptions(opts ...PostsOption) PostsOptions {
	return makePostsOptionImpl(opts...)
}
