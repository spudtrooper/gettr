package api

//go:generate genopts --prefix=Muted --outfile=api/mutedoptions.go "offset:int" "max:int" "incl:[]string"

type MutedOption func(*mutedOptionImpl)

type MutedOptions interface {
	Offset() int
	Max() int
	Incl() []string
}

func MutedOffset(offset int) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.offset = offset
	}
}
func MutedOffsetFlag(offset *int) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.offset = *offset
	}
}

func MutedMax(max int) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.max = max
	}
}
func MutedMaxFlag(max *int) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.max = *max
	}
}

func MutedIncl(incl []string) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.incl = incl
	}
}
func MutedInclFlag(incl *[]string) MutedOption {
	return func(opts *mutedOptionImpl) {
		opts.incl = *incl
	}
}

type mutedOptionImpl struct {
	offset int
	max    int
	incl   []string
}

func (m *mutedOptionImpl) Offset() int    { return m.offset }
func (m *mutedOptionImpl) Max() int       { return m.max }
func (m *mutedOptionImpl) Incl() []string { return m.incl }

func makeMutedOptionImpl(opts ...MutedOption) *mutedOptionImpl {
	res := &mutedOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeMutedOptions(opts ...MutedOption) MutedOptions {
	return makeMutedOptionImpl(opts...)
}
