package api

//go:generate genopts --prefix=UpdateProfile --outfile=api/updateprofileoptions.go "description:string" "backgroundImage:string" "icon:string" "location:string" "website:string"

type UpdateProfileOption func(*updateProfileOptionImpl)

type UpdateProfileOptions interface {
	Description() string
	BackgroundImage() string
	Icon() string
	Location() string
	Website() string
}

func UpdateProfileDescription(description string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.description = description
	}
}
func UpdateProfileDescriptionFlag(description *string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.description = *description
	}
}

func UpdateProfileBackgroundImage(backgroundImage string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.backgroundImage = backgroundImage
	}
}
func UpdateProfileBackgroundImageFlag(backgroundImage *string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.backgroundImage = *backgroundImage
	}
}

func UpdateProfileIcon(icon string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.icon = icon
	}
}
func UpdateProfileIconFlag(icon *string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.icon = *icon
	}
}

func UpdateProfileLocation(location string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.location = location
	}
}
func UpdateProfileLocationFlag(location *string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.location = *location
	}
}

func UpdateProfileWebsite(website string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.website = website
	}
}
func UpdateProfileWebsiteFlag(website *string) UpdateProfileOption {
	return func(opts *updateProfileOptionImpl) {
		opts.website = *website
	}
}

type updateProfileOptionImpl struct {
	description     string
	backgroundImage string
	icon            string
	location        string
	website         string
}

func (u *updateProfileOptionImpl) Description() string     { return u.description }
func (u *updateProfileOptionImpl) BackgroundImage() string { return u.backgroundImage }
func (u *updateProfileOptionImpl) Icon() string            { return u.icon }
func (u *updateProfileOptionImpl) Location() string        { return u.location }
func (u *updateProfileOptionImpl) Website() string         { return u.website }

func makeUpdateProfileOptionImpl(opts ...UpdateProfileOption) *updateProfileOptionImpl {
	res := &updateProfileOptionImpl{}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func MakeUpdateProfileOptions(opts ...UpdateProfileOption) UpdateProfileOptions {
	return makeUpdateProfileOptionImpl(opts...)
}
