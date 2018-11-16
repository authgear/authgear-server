package sso

type LinkedInImpl struct {
	Setting Setting
	Config  Config
}

func (f *LinkedInImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "linkedin"
	return
}
