package protocol

type EndSessionRequest map[string]string

func (r EndSessionRequest) IDTokenHint() string           { return r["id_token_hint"] }
func (r EndSessionRequest) PostLogoutRedirectURI() string { return r["post_logout_redirect_uri"] }
func (r EndSessionRequest) State() string                 { return r["state"] }
