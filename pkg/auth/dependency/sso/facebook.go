package sso

type FacebookImpl struct {
}

func (f FacebookImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "facebook"
	return
}
