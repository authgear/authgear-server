package sso

type InstagramImpl struct {
}

func (f InstagramImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "instagram"
	return
}
