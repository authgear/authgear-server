package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type IdentificationData struct {
	TypedData
	Options []IdentificationOption `json:"options"`
}

func NewIdentificationData(d IdentificationData) IdentificationData {
	d.Type = DataTypeIdentificationData
	return d
}

var _ authflow.Data = IdentificationData{}

func (IdentificationData) Data() {}

type IdentificationOption struct {
	Identification model.AuthenticationFlowIdentification `json:"identification"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// ProviderType is specific to OAuth. Kept for backwards compatibility; use OAuthProviderType.
	ProviderType string `json:"provider_type,omitempty"`
	// OAuthProviderType is the canonical name for ProviderType.
	OAuthProviderType string `json:"oauth_provider_type,omitempty"`
	// Alias is specific to OAuth. Kept for backwards compatibility; use OAuthProviderAlias.
	Alias string `json:"alias,omitempty"`
	// OAuthProviderAlias is the canonical name for Alias.
	OAuthProviderAlias string `json:"oauth_provider_alias,omitempty"`
	// WechatAppType is specific to OAuth.
	WechatAppType wechat.AppType `json:"wechat_app_type,omitempty"`
	// ProviderStatus is specific to OAuth.
	ProviderStatus OAuthProviderStatus `json:"provider_status,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`

	// Server is specific to LDAP
	ServerName string `json:"server_name,omitempty"`

	// DisplayName is specific to SelectAccount. Unmasked: it identifies the
	// account already bound to the caller's own session cookie, not an
	// as-yet-unauthenticated identity, so there is nothing to mask.
	DisplayName string `json:"display_name,omitempty"`

	// UserID is specific to SelectAccount. Exposed so a UI (Custom or
	// built-in) can apply its own eligibility policy — e.g. checking this
	// option's user against login_hint/id_token_hint — using only this API's
	// data, since the flow itself deliberately does not implement that
	// filtering (see NewIdentificationOptionsSelectAccount).
	UserID string `json:"user_id,omitempty"`
}

func NewIdentificationOptionIDToken(flows authflow.Flows, i model.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
	}
}

func NewIdentificationOptionLoginID(flows authflow.Flows, i model.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
	}
}

func NewIdentificationOptionsOAuth(flows authflow.Flows, oauthConfig *config.OAuthSSOConfig, oauthFeatureConfig *config.OAuthSSOProvidersFeatureConfig, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig, demoCredentials *config.SSOOAuthDemoCredentials) []IdentificationOption {
	output := []IdentificationOption{}
	for _, p := range oauthConfig.Providers {
		if !identity.IsOAuthSSOProviderTypeDisabled(p.AsProviderConfig(), oauthFeatureConfig) {
			status := p.ComputeProviderStatus(demoCredentials)

			output = append(output, IdentificationOption{
				Identification:     model.AuthenticationFlowIdentificationOAuth,
				BotProtection:      GetBotProtectionData(flows, authflowCfg, appCfg),
				ProviderType:       p.AsProviderConfig().Type(),
				OAuthProviderType:  p.AsProviderConfig().Type(),
				Alias:              p.Alias(),
				OAuthProviderAlias: p.Alias(),
				WechatAppType:      wechat.ProviderConfig(p).AppType(),
				ProviderStatus:     status,
			})
		}
	}
	return output
}

func NewIdentificationOptionPasskey(flows authflow.Flows, requestOptions *model.WebAuthnRequestOptions, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: model.AuthenticationFlowIdentificationPasskey,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
		RequestOptions: requestOptions,
	}
}

func NewIdentificationOptionLDAP(ldapConfig *config.LDAPConfig, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) []IdentificationOption {
	output := []IdentificationOption{}
	for _, s := range ldapConfig.Servers {
		output = append(output, IdentificationOption{
			Identification: model.AuthenticationFlowIdentificationLDAP,
			ServerName:     s.Name,
			// TODO(DEV-1659): Support bot protection in LDAP
			// BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
		})
	}
	return output
}

// NewIdentificationOptionsSelectAccount returns the select_account options
// derived from the current IDP session cookie. Each option's UserID is both
// exposed to the API response (so a UI can apply its own policy, e.g.
// login_hint matching) and used server-side for the later session-freshness
// re-check (see resolveSelectAccountSession). It returns a slice rather than
// a single option so that supporting multiple concurrent accounts later only
// requires enumerating more than one entry here — every caller of this
// function stays unchanged.
//
// Today it omits the option (returning nil, nil) when:
//   - there is no IDP session,
//   - the session does not carry the pre-authenticated-url scope (every
//     ordinary identity-provider cookie session carries it implicitly, see
//     oauth.SessionScopes — this only actually excludes an offline-grant-
//     derived session that was not granted it),
//   - the session was established with "do not persist" semantics
//     (x_suppress_idp_session_cookie), or
//   - the resolved prompt contains "login" (including when an expired
//     max_age was folded into an implied prompt=login upstream).
//
// This scope check is the authoritative gate: it must live here, not only
// in the built-in Auth UI's webapp handler, because this function (and
// resolveSelectAccountSession's matching re-check) is reachable directly
// via the JSON authentication-flow API by a Custom UI, which never goes
// through the webapp handler at all.
//
// login_hint/id_token_hint eligibility filtering is intentionally not
// implemented yet — the option is offered whenever a usable session exists,
// regardless of any hint.
func NewIdentificationOptionsSelectAccount(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	authflowCfg *config.AuthenticationFlowBotProtection,
	appCfg *config.BotProtectionConfig,
) (options []InternalIdentificationOption, err error) {
	sess := session.GetSession(ctx)
	if sess == nil {
		return nil, nil
	}
	if !oauth.ContainsAllScopes(oauth.SessionScopes(sess), []string{oauth.PreAuthenticatedURLScope}) {
		return nil, nil
	}
	if authflow.GetSuppressIDPSessionCookie(ctx) {
		return nil, nil
	}
	if slice.ContainsString(authflow.GetSession(ctx).Prompt, "login") {
		return nil, nil
	}

	userID := sess.GetAuthenticationInfo().UserID

	identities, err := deps.Identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	displayName := selectAccountDisplayName(identities)

	return []InternalIdentificationOption{
		{
			Option: IdentificationOption{
				Identification: model.AuthenticationFlowIdentificationSelectAccount,
				BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
				DisplayName:    displayName,
				UserID:         userID,
			},
		},
	}, nil
}

// selectAccountDisplayNamePriorities indicates which identity type will be
// shown as the select_account option's display_name.
// This mirrors pkg/auth/handler/webapp.identitiesDisplayNamePriorities;
// duplicated rather than imported because this package cannot import
// pkg/auth/handler/webapp (that package already imports this one).
var selectAccountDisplayNamePriorities = map[model.IdentityType]int{
	model.IdentityTypeLoginID: 2,
	model.IdentityTypeOAuth:   1,
}

func selectAccountDisplayName(identities []*identity.Info) string {
	level := 0
	var i *identity.Info
	for _, perIdentity := range identities {
		l := selectAccountDisplayNamePriorities[perIdentity.Type]
		if l >= level {
			level = l
			i = perIdentity
		}
	}

	if i == nil {
		return ""
	}

	switch i.Type {
	case model.IdentityTypeLoginID:
		return i.DisplayID()
	case model.IdentityTypeOAuth:
		providerType := i.OAuth.ProviderID.Type
		displayID := i.DisplayID()
		if displayID != "" {
			return fmt.Sprintf("%s:%s", providerType, displayID)
		}
		return providerType
	case model.IdentityTypeAnonymous:
		return "anonymous"
	case model.IdentityTypeBiometric:
		return "biometric"
	case model.IdentityTypePasskey:
		return "passkey"
	case model.IdentityTypeSIWE:
		return "siwe"
	case model.IdentityTypeLDAP:
		return "ldap"
	default:
		return ""
	}
}

// InternalIdentificationOption wraps an IdentificationOption. An
// IntentLoginFlowStepIdentify/IntentSignupLoginFlowStepIdentify's Options
// slice is stored as []InternalIdentificationOption precisely so that a
// client-supplied "index" can look up the same option a client-facing
// response was built from with a single, direct slice index — see
// OutputData (which maps this down to []IdentificationOption via slice.Map)
// and each intent's select_account dispatch case, which reads
// Option.UserID back out by index to re-check against the current session
// on submission (see resolveSelectAccountSession).
type InternalIdentificationOption struct {
	Option IdentificationOption
}

func (i InternalIdentificationOption) ToIdentificationOption() IdentificationOption {
	return i.Option
}

func (i *IdentificationOption) isBotProtectionRequired() bool {
	if i.BotProtection == nil {
		return false
	}
	if i.BotProtection.Enabled != nil && *i.BotProtection.Enabled && i.BotProtection.Provider != nil && i.BotProtection.Provider.Type != "" {
		return true
	}

	return false
}
