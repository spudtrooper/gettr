// DO NOT EDIT MANUALLY: Generated from https://github.com/spudtrooper/genopts
package api

//go:generate genopts --prefix=Request --outfile=requestoptions.go "extraHeaders:map[string]string" "host:string" "customPayload:interface{}" "noRedirect"

type RequestOption func(*requestOptionImpl)

type RequestOptions interface {
	ExtraHeaders() map[string]string
	Host() string
	CustomPayload() interface{}
	NoRedirect() bool
}

func RequestExtraHeaders(extraHeaders map[string]string) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.extraHeaders = extraHeaders
	}
}
func RequestExtraHeadersFlag(extraHeaders *map[string]string) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.extraHeaders = *extraHeaders
	}
}

func RequestHost(host string) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.host = host
	}
}
func RequestHostFlag(host *string) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.host = *host
	}
}

func RequestCustomPayload(customPayload interface{}) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.customPayload = customPayload
	}
}
func RequestCustomPayloadFlag(customPayload *interface{}) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.customPayload = *customPayload
	}
}

func RequestNoRedirect(noRedirect bool) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.noRedirect = noRedirect
	}
}
func RequestNoRedirectFlag(noRedirect *bool) RequestOption {
	return func(opts *requestOptionImpl) {
		opts.noRedirect = *noRedirect
	}
}

type requestOptionImpl struct {
	extraHeaders  map[string]string
	host          string
	customPayload interface{}
	noRedirect    bool
}

func (r *requestOptionImpl) ExtraHeaders() map[string]string { return r.extraHeaders }
func (r *requestOptionImpl) Host() string                    { return r.host }
func (r *requestOptionImpl) CustomPayload() interface{}      { return r.customPayload }
func (r *requestOptionImpl) NoRedirect() bool                { return r.noRedirect }

func makeRequestOptionImpl(opts ...RequestOption) *requestOptionImpl {
	res := &requestOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeRequestOptions(opts ...RequestOption) RequestOptions {
	return makeRequestOptionImpl(opts...)
}
