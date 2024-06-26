package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

type TokenRequest map[string]string
type TokenResponse map[string]interface{}

// OAuth 2.0

func (r TokenRequest) GrantType() string           { return r["grant_type"] }
func (r TokenRequest) Code() string                { return r["code"] }
func (r TokenRequest) RedirectURI() string         { return r["redirect_uri"] }
func (r TokenRequest) ClientID() string            { return r["client_id"] }
func (r TokenRequest) RefreshToken() string        { return r["refresh_token"] }
func (r TokenRequest) JWT() string                 { return r["jwt"] }
func (r TokenRequest) App2AppDeviceKeyJWT() string { return r["x_app2app_device_key_jwt"] }
func (r TokenRequest) ClientSecret() string        { return r["client_secret"] }
func (r TokenRequest) Scope() []string             { return parseSpaceDelimitedString(r["scope"]) }

func (r TokenResponse) AccessToken(v string)  { r["access_token"] = v }
func (r TokenResponse) TokenType(v string)    { r["token_type"] = v }
func (r TokenResponse) ExpiresIn(v int)       { r["expires_in"] = v }
func (r TokenResponse) RefreshToken(v string) { r["refresh_token"] = v }
func (r TokenResponse) DeviceSecret(v string) { r["device_secret"] = v }
func (r TokenResponse) Scope(v string)        { r["scope"] = v }

// OIDC extension

func (r TokenResponse) IDToken(v string) { r["id_token"] = v }

// PKCE extension

func (r TokenRequest) CodeVerifier() string        { return r["code_verifier"] }
func (r TokenRequest) CodeChallenge() string       { return r["code_challenge"] }
func (r TokenRequest) CodeChallengeMethod() string { return r["code_challenge_method"] }

// Proprietary
func (r TokenRequest) DeviceInfo() (map[string]interface{}, error) {
	encoded := r["x_device_info"]
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
