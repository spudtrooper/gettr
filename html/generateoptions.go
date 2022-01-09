package html

// genopts --opt_type=GeneratOption --prefix=Generate --outfile=html/generateoptions.go 'readCache' 'writeCSV' 'writeSimpleHTML' 'writeDescriptionsHTML' 'writeTwitterFollowersHTML' 'writeHTML' 'limit:int'

type GeneratOption func(*generatOptionImpl)

type GeneratOptions interface {
	ReadCache() bool
	WriteCSV() bool
	WriteSimpleHTML() bool
	WriteDescriptionsHTML() bool
	WriteTwitterFollowersHTML() bool
	WriteHTML() bool
	Limit() int
}

func GenerateReadCache(readCache bool) GeneratOption {
	return func(opts *generatOptionImpl) {
		opts.readCache = readCache
	}
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

type generatOptionImpl struct {
	readCache                 bool
	writeCSV                  bool
	writeSimpleHTML           bool
	writeDescriptionsHTML     bool
	writeTwitterFollowersHTML bool
	writeHTML                 bool
	limit                     int
}

func (g *generatOptionImpl) ReadCache() bool                 { return g.readCache }
func (g *generatOptionImpl) WriteCSV() bool                  { return g.writeCSV }
func (g *generatOptionImpl) WriteSimpleHTML() bool           { return g.writeSimpleHTML }
func (g *generatOptionImpl) WriteDescriptionsHTML() bool     { return g.writeDescriptionsHTML }
func (g *generatOptionImpl) WriteTwitterFollowersHTML() bool { return g.writeTwitterFollowersHTML }
func (g *generatOptionImpl) WriteHTML() bool                 { return g.writeHTML }
func (g *generatOptionImpl) Limit() int                      { return g.limit }

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
