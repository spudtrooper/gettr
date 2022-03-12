package model

//go:generate genopts --prefix=UserInfo --outfile=model/userinfooptions.go "dontRetry"

type UserInfoOption func(*userInfoOptionImpl)

type UserInfoOptions interface {
	DontRetry() bool
}

func UserInfoDontRetry(dontRetry bool) UserInfoOption {
	return func(opts *userInfoOptionImpl) {
		opts.dontRetry = dontRetry
	}
}
func UserInfoDontRetryFlag(dontRetry *bool) UserInfoOption {
	return func(opts *userInfoOptionImpl) {
		opts.dontRetry = *dontRetry
	}
}

type userInfoOptionImpl struct {
	dontRetry bool
}

func (u *userInfoOptionImpl) DontRetry() bool { return u.dontRetry }

func makeUserInfoOptionImpl(opts ...UserInfoOption) *userInfoOptionImpl {
	res := &userInfoOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUserInfoOptions(opts ...UserInfoOption) UserInfoOptions {
	return makeUserInfoOptionImpl(opts...)
}
