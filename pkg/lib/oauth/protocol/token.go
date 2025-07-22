package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

type TokenRequest url.Values
type TokenResponse map[string]interface{}

// OAuth 2.0

func (r TokenRequest) GrantType() string    { return url.Values(r).Get("grant_type") }
func (r TokenRequest) Code() string         { return url.Values(r).Get("code") }
func (r TokenRequest) RedirectURI() string  { return url.Values(r).Get("redirect_uri") }
func (r TokenRequest) ClientID() string     { return url.Values(r).Get("client_id") }
func (r TokenRequest) RefreshToken() string { return url.Values(r).Get("refresh_token") }
func (r TokenRequest) JWT() string          { return url.Values(r).Get("jwt") }
func (r TokenRequest) App2AppDeviceKeyJWT() string {
	return url.Values(r).Get("x_app2app_device_key_jwt")
}
func (r TokenRequest) ClientSecret() string       { return url.Values(r).Get("client_secret") }
func (r TokenRequest) Scope() []string            { return parseSpaceDelimitedString(url.Values(r).Get("scope")) }
func (r TokenRequest) RequestedTokenType() string { return url.Values(r).Get("requested_token_type") }
func (r TokenRequest) Audience() string           { return url.Values(r).Get("audience") }
func (r TokenRequest) SubjectTokenType() string   { return url.Values(r).Get("subject_token_type") }
func (r TokenRequest) SubjectToken() string       { return url.Values(r).Get("subject_token") }
func (r TokenRequest) ActorTokenType() string     { return url.Values(r).Get("actor_token_type") }
func (r TokenRequest) ActorToken() string         { return url.Values(r).Get("actor_token") }
func (r TokenRequest) DeviceSecret() string       { return url.Values(r).Get("device_secret") }
func (r TokenRequest) Resource() string           { return url.Values(r).Get("resource") }

func (r TokenResponse) AccessToken(v string)     { r["access_token"] = v }
func (r TokenResponse) TokenType(v string)       { r["token_type"] = v }
func (r TokenResponse) IssuedTokenType(v string) { r["issued_token_type"] = v }
func (r TokenResponse) ExpiresIn(v int)          { r["expires_in"] = v }
func (r TokenResponse) RefreshToken(v string)    { r["refresh_token"] = v }
func (r TokenResponse) DeviceSecret(v string)    { r["device_secret"] = v }
func (r TokenResponse) Scope(v string)           { r["scope"] = v }

// OIDC extension

func (r TokenResponse) IDToken(v string) { r["id_token"] = v }

// PKCE extension

func (r TokenRequest) CodeVerifier() string        { return url.Values(r).Get("code_verifier") }
func (r TokenRequest) CodeChallenge() string       { return url.Values(r).Get("code_challenge") }
func (r TokenRequest) CodeChallengeMethod() string { return url.Values(r).Get("code_challenge_method") }

// Proprietary
func (r TokenRequest) DeviceInfo() (map[string]interface{}, error) {
	encoded := url.Values(r).Get("x_device_info")
	if encoded == "" {
		return nil, nil
	}

	bytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var deviceInfo map[string]interface{}
	err = json.Unmarshal(bytes, &deviceInfo)
	if err != nil {
		return nil, err
	}

	formatted := deviceinfo.DeviceModel(deviceInfo)
	if formatted == "" {
		return nil, fmt.Errorf("invalid device info: %s", string(bytes))
	}

	return deviceInfo, nil
}

// App2App

func (r TokenResponse) Code(v string) { r["code"] = v }
