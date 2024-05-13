package oauthrelyingpartyutil

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwsutil"
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

type jwtClock struct {
	Clock oauthrelyingparty.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type OIDCDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, InvalidConfiguration.New(fmt.Sprintf("failed to fetch OIDC discovery document with HTTP status code 404: %s", endpoint))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OIDC discovery document: unexpected status code: %d", resp.StatusCode)
	}

	var document OIDCDiscoveryDocument
	err = json.NewDecoder(resp.Body).Decode(&document)
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (d *OIDCDiscoveryDocument) MakeOAuthURL(params AuthorizationURLParams) string {
	return MakeAuthorizationURL(d.AuthorizationEndpoint, params.Query())
}

func (d *OIDCDiscoveryDocument) FetchJWKs(client *http.Client) (jwk.Set, error) {
	resp, err := client.Get(d.JWKSUri)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OIDC JWKs: unexpected status code: %d", resp.StatusCode)
	}
	return jwk.ParseReader(resp.Body)
}

func (d *OIDCDiscoveryDocument) ExchangeCode(
	client *http.Client,
	clock oauthrelyingparty.Clock,
	code string,
	jwks jwk.Set,
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		if err != nil {
			return nil, err
		}
	} else {
		var errorResp oauthrelyingparty.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, err
		}
		err = ErrorResponseAsError(errorResp)
		return nil, err
	}

	idToken := []byte(tokenResp.IDToken())

	_, payload, err := jwsutil.VerifyWithSet(jwks, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token signature: %w", err)
	}

	err = jwt.Validate(
		payload,
		jwt.WithClock(jwtClock{clock}),
		jwt.WithAudience(clientID),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to validate JWT claims: %w", err)
	}

	// Verify nonce only when it was specified in the authorization url.
	if nonce != "" {
		hashedNonceIface, ok := payload.Get("nonce")
		if !ok {
			return nil, OAuthProtocolError.New("nonce not found in ID token")
		}

		hashedNonce, ok := hashedNonceIface.(string)
		if !ok {
			return nil, OAuthProtocolError.New(fmt.Sprintf("nonce in ID token is of invalid type: %T", hashedNonceIface))
		}

		if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
			return nil, fmt.Errorf("invalid nonce")
		}
	}

	return payload, nil
}
