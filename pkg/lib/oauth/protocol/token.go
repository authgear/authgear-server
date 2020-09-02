package protocol

type TokenRequest map[string]string
type TokenResponse map[string]interface{}

// OAuth 2.0

func (r TokenRequest) GrantType() string    { return r["grant_type"] }
func (r TokenRequest) Code() string         { return r["code"] }
func (r TokenRequest) RedirectURI() string  { return r["redirect_uri"] }
func (r TokenRequest) ClientID() string     { return r["client_id"] }
func (r TokenRequest) RefreshToken() string { return r["refresh_token"] }
func (r TokenRequest) JWT() string          { return r["jwt"] }

func (r TokenResponse) AccessToken(v string)  { r["access_token"] = v }
func (r TokenResponse) TokenType(v string)    { r["token_type"] = v }
func (r TokenResponse) ExpiresIn(v int)       { r["expires_in"] = v }
func (r TokenResponse) RefreshToken(v string) { r["refresh_token"] = v }
func (r TokenResponse) Scope(v string)        { r["scope"] = v }

// OIDC extension

func (r TokenResponse) IDToken(v string) { r["id_token"] = v }

// PKCE extension

func (r TokenRequest) CodeVerifier() string { return r["code_verifier"] }
