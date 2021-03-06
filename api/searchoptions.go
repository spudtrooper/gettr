// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Search --outfile=searchoptions.go "max:int" "incl:[]string" "offset:int" "debug"

type SearchOption func(*searchOptionImpl)

type SearchOptions interface {
	Max() int
	Incl() []string
	Offset() int
	Debug() bool
}

func SearchMax(max int) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.max = max
	}
}
func SearchMaxFlag(max *int) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.max = *max
	}
}

func SearchIncl(incl []string) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.incl = incl
	}
}
func SearchInclFlag(incl *[]string) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.incl = *incl
	}
}

func SearchOffset(offset int) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.offset = offset
	}
}
func SearchOffsetFlag(offset *int) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.offset = *offset
	}
}

func SearchDebug(debug bool) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.debug = debug
	}
}
func SearchDebugFlag(debug *bool) SearchOption {
	return func(opts *searchOptionImpl) {
		opts.debug = *debug
	}
}

type searchOptionImpl struct {
	max    int
	incl   []string
	offset int
	debug  bool
}

func (s *searchOptionImpl) Max() int       { return s.max }
func (s *searchOptionImpl) Incl() []string { return s.incl }
func (s *searchOptionImpl) Offset() int    { return s.offset }
func (s *searchOptionImpl) Debug() bool    { return s.debug }

func makeSearchOptionImpl(opts ...SearchOption) *searchOptionImpl {
	res := &searchOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeSearchOptions(opts ...SearchOption) SearchOptions {
	return makeSearchOptionImpl(opts...)
}
