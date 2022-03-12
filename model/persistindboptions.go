package model

//go:generate genopts --prefix=PersistInDB --outfile=model/persistindboptions.go "threads:int"

type PersistInDBOption func(*persistInDBOptionImpl)

type PersistInDBOptions interface {
	Threads() int
}

func PersistInDBThreads(threads int) PersistInDBOption {
	return func(opts *persistInDBOptionImpl) {
		opts.threads = threads
	}
}
func PersistInDBThreadsFlag(threads *int) PersistInDBOption {
	return func(opts *persistInDBOptionImpl) {
		opts.threads = *threads
	}
}

type persistInDBOptionImpl struct {
	threads int
}

func (p *persistInDBOptionImpl) Threads() int { return p.threads }

func makePersistInDBOptionImpl(opts ...PersistInDBOption) *persistInDBOptionImpl {
	res := &persistInDBOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakePersistInDBOptions(opts ...PersistInDBOption) PersistInDBOptions {
	return makePersistInDBOptionImpl(opts...)
}
