package sso

type GoogleImpl struct {
}

func (f GoogleImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "google"
	return
}
