package api

//go:generate genopts --prefix=Post --outfile=api/postoptions.go "incl:[]string" "max"

type PostOption func(*postOptionImpl)

type PostOptions interface {
	Incl() []string
	Max() bool
}

func PostIncl(incl []string) PostOption {
	return func(opts *postOptionImpl) {
		opts.incl = incl
	}
}
func PostInclFlag(incl *[]string) PostOption {
	return func(opts *postOptionImpl) {
		opts.incl = *incl
	}
}

func PostMax(max bool) PostOption {
	return func(opts *postOptionImpl) {
		opts.max = max
	}
}
func PostMaxFlag(max *bool) PostOption {
	return func(opts *postOptionImpl) {
		opts.max = *max
	}
}

type postOptionImpl struct {
	incl []string
	max  bool
}

func (p *postOptionImpl) Incl() []string { return p.incl }
func (p *postOptionImpl) Max() bool      { return p.max }

func makePostOptionImpl(opts ...PostOption) *postOptionImpl {
	res := &postOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakePostOptions(opts ...PostOption) PostOptions {
	return makePostOptionImpl(opts...)
}
