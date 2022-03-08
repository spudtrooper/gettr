package api

//go:generate genopts --opt_type=CommentsOption --prefix=Comments --outfile=commentsoptions.go "offset:int" "max:int" "dir:string" "incl:[]string"

type CommentsOption func(*commentsOptionImpl)

type CommentsOptions interface {
	Offset() int
	Max() int
	Dir() string
	Incl() []string
}

func CommentsOffset(offset int) CommentsOption {
	return func(opts *commentsOptionImpl) {
		opts.offset = offset
	}
}

func CommentsMax(max int) CommentsOption {
	return func(opts *commentsOptionImpl) {
		opts.max = max
	}
}

func CommentsDir(dir string) CommentsOption {
	return func(opts *commentsOptionImpl) {
		opts.dir = dir
	}
}

func CommentsIncl(incl []string) CommentsOption {
	return func(opts *commentsOptionImpl) {
		opts.incl = incl
	}
}

type commentsOptionImpl struct {
	offset int
	max    int
	dir    string
	incl   []string
}

func (c *commentsOptionImpl) Offset() int    { return c.offset }
func (c *commentsOptionImpl) Max() int       { return c.max }
func (c *commentsOptionImpl) Dir() string    { return c.dir }
func (c *commentsOptionImpl) Incl() []string { return c.incl }

func makeCommentsOptionImpl(opts ...CommentsOption) *commentsOptionImpl {
	res := &commentsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeCommentsOptions(opts ...CommentsOption) CommentsOptions {
	return makeCommentsOptionImpl(opts...)
}
