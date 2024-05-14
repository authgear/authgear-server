package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Github{})
}

const Type = liboauthrelyingparty.TypeGithub

var _ oauthrelyingparty.Provider = Github{}
var _ liboauthrelyingparty.BuiltinProvider = Github{}

var Schema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" },
		"type": { "type": "string" },
		"modify_disabled": { "type": "boolean" },
		"client_id": { "type": "string", "minLength": 1 },
		"claims": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"email": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"assume_verified": { "type": "boolean" },
						"required": { "type": "boolean" }
					}
				}
			}
		}
	},
	"required": ["alias", "type", "client_id"]
}
`)

const (
	githubAuthorizationURL string = "https://github.com/login/oauth/authorize"
	// nolint: gosec
	githubTokenURL    string = "https://github.com/login/oauth/access_token"
	githubUserInfoURL string = "https://api.github.com/user"
)

type Github struct{}

func (Github) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Github) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (Github) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// Github does NOT support OIDC.
	// Github user ID is public, not scoped to anything.
	return oauthrelyingparty.NewProviderID(cfg.Type(), nil)
}

func (Github) scope() []string {
	// https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps
	return []string{"read:user", "user:email"}
}

func (p Github) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
	return oauthrelyingpartyutil.MakeAuthorizationURL(githubAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:    deps.ProviderConfig.ClientID(),
		RedirectURI: param.RedirectURI,
		Scope:       p.scope(),
		// ResponseType is unset.
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Nonce is unset.
	}.Query()), nil
}

func (p Github) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	accessTokenResp, err := p.exchangeCode(deps, param)
	if err != nil {
		return
	}

	userProfile, err := p.fetchUserInfo(deps, accessTokenResp)
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
	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
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
			"ProviderType": apierrors.APIErrorDetail.Value(deps.ProviderConfig.Type()),
		})
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

func (Github) exchangeCode(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (accessTokenResp oauthrelyingpartyutil.AccessTokenResp, err error) {
	q := make(url.Values)
	q.Set("client_id", deps.ProviderConfig.ClientID())
	q.Set("client_secret", deps.ClientSecret)
	q.Set("code", param.Code)
	q.Set("redirect_uri", param.RedirectURI)

	body := strings.NewReader(q.Encode())
	req, _ := http.NewRequest("POST", githubTokenURL, body)
	// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := deps.HTTPClient.Do(req)
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

func (Github) fetchUserInfo(deps oauthrelyingparty.Dependencies, accessTokenResp oauthrelyingpartyutil.AccessTokenResp) (userProfile map[string]interface{}, err error) {
	tokenType := accessTokenResp.TokenType()
	accessTokenValue := accessTokenResp.AccessToken()
	authorizationHeader := fmt.Sprintf("%s %s", tokenType, accessTokenValue)

	req, err := http.NewRequest(http.MethodGet, githubUserInfoURL, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", authorizationHeader)

	resp, err := deps.HTTPClient.Do(req)
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
