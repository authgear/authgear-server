package protocol

import "strings"

type AuthorizationRequest map[string]string
type AuthorizationResponse map[string]string

// OAuth 2.0

func (r AuthorizationRequest) ResponseType() string { return r["response_type"] }
func (r AuthorizationRequest) ClientID() string     { return r["client_id"] }
func (r AuthorizationRequest) RedirectURI() string  { return r["redirect_uri"] }
func (r AuthorizationRequest) Scope() []string      { return parseSpaceDelimitedString(r["scope"]) }
func (r AuthorizationRequest) Prompt() []string     { return parseSpaceDelimitedString(r["prompt"]) }
func (r AuthorizationRequest) State() string        { return r["state"] }
func (r AuthorizationRequest) LoginHint() string    { return r["login_hint"] }

func (r AuthorizationRequest) SetPrompt(prompt []string)     { r["prompt"] = strings.Join(prompt, " ") }
func (r AuthorizationRequest) SetLoginHint(loginHint string) { r["login_hint"] = loginHint }

func (r AuthorizationResponse) Code(v string)  { r["code"] = v }
func (r AuthorizationResponse) State(v string) { r["state"] = v }

// OIDC extension

func (r AuthorizationRequest) Nonce() string       { return r["nonce"] }
func (r AuthorizationRequest) UILocales() []string { return parseSpaceDelimitedString(r["ui_locales"]) }

// PKCE extension

func (r AuthorizationRequest) CodeChallenge() string       { return r["code_challenge"] }
func (r AuthorizationRequest) CodeChallengeMethod() string { return r["code_challenge_method"] }
