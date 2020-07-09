package sso

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/jwtutil"
)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type OIDCAuthParams struct {
	ProviderConfig config.OAuthSSOProviderConfig
	RedirectURI    string
	Nonce          string
	State          string
	ExtraParams    map[string]string
}

type OIDCDiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSUri               string `json:"jwks_uri"`
}

func FetchOIDCDiscoveryDocument(client *http.Client, endpoint string) (*OIDCDiscoveryDocument, error) {
	resp, err := client.Get(endpoint)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("unexpected status code: %d", resp.StatusCode)
	}
	var document OIDCDiscoveryDocument
	err = json.NewDecoder(resp.Body).Decode(&document)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (d *OIDCDiscoveryDocument) MakeOAuthURL(params OIDCAuthParams) string {
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", params.ProviderConfig.ClientID)
	v.Add("redirect_uri", params.RedirectURI)
	v.Add("scope", params.ProviderConfig.Type.Scope())
	v.Add("nonce", params.Nonce)
	v.Add("response_mode", "form_post")
	for key, value := range params.ExtraParams {
		v.Add(key, value)
	}
	v.Add("state", params.State)

	return d.AuthorizationEndpoint + "?" + v.Encode()
}

func (d *OIDCDiscoveryDocument) FetchJWKs(client *http.Client) (*jwk.Set, error) {
	resp, err := client.Get(d.JWKSUri)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("unexpected status code: %d", resp.StatusCode)
	}
	return jwk.Parse(resp.Body)
}

func (d *OIDCDiscoveryDocument) ExchangeCode(
	client *http.Client,
	clock clock.Clock,
	code string,
	jwks *jwk.Set,
	clientID string,
	clientSecret string,
	redirectURI string,
	nonce string,
	tokenResp *AccessTokenResp,
) (jwt.Token, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", clientID)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	body.Set("client_secret", clientSecret)

	resp, err := client.PostForm(d.TokenEndpoint, body)
	if err != nil {
		return nil, NewSSOFailed(NetworkFailed, "failed to connect authorization server")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		if err != nil {
			return nil, err
		}
	} else {
		var errorResp oauthErrorResp
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, err
		}
		err = errorResp.AsError()
		return nil, err
	}

	idToken := []byte(tokenResp.IDToken())
	hdr, payload, err := jwtutil.SplitWithoutVerify(idToken)
	if err != nil {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid ID token")
	}

	keyID := hdr.KeyID()
	if keyID == "" {
		return nil, NewSSOFailed(SSOUnauthorized, "no kid")
	}

	keys := jwks.LookupKeyID(keyID)
	if len(keys) != 1 {
		return nil, NewSSOFailed(SSOUnauthorized, "failed to find signing key")
	}
	key := keys[0]

	_, err = jws.VerifyWithJWK(idToken, key)
	if err != nil {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid JWT signature")
	}

	err = jwt.Verify(
		payload,
		jwt.WithClock(jwtClock{clock}),
		jwt.WithAudience(clientID),
	)
	if err != nil {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid aud")
	}

	hashedNonceIface, ok := payload.Get("nonce")
	if !ok {
		return nil, NewSSOFailed(InvalidParams, "no nonce")
	}

	hashedNonce, ok := hashedNonceIface.(string)
	if !ok {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid nonce")
	}

	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid nonce")
	}

	return payload, nil
}
