package protocol

type RevokeRequest map[string]string

func (r RevokeRequest) Token() string         { return r["token"] }
func (r RevokeRequest) TokenTypeHint() string { return r["token_type_hint"] }
