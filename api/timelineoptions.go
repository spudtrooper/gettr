package api

import "time"

//go:generate genopts --prefix=Timeline --outfile=api/timelineoptions.go "offset:int" "max:int" "dir:string" "incl:[]string" "merge:string" "start:time.Time"

type TimelineOption func(*timelineOptionImpl)

type TimelineOptions interface {
	Offset() int
	Max() int
	Dir() string
	Incl() []string
	Merge() string
	Start() time.Time
}

func TimelineOffset(offset int) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.offset = offset
	}
}
func TimelineOffsetFlag(offset *int) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.offset = *offset
	}
}

func TimelineMax(max int) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.max = max
	}
}
func TimelineMaxFlag(max *int) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.max = *max
	}
}

func TimelineDir(dir string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.dir = dir
	}
}
func TimelineDirFlag(dir *string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.dir = *dir
	}
}

func TimelineIncl(incl []string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.incl = incl
	}
}
func TimelineInclFlag(incl *[]string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.incl = *incl
	}
}

func TimelineMerge(merge string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.merge = merge
	}
}
func TimelineMergeFlag(merge *string) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.merge = *merge
	}
}

func TimelineStart(start time.Time) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.start = start
	}
}
func TimelineStartFlag(start *time.Time) TimelineOption {
	return func(opts *timelineOptionImpl) {
		opts.start = *start
	}
}

type timelineOptionImpl struct {
	offset int
	max    int
	dir    string
	incl   []string
	merge  string
	start  time.Time
}

func (t *timelineOptionImpl) Offset() int      { return t.offset }
func (t *timelineOptionImpl) Max() int         { return t.max }
func (t *timelineOptionImpl) Dir() string      { return t.dir }
func (t *timelineOptionImpl) Incl() []string   { return t.incl }
func (t *timelineOptionImpl) Merge() string    { return t.merge }
func (t *timelineOptionImpl) Start() time.Time { return t.start }

func makeTimelineOptionImpl(opts ...TimelineOption) *timelineOptionImpl {
	res := &timelineOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeTimelineOptions(opts ...TimelineOption) TimelineOptions {
	return makeTimelineOptionImpl(opts...)
}
