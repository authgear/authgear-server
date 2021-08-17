package protocol

import (
	"strconv"
	"time"
)

type AuthorizationRequest map[string]string

// OAuth 2.0
func (r AuthorizationRequest) ClientID() string     { return r["client_id"] }
func (r AuthorizationRequest) RedirectURI() string  { return r["redirect_uri"] }
func (r AuthorizationRequest) ResponseType() string { return r["response_type"] }
func (r AuthorizationRequest) ResponseMode() string { return r["response_mode"] }
func (r AuthorizationRequest) Scope() []string      { return parseSpaceDelimitedString(r["scope"]) }
func (r AuthorizationRequest) State() string        { return r["state"] }

// OIDC extension
func (r AuthorizationRequest) Prompt() []string    { return parseSpaceDelimitedString(r["prompt"]) }
func (r AuthorizationRequest) Nonce() string       { return r["nonce"] }
func (r AuthorizationRequest) UILocales() []string { return parseSpaceDelimitedString(r["ui_locales"]) }
func (r AuthorizationRequest) LoginHint() (string, bool) {
	loginHint, ok := r["login_hint"]
	return loginHint, ok
}
func (r AuthorizationRequest) IDTokenHint() (string, bool) {
	idTokenHint, ok := r["id_token_hint"]
	return idTokenHint, ok
}
func (r AuthorizationRequest) HasMaxAge() bool {
	_, ok := r["max_age"]
	return ok
}
func (r AuthorizationRequest) MaxAge() (duration time.Duration, ok bool) {
	numSecondsStr, ok := r["max_age"]
	if !ok {
		return
	}

	numSeconds, err := strconv.ParseInt(numSecondsStr, 10, 64)
	if err != nil {
		ok = false
		return
	}

	// Duration cannot be negative.
	if numSeconds < 0 {
		ok = false
		return
	}

	duration = time.Duration(numSeconds) * time.Second
	return
}

// PKCE extension
func (r AuthorizationRequest) CodeChallenge() string       { return r["code_challenge"] }
func (r AuthorizationRequest) CodeChallengeMethod() string { return r["code_challenge_method"] }

// Proprietary
func (r AuthorizationRequest) Platform() string          { return r["x_platform"] }
func (r AuthorizationRequest) WeChatRedirectURI() string { return r["x_wechat_redirect_uri"] }
func (r AuthorizationRequest) Page() string              { return r["x_page"] }

type AuthorizationResponse map[string]string

func (r AuthorizationResponse) Code(v string)  { r["code"] = v }
func (r AuthorizationResponse) State(v string) { r["state"] = v }
