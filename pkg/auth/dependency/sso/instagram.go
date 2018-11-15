package sso

type InstagramImpl struct {
	Config Config
}

func (f *InstagramImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "instagram"
	return
}
