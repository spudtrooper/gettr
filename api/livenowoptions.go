package api

import "time"

//go:generate genopts --prefix=LiveNow --outfile=api/livenowoptions.go "offset:int" "max:int" "dir:string" "incl:[]string" "merge:string" "start:time.Time" "lang:string"

type LiveNowOption func(*liveNowOptionImpl)

type LiveNowOptions interface {
	Offset() int
	Max() int
	Dir() string
	Incl() []string
	Merge() string
	Start() time.Time
	Lang() string
}

func LiveNowOffset(offset int) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.offset = offset
	}
}
func LiveNowOffsetFlag(offset *int) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.offset = *offset
	}
}

func LiveNowMax(max int) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.max = max
	}
}
func LiveNowMaxFlag(max *int) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.max = *max
	}
}

func LiveNowDir(dir string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.dir = dir
	}
}
func LiveNowDirFlag(dir *string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.dir = *dir
	}
}

func LiveNowIncl(incl []string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.incl = incl
	}
}
func LiveNowInclFlag(incl *[]string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.incl = *incl
	}
}

func LiveNowMerge(merge string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.merge = merge
	}
}
func LiveNowMergeFlag(merge *string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.merge = *merge
	}
}

func LiveNowStart(start time.Time) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.start = start
	}
}
func LiveNowStartFlag(start *time.Time) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.start = *start
	}
}

func LiveNowLang(lang string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.lang = lang
	}
}
func LiveNowLangFlag(lang *string) LiveNowOption {
	return func(opts *liveNowOptionImpl) {
		opts.lang = *lang
	}
}

type liveNowOptionImpl struct {
	offset int
	max    int
	dir    string
	incl   []string
	merge  string
	start  time.Time
	lang   string
}

func (l *liveNowOptionImpl) Offset() int      { return l.offset }
func (l *liveNowOptionImpl) Max() int         { return l.max }
func (l *liveNowOptionImpl) Dir() string      { return l.dir }
func (l *liveNowOptionImpl) Incl() []string   { return l.incl }
func (l *liveNowOptionImpl) Merge() string    { return l.merge }
func (l *liveNowOptionImpl) Start() time.Time { return l.start }
func (l *liveNowOptionImpl) Lang() string     { return l.lang }

func makeLiveNowOptionImpl(opts ...LiveNowOption) *liveNowOptionImpl {
	res := &liveNowOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeLiveNowOptions(opts ...LiveNowOption) LiveNowOptions {
	return makeLiveNowOptionImpl(opts...)
}
