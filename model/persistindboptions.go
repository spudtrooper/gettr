package model

//go:generate genopts --opt_type=PersistInDBOption --prefix=PersistInDB --outfile=persistindboptions.go "threads:int"

type PersistInDBOption func(*persistInDBOptionImpl)

type PersistInDBOptions interface {
	Threads() int
}

func PersistInDBThreads(threads int) PersistInDBOption {
	return func(opts *persistInDBOptionImpl) {
		opts.threads = threads
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
