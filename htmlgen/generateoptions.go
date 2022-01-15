package htmlgen

// genopts --opt_type=GenerateOption --prefix=Generate --outfile=htmlgen/generateoptions.go 'writeCSV' 'writeSimpleHTML' 'writeDescriptionsHTML' 'writeTwitterFollowersHTML' 'writeHTML' 'limit:int' 'all' 'threads:int'

type GenerateOption func(*generateOptionImpl)

type GenerateOptions interface {
	WriteCSV() bool
	WriteSimpleHTML() bool
	WriteDescriptionsHTML() bool
	WriteTwitterFollowersHTML() bool
	WriteHTML() bool
	Limit() int
	All() bool
	Threads() int
}

func GenerateWriteCSV(writeCSV bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.writeCSV = writeCSV
	}
}

func GenerateWriteSimpleHTML(writeSimpleHTML bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.writeSimpleHTML = writeSimpleHTML
	}
}

func GenerateWriteDescriptionsHTML(writeDescriptionsHTML bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.writeDescriptionsHTML = writeDescriptionsHTML
	}
}

func GenerateWriteTwitterFollowersHTML(writeTwitterFollowersHTML bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.writeTwitterFollowersHTML = writeTwitterFollowersHTML
	}
}

func GenerateWriteHTML(writeHTML bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.writeHTML = writeHTML
	}
}

func GenerateLimit(limit int) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.limit = limit
	}
}

func GenerateAll(all bool) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.all = all
	}
}

func GenerateThreads(threads int) GenerateOption {
	return func(opts *generateOptionImpl) {
		opts.threads = threads
	}
}

type generateOptionImpl struct {
	writeCSV                  bool
	writeSimpleHTML           bool
	writeDescriptionsHTML     bool
	writeTwitterFollowersHTML bool
	writeHTML                 bool
	limit                     int
	all                       bool
	threads                   int
}

func (g *generateOptionImpl) WriteCSV() bool                  { return g.writeCSV }
func (g *generateOptionImpl) WriteSimpleHTML() bool           { return g.writeSimpleHTML }
func (g *generateOptionImpl) WriteDescriptionsHTML() bool     { return g.writeDescriptionsHTML }
func (g *generateOptionImpl) WriteTwitterFollowersHTML() bool { return g.writeTwitterFollowersHTML }
func (g *generateOptionImpl) WriteHTML() bool                 { return g.writeHTML }
func (g *generateOptionImpl) Limit() int                      { return g.limit }
func (g *generateOptionImpl) All() bool                       { return g.all }
func (g *generateOptionImpl) Threads() int                    { return g.threads }

func makeGenerateOptionImpl(opts ...GenerateOption) *generateOptionImpl {
	res := &generateOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeGenerateOptions(opts ...GenerateOption) GenerateOptions {
	return makeGenerateOptionImpl(opts...)
}
