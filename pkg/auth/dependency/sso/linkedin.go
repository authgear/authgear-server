package sso

type LinkedInImpl struct {
}

func (f LinkedInImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "linkedin"
	return
}
