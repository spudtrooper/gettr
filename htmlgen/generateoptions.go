package htmlgen

// genopts --opt_type=GeneratOption --prefix=Generate --outfile=htmlgen/generateoptions.go 'writeCSV' 'writeSimpleHTML' 'writeDescriptionsHTML' 'writeTwitterFollowersHTML' 'writeHTML' 'limit:int' 'all' 'threads:int'

type GeneratOption func(*generatOptionImpl)

type GeneratOptions interface {
	WriteCSV() bool
	WriteSimpleHTML() bool
	WriteDescriptionsHTML() bool
	WriteTwitterFollowersHTML() bool
	WriteHTML() bool
	Limit() int
	All() bool
	Threads() int
}

func GenerateWriteCSV(writeCSV bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.writeCSV = writeCSV
	}
}

func GenerateWriteSimpleHTML(writeSimpleHTML bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.writeSimpleHTML = writeSimpleHTML
	}
}

func GenerateWriteDescriptionsHTML(writeDescriptionsHTML bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.writeDescriptionsHTML = writeDescriptionsHTML
	}
}

func GenerateWriteTwitterFollowersHTML(writeTwitterFollowersHTML bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.writeTwitterFollowersHTML = writeTwitterFollowersHTML
	}
}

func GenerateWriteHTML(writeHTML bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.writeHTML = writeHTML
	}
}

func GenerateLimit(limit int) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.limit = limit
	}
}

func GenerateAll(all bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.all = all
	}
}

func GenerateThreads(threads int) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.threads = threads
	}
}

type generatOptionImpl struct {
	writeCSV                  bool
	writeSimpleHTML           bool
	writeDescriptionsHTML     bool
	writeTwitterFollowersHTML bool
	writeHTML                 bool
	limit                     int
	all                       bool
	threads                   int
}

func (g *generatOptionImpl) WriteCSV() bool                  { return g.writeCSV }
func (g *generatOptionImpl) WriteSimpleHTML() bool           { return g.writeSimpleHTML }
func (g *generatOptionImpl) WriteDescriptionsHTML() bool     { return g.writeDescriptionsHTML }
func (g *generatOptionImpl) WriteTwitterFollowersHTML() bool { return g.writeTwitterFollowersHTML }
func (g *generatOptionImpl) WriteHTML() bool                 { return g.writeHTML }
func (g *generatOptionImpl) Limit() int                      { return g.limit }
func (g *generatOptionImpl) All() bool                       { return g.all }
func (g *generatOptionImpl) Threads() int                    { return g.threads }

func makeGeneratOptionImpl(opts ...GeneratOption) *generatOptionImpl {
	res := &generatOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeGeneratOptions(opts ...GeneratOption) GeneratOptions {
	return makeGeneratOptionImpl(opts...)
}