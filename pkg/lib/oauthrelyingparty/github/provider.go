package github

import (
	"context"
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

const (
	githubAuthorizationURL string = "https://github.com/login/oauth/authorize"
	// nolint: gosec
	githubTokenURL    string = "https://github.com/login/oauth/access_token"
	githubUserInfoURL string = "https://api.github.com/user"
)

type Github struct{}

func (Github) GetJSONSchema() map[string]interface{} {
	builder := validation.SchemaBuilder{}
	builder.Type(validation.TypeObject)
	builder.Properties().
		Property("type", validation.SchemaBuilder{}.Type(validation.TypeString)).
		Property("client_id", validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1)).
		Property("claims", validation.SchemaBuilder{}.Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Properties().
			Property("email", validation.SchemaBuilder{}.Type(validation.TypeObject).
				AdditionalPropertiesFalse().Properties().
				Property("assume_verified", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
				Property("required", validation.SchemaBuilder{}.Type(validation.TypeBoolean)),
			),
		)
	builder.Required("type", "client_id")
	return builder
}

func (Github) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
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

func (p Github) GetAuthorizationURL(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
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

func (p Github) GetUserProfile(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	accessTokenResp, err := p.exchangeCode(ctx, deps, param)
	if err != nil {
		return
	}

	userProfile, err := p.fetchUserInfo(ctx, deps, accessTokenResp)
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
		err = errorutil.WithDetails(err, errorutil.Details{
			"ProviderType": apierrors.APIErrorDetail.Value(deps.ProviderConfig.Type()),
		})
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

func (Github) exchangeCode(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (accessTokenResp oauthrelyingpartyutil.AccessTokenResp, err error) {
	code, err := oauthrelyingpartyutil.GetCode(param.Query)
	if err != nil {
		return
	}

	q := make(url.Values)
	q.Set("client_id", deps.ProviderConfig.ClientID())
	q.Set("client_secret", deps.ClientSecret)
	q.Set("code", code)
	q.Set("redirect_uri", param.RedirectURI)

	body := strings.NewReader(q.Encode())
	req, _ := http.NewRequestWithContext(ctx, "POST", githubTokenURL, body)
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
		err = &errResp
		return
	}

	return
}

func (Github) fetchUserInfo(ctx context.Context, deps oauthrelyingparty.Dependencies, accessTokenResp oauthrelyingpartyutil.AccessTokenResp) (userProfile map[string]interface{}, err error) {
	tokenType := accessTokenResp.TokenType()
	accessTokenValue := accessTokenResp.AccessToken()
	authorizationHeader := fmt.Sprintf("%s %s", tokenType, accessTokenValue)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserInfoURL, nil)
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
