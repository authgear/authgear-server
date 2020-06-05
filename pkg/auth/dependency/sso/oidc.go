package sso

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corejwt "github.com/skygeario/skygear-server/pkg/core/jwt"
)

type OIDCAuthParams struct {
	ProviderConfig config.OAuthProviderConfiguration
	RedirectURI    string
	Nonce          string
	EncodedState   string
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
	v.Add("scope", params.ProviderConfig.Scope)
	v.Add("nonce", params.Nonce)
	v.Add("response_mode", "form_post")
	for key, value := range params.ExtraParams {
		v.Add(key, value)
	}
	v.Add("state", params.EncodedState)

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
	code string,
	jwks *jwk.Set,
	urlPrefix *url.URL,
	clientID string,
	clientSecret string,
	redirectURI string,
	nonce string,
	nowUTC func() time.Time,
	tokenResp *AccessTokenResp,
) (corejwt.MapClaims, error) {
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

	idToken := tokenResp.IDToken()
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, NewSSOFailed(SSOUnauthorized, "no kid")
		}
		if key := jwks.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}
		return nil, NewSSOFailed(SSOUnauthorized, "failed to find signing key")
	}

	mapClaims := corejwt.MapClaims{}
	_, err = jwt.ParseWithClaims(idToken, &mapClaims, keyFunc)
	if err != nil {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid JWT signature")
	}

	if !mapClaims.VerifyAudience(clientID, true) {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid aud")
	}

	hashedNonce, ok := mapClaims["nonce"].(string)
	if !ok {
		return nil, NewSSOFailed(InvalidParams, "no nonce")
	}
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, NewSSOFailed(SSOUnauthorized, "invalid nonce")
	}

	return mapClaims, nil
}
