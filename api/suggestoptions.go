// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Suggest --outfile=suggestoptions.go "max:int" "incl:[]string" "offset:int"

type SuggestOption func(*suggestOptionImpl)

type SuggestOptions interface {
	Max() int
	Incl() []string
	Offset() int
}

func SuggestMax(max int) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.max = max
	}
}
func SuggestMaxFlag(max *int) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.max = *max
	}
}

func SuggestIncl(incl []string) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.incl = incl
	}
}
func SuggestInclFlag(incl *[]string) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.incl = *incl
	}
}

func SuggestOffset(offset int) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.offset = offset
	}
}
func SuggestOffsetFlag(offset *int) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.offset = *offset
	}
}

type suggestOptionImpl struct {
	max    int
	incl   []string
	offset int
}

func (s *suggestOptionImpl) Max() int       { return s.max }
func (s *suggestOptionImpl) Incl() []string { return s.incl }
func (s *suggestOptionImpl) Offset() int    { return s.offset }

func makeSuggestOptionImpl(opts ...SuggestOption) *suggestOptionImpl {
	res := &suggestOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeSuggestOptions(opts ...SuggestOption) SuggestOptions {
	return makeSuggestOptionImpl(opts...)
}
