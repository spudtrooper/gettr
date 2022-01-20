package model

import "go.mongodb.org/mongo-driver/bson"

// genopts --opt_type=CountPostsOption --prefix=CountPosts --outfile=model/countpostsoptions.go 'filter:bson.D'

type CountPostsOption func(*countPostsOptionImpl)

type CountPostsOptions interface {
	Filter() bson.D
}

func CountPostsFilter(filter bson.D) CountPostsOption {
	return func(opts *countPostsOptionImpl) {
		opts.filter = filter
	}
}

type countPostsOptionImpl struct {
	filter bson.D
}

func (c *countPostsOptionImpl) Filter() bson.D { return c.filter }

func makeCountPostsOptionImpl(opts ...CountPostsOption) *countPostsOptionImpl {
	res := &countPostsOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeCountPostsOptions(opts ...CountPostsOption) CountPostsOptions {
	return makeCountPostsOptionImpl(opts...)
}
