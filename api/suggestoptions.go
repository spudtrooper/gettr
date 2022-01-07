package api

// genopts --opt_type=SuggestOption --prefix=Suggest --outfile=api/suggestoptions.go 'max:int'

type SuggestOption func(*suggestOptionImpl)

type SuggestOptions interface {
	Max() int
}

func SuggestMax(max int) SuggestOption {
	return func(opts *suggestOptionImpl) {
		opts.max = max
	}
}

type suggestOptionImpl struct {
	max int
}

func (s *suggestOptionImpl) Max() int { return s.max }

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
