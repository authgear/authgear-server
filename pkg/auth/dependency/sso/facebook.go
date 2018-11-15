package sso

type FacebookImpl struct {
	Config Config
}

func (f *FacebookImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	url = "facebook"
	return
}
