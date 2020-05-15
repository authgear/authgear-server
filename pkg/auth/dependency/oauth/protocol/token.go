package protocol

import "strconv"

type TokenRequest map[string]string
type TokenResponse map[string]string

// OAuth 2.0

func (r TokenRequest) GrantType() string    { return r["grant_type"] }
func (r TokenRequest) Code() string         { return r["code"] }
func (r TokenRequest) RedirectURI() string  { return r["redirect_uri"] }
func (r TokenRequest) ClientID() string     { return r["client_id"] }
func (r TokenRequest) RefreshToken() string { return r["refresh_token"] }
func (r TokenRequest) JWT() string          { return r["jwt"] }

func (r TokenResponse) AccessToken(v string)  { r["access_token"] = v }
func (r TokenResponse) TokenType(v string)    { r["token_type"] = v }
func (r TokenResponse) ExpiresIn(v int)       { r["expires_in"] = strconv.Itoa(v) }
func (r TokenResponse) RefreshToken(v string) { r["refresh_token"] = v }
func (r TokenResponse) Scope(v string)        { r["scope"] = v }

func (r TokenResponse) GetAccessToken() string  { return r["access_token"] }
func (r TokenResponse) GetRefreshToken() string { return r["refresh_token"] }
func (r TokenResponse) GetExpiresIn() int {
	expiresIn, err := strconv.Atoi(r["expires_in"])
	if err != nil {
		panic(err)
	}
	return expiresIn
}

// OIDC extension

func (r TokenResponse) IDToken(v string) { r["id_token"] = v }

// PKCE extension

func (r TokenRequest) CodeVerifier() string { return r["code_verifier"] }
