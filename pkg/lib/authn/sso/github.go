package sso

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

const (
	githubAuthorizationURL string = "https://github.com/login/oauth/authorize"
	// nolint: gosec
	githubTokenURL    string = "https://github.com/login/oauth/access_token"
	githubUserInfoURL string = "https://api.github.com/user"
)

type GithubImpl struct {
	ProviderConfig               oauthrelyingparty.ProviderConfig
	ClientSecret                 string
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (g *GithubImpl) Config() oauthrelyingparty.ProviderConfig {
	return g.ProviderConfig
}

func (g *GithubImpl) GetAuthorizationURL(param GetAuthorizationURLOptions) (string, error) {
	// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
	return oauthrelyingpartyutil.MakeAuthorizationURL(githubAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:    g.ProviderConfig.ClientID(),
		RedirectURI: param.RedirectURI,
		Scope:       g.ProviderConfig.Scope(),
		// ResponseType is unset.
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Nonce is unset.
	}.Query()), nil
}

func (g *GithubImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	accessTokenResp, err := g.exchangeCode(r, param)
	if err != nil {
		return
	}

	userProfile, err := g.fetchUserInfo(accessTokenResp)
	if err != nil {
		return
	}
	authInfo.ProviderRawProfile = userProfile

	idJSONNumber, _ := userProfile["id"].(json.Number)
	email, _ := userProfile["email"].(string)
	login, _ := userProfile["login"].(string)
	picture, _ := userProfile["avatar_url"].(string)
	profile, _ := userProfile["html_url"].(string)

	id := string(idJSONNumber)

	authInfo.ProviderUserID = id
	emailRequired := g.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(map[string]interface{}{
		stdattrs.Email:     email,
		stdattrs.Name:      login,
		stdattrs.GivenName: login,
		stdattrs.Picture:   picture,
		stdattrs.Profile:   profile,
	}, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		err = apierrors.AddDetails(err, errorutil.Details{
			"ProviderType": apierrors.APIErrorDetail.Value(g.ProviderConfig.Type),
		})
		return
	}
	authInfo.StandardAttributes = stdAttrs

	err = g.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (g *GithubImpl) exchangeCode(r OAuthAuthorizationResponse, param GetAuthInfoParam) (accessTokenResp oauthrelyingpartyutil.AccessTokenResp, err error) {
	q := make(url.Values)
	q.Set("client_id", g.ProviderConfig.ClientID())
	q.Set("client_secret", g.ClientSecret)
	q.Set("code", r.Code)
	q.Set("redirect_uri", param.RedirectURI)

	body := strings.NewReader(q.Encode())
	req, _ := http.NewRequest("POST", githubTokenURL, body)
	// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&accessTokenResp)
		if err != nil {
			return
		}
	} else {
		var errResp oauthrelyingparty.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return
		}
		err = oauthrelyingpartyutil.ErrorResponseAsError(errResp)
	}

	return
}

func (g *GithubImpl) fetchUserInfo(accessTokenResp oauthrelyingpartyutil.AccessTokenResp) (userProfile map[string]interface{}, err error) {
	tokenType := accessTokenResp.TokenType()
	accessTokenValue := accessTokenResp.AccessToken()
	authorizationHeader := fmt.Sprintf("%s %s", tokenType, accessTokenValue)

	req, err := http.NewRequest(http.MethodGet, githubUserInfoURL, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", authorizationHeader)

	resp, err := g.HTTPClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to fetch user profile: unexpected status code: %d", resp.StatusCode)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	// Deserialize "id" as json.Number.
	decoder.UseNumber()
	err = decoder.Decode(&userProfile)
	if err != nil {
		return
	}

	return
}

var (
	_ OAuthProvider = &GithubImpl{}
)
