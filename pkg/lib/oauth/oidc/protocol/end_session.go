package protocol

type EndSessionRequest map[string]string

func (r EndSessionRequest) IDTokenHint() string           { return r["id_token_hint"] }
func (r EndSessionRequest) PostLogoutRedirectURI() string { return r["post_logout_redirect_uri"] }
func (r EndSessionRequest) State() string                 { return r["state"] }

// WithoutIDTokenHint returns a copy of r with id_token_hint removed. Use this
// before forwarding r anywhere id_token_hint should not be re-exposed (e.g.
// into a new redirect URL), once id_token_hint has already served its purpose.
func (r EndSessionRequest) WithoutIDTokenHint() EndSessionRequest {
	out := EndSessionRequest{}
	for k, v := range r {
		if k == "id_token_hint" {
			continue
		}
		out[k] = v
	}
	return out
}
