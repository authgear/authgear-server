package sso

type LinkedInImpl struct {
	Config Config
}

func (f *LinkedInImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "linkedin"
	return
}
