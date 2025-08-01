package viewmodels

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

type AuthflowViewModel struct {
	IdentificationOptions []declarative.IdentificationOption
	IdentityCandidates    []map[string]interface{}

	LoginIDDisabled        bool
	PhoneLoginIDEnabled    bool
	EmailLoginIDEnabled    bool
	UsernameLoginIDEnabled bool
	PasskeyEnabled         bool

	// NonPhoneLoginIDInputType is the "type" attribute for the non-phone <input>.
	// It is "email" or "text".
	NonPhoneLoginIDInputType string

	// NonPhoneLoginIDType is the type of non-phone login ID.
	// It is "email", "username" or "email_or_username".
	NonPhoneLoginIDType string

	// LoginIDKey is the key the end-user has chosen.
	// It is "email", "phone", or "username".
	LoginIDKey string

	// LoginIDInputType is the input the end-user has chosen.
	// It is "email", "phone" or "text".
	LoginIDInputType     string
	LoginIDDefaultValue  string
	LoginIDInputReadOnly bool
	AlternativesDisabled bool

	// LoginIDContextualType is the type the end-user thinks they should enter.
	// It depends on LoginIDInputType.
	// It is "email", "phone", "username", or "email_or_username".
	LoginIDContextualType string

	PasskeyRequestOptionsJSON string

	PhoneLoginIDBotProtectionRequired    bool
	EmailLoginIDBotProtectionRequired    bool
	UsernameLoginIDBotProtectionRequired bool
	PasskeyBotProtectionRequired         bool
	OAuthBotProtectionRequired           bool
}

type AuthflowViewModeler struct {
	Authentication          *config.AuthenticationConfig
	LoginID                 *config.LoginIDConfig
	Identity                *config.IdentityConfig
	SSOOAuthDemoCredentials *config.SSOOAuthDemoCredentials
}

// nolint: gocognit
func (m *AuthflowViewModeler) NewWithAuthflow(
	s *webapp.Session,
	f *authflow.FlowResponse,
	r *http.Request,
) AuthflowViewModel {
	options := webapp.GetIdentificationOptions(f)

	var firstLoginIDIdentification model.AuthenticationFlowIdentification
	var firstNonPhoneLoginIDIdentification model.AuthenticationFlowIdentification
	hasEmail := false
	hasUsername := false
	hasPhone := false
	passkeyEnabled := false
	passkeyRequestOptionsJSON := ""

	bpRequiredEmail := false
	bpRequiredPhone := false
	bpRequiredUsername := false
	bpRequiredPasskey := false
	bpRequiredOAuth := false

	for _, o := range options {
		switch o.Identification {
		case model.AuthenticationFlowIdentificationEmail:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationEmail
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = model.AuthenticationFlowIdentificationEmail
			}
			hasEmail = true
			if o.BotProtection.IsRequired() {
				bpRequiredEmail = true
			}
		case model.AuthenticationFlowIdentificationPhone:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationPhone
			}
			hasPhone = true
			if o.BotProtection.IsRequired() {
				bpRequiredPhone = true
			}
		case model.AuthenticationFlowIdentificationUsername:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationUsername
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = model.AuthenticationFlowIdentificationUsername
			}
			hasUsername = true
			if o.BotProtection.IsRequired() {
				bpRequiredUsername = true
			}
		case model.AuthenticationFlowIdentificationPasskey:
			passkeyEnabled = true
			bytes, err := json.Marshal(o.RequestOptions)
			if err != nil {
				panic(err)
			}
			passkeyRequestOptionsJSON = string(bytes)
			if o.BotProtection.IsRequired() {
				bpRequiredPasskey = true
			}
		case model.AuthenticationFlowIdentificationOAuth:
			if o.BotProtection.IsRequired() {
				bpRequiredOAuth = true
			}
		}
	}

	// Then we determine NonPhoneLoginIDInputType.
	nonPhoneLoginIDInputType := "text"
	if hasEmail && !hasUsername {
		nonPhoneLoginIDInputType = "email"
	}

	nonPhoneLoginIDType := "email"
	switch {
	case hasEmail && hasUsername:
		nonPhoneLoginIDType = "email_or_username"
	case hasUsername:
		nonPhoneLoginIDType = "username"
	}

	loginIDInputType := r.FormValue("q_login_id_input_type")
	switch {
	case loginIDInputType == "phone" && hasPhone:
		// valid
		break
	case loginIDInputType == nonPhoneLoginIDInputType:
		// valid
		break
	default:
		if firstLoginIDIdentification != "" {
			if firstLoginIDIdentification == model.AuthenticationFlowIdentificationPhone {
				loginIDInputType = "phone"
			} else {
				loginIDInputType = nonPhoneLoginIDInputType
			}
		} else {
			// Otherwise set a default value.
			loginIDInputType = "text"
		}
	}

	loginIDKey := r.FormValue("q_login_id_key")
	switch {
	case loginIDKey == "email" && hasEmail:
		// valid
		break
	case loginIDKey == "phone" && hasPhone:
		// valid
		break
	case loginIDKey == "username" && hasUsername:
		// valid
		break
	default:
		// Otherwise set q_login_id_key to match q_login_id_input_type.
		switch loginIDInputType {
		case "phone":
			loginIDKey = "phone"
		case "email":
			loginIDKey = "email"
		default:
			if firstNonPhoneLoginIDIdentification != "" {
				loginIDKey = string(firstNonPhoneLoginIDIdentification)
			} else {
				// Otherwise set a default value.
				loginIDKey = "email"
			}
		}
	}

	var loginIDDefaultValue string
	var loginIDInputReadOnly bool
	var alternativesDisabled bool

	var loginHint *oauth.LoginHint
	if s != nil && s.LoginHint != "" {
		loginHint, _ = oauth.ParseLoginHint(s.LoginHint)
	}
	if loginHint != nil && loginHint.Type == oauth.LoginHintTypeLoginID {
		switch {
		case loginHint.LoginIDEmail != "" && hasEmail:
			loginIDKey = "email"
			loginIDInputType = "email"
			loginIDDefaultValue = loginHint.LoginIDEmail
		case loginHint.LoginIDPhone != "" && hasPhone:
			loginIDKey = "phone"
			loginIDInputType = "phone"
			loginIDDefaultValue = loginHint.LoginIDPhone
		case loginHint.LoginIDUsername != "" && hasUsername:
			loginIDKey = "username"
			loginIDInputType = "text"
			loginIDDefaultValue = loginHint.LoginIDUsername
		}
		if loginHint.Enforce {
			loginIDInputReadOnly = true
			alternativesDisabled = true
		}
	}

	var loginIDContextualType string
	switch {
	case loginIDInputType == "phone":
		loginIDContextualType = "phone"
	default:
		loginIDContextualType = nonPhoneLoginIDType
	}

	loginIDDisabled := !hasEmail && !hasUsername && !hasPhone

	makeLoginIDCandidate := func(t model.LoginIDKeyType) map[string]interface{} {
		var inputType string
		switch t {
		case model.LoginIDKeyTypePhone:
			inputType = "phone"
		case model.LoginIDKeyTypeEmail:
			fallthrough
		case model.LoginIDKeyTypeUsername:
			inputType = nonPhoneLoginIDInputType
		default:
			panic(fmt.Errorf("unexpected login id key: %v", t))
		}
		candidate := map[string]interface{}{
			"type":                string(model.IdentityTypeLoginID),
			"login_id_type":       string(t),
			"login_id_key":        string(t),
			"login_id_input_type": inputType,
		}
		return candidate
	}

	var candidates []map[string]interface{}
	for _, o := range options {
		switch o.Identification {
		case model.AuthenticationFlowIdentificationEmail:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeEmail))
		case model.AuthenticationFlowIdentificationPhone:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypePhone))
		case model.AuthenticationFlowIdentificationUsername:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeUsername))
		case model.AuthenticationFlowIdentificationOAuth:
			candidate := map[string]interface{}{
				"type":              string(model.IdentityTypeOAuth),
				"provider_type":     string(o.ProviderType),
				"provider_alias":    o.Alias,
				"provider_app_type": string(o.WechatAppType),
				"provider_status":   string(o.ProviderStatus),
			}
			candidates = append(candidates, candidate)
		case model.AuthenticationFlowIdentificationLDAP:
			candidate := map[string]interface{}{
				"type":        string(model.IdentityTypeLDAP),
				"server_name": o.ServerName,
			}
			candidates = append(candidates, candidate)
		case model.AuthenticationFlowIdentificationPasskey:
			// Passkey was not handled by candidates.
			break
		}
	}

	return AuthflowViewModel{
		IdentificationOptions: options,
		IdentityCandidates:    candidates,

		LoginIDDisabled:           loginIDDisabled,
		PhoneLoginIDEnabled:       hasPhone,
		EmailLoginIDEnabled:       hasEmail,
		UsernameLoginIDEnabled:    hasUsername,
		PasskeyEnabled:            passkeyEnabled,
		PasskeyRequestOptionsJSON: passkeyRequestOptionsJSON,

		LoginIDKey:               loginIDKey,
		LoginIDInputType:         loginIDInputType,
		LoginIDDefaultValue:      loginIDDefaultValue,
		LoginIDInputReadOnly:     loginIDInputReadOnly,
		AlternativesDisabled:     alternativesDisabled,
		NonPhoneLoginIDInputType: nonPhoneLoginIDInputType,
		NonPhoneLoginIDType:      nonPhoneLoginIDType,
		LoginIDContextualType:    loginIDContextualType,

		PhoneLoginIDBotProtectionRequired:    bpRequiredPhone,
		EmailLoginIDBotProtectionRequired:    bpRequiredEmail,
		UsernameLoginIDBotProtectionRequired: bpRequiredUsername,
		PasskeyBotProtectionRequired:         bpRequiredPasskey,
		OAuthBotProtectionRequired:           bpRequiredOAuth,
	}
}

func (m *AuthflowViewModeler) NewWithAccountRecoveryAuthflow(f *authflow.FlowResponse, r *http.Request) AuthflowViewModel {
	options := webapp.GetAccountRecoveryIdentificationOptions(f)
	bpRequiredEmail := false
	bpRequiredPhone := false

	for _, opt := range options {
		switch opt.Identification {
		case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
			bpRequiredEmail = opt.BotProtection.IsRequired()
		case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
			bpRequiredPhone = opt.BotProtection.IsRequired()
		}
	}

	return AuthflowViewModel{
		EmailLoginIDBotProtectionRequired: bpRequiredEmail,
		PhoneLoginIDBotProtectionRequired: bpRequiredPhone,
	}
}

// nolint: gocognit
func (m *AuthflowViewModeler) NewWithConfig() AuthflowViewModel {
	var firstLoginIDIdentification model.AuthenticationFlowIdentification
	var firstNonPhoneLoginIDIdentification model.AuthenticationFlowIdentification
	hasEmail := false
	hasUsername := false
	hasPhone := false
	passkeyEnabled := false
	passkeyRequestOptionsJSON := ""

	for _, loginIDKey := range m.Identity.LoginID.Keys {
		switch loginIDKey.Type {
		case model.LoginIDKeyTypeEmail:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationEmail
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = model.AuthenticationFlowIdentificationEmail
			}
			hasEmail = true
		case model.LoginIDKeyTypePhone:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationPhone
			}
			hasPhone = true
		case model.LoginIDKeyTypeUsername:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = model.AuthenticationFlowIdentificationUsername
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = model.AuthenticationFlowIdentificationUsername
			}
			hasUsername = true
		}
	}
	for _, identityType := range m.Authentication.Identities {
		if identityType == model.IdentityTypePasskey {
			passkeyEnabled = true
			passkeyRequestOptionsJSON = "{}"
			break
		}
	}

	// Then we determine NonPhoneLoginIDInputType.
	nonPhoneLoginIDInputType := "text"
	if hasEmail && !hasUsername {
		nonPhoneLoginIDInputType = "email"
	}

	nonPhoneLoginIDType := "email"
	switch {
	case hasEmail && hasUsername:
		nonPhoneLoginIDType = "email_or_username"
	case hasUsername:
		nonPhoneLoginIDType = "username"
	}

	loginIDInputType := ""
	if firstLoginIDIdentification != "" {
		if firstLoginIDIdentification == model.AuthenticationFlowIdentificationPhone {
			loginIDInputType = "phone"
		} else {
			loginIDInputType = nonPhoneLoginIDInputType
		}
	} else {
		// Otherwise set a default value.
		loginIDInputType = "text"
	}

	loginIDKey := ""
	switch loginIDInputType {
	case "phone":
		loginIDKey = "phone"
	case "email":
		loginIDKey = "email"
	default:
		if firstNonPhoneLoginIDIdentification != "" {
			loginIDKey = string(firstNonPhoneLoginIDIdentification)
		} else {
			// Otherwise set a default value.
			loginIDKey = "email"
		}
	}

	var loginIDContextualType string
	switch {
	case loginIDInputType == "phone":
		loginIDContextualType = "phone"
	default:
		loginIDContextualType = nonPhoneLoginIDType
	}

	loginIDDisabled := !hasEmail && !hasUsername && !hasPhone

	makeLoginIDCandidate := func(t model.LoginIDKeyType) map[string]interface{} {
		var inputType string
		switch t {
		case model.LoginIDKeyTypePhone:
			inputType = "phone"
		case model.LoginIDKeyTypeEmail:
			fallthrough
		case model.LoginIDKeyTypeUsername:
			inputType = nonPhoneLoginIDInputType
		default:
			panic(fmt.Errorf("unexpected login id key: %v", t))
		}
		candidate := map[string]interface{}{
			"type":                string(model.IdentityTypeLoginID),
			"login_id_type":       string(t),
			"login_id_key":        string(t),
			"login_id_input_type": inputType,
		}
		return candidate
	}

	var candidates []map[string]interface{}
	for _, loginIDKey := range m.Identity.LoginID.Keys {
		switch loginIDKey.Type {
		case model.LoginIDKeyTypeEmail:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeEmail))
		case model.LoginIDKeyTypePhone:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypePhone))
		case model.LoginIDKeyTypeUsername:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeUsername))
		}
	}

	for _, oauthProvider := range m.Identity.OAuth.Providers {
		candidate := map[string]interface{}{
			"type":              string(model.IdentityTypeOAuth),
			"provider_type":     oauthProvider.AsProviderConfig().Type(),
			"provider_alias":    oauthProvider.Alias(),
			"provider_app_type": wechat.ProviderConfig(oauthProvider).AppType(),
			"provider_status":   oauthProvider.ComputeProviderStatus(m.SSOOAuthDemoCredentials),
		}
		candidates = append(candidates, candidate)
	}

	for _, ldapServer := range m.Identity.LDAP.Servers {
		candidate := map[string]interface{}{
			"type":        string(model.IdentityTypeLDAP),
			"server_name": ldapServer.Name,
		}
		candidates = append(candidates, candidate)
	}

	return AuthflowViewModel{
		IdentityCandidates: candidates,

		LoginIDDisabled:           loginIDDisabled,
		PhoneLoginIDEnabled:       hasPhone,
		EmailLoginIDEnabled:       hasEmail,
		UsernameLoginIDEnabled:    hasUsername,
		PasskeyEnabled:            passkeyEnabled,
		PasskeyRequestOptionsJSON: passkeyRequestOptionsJSON,

		LoginIDKey:               loginIDKey,
		LoginIDInputType:         loginIDInputType,
		NonPhoneLoginIDInputType: nonPhoneLoginIDInputType,
		NonPhoneLoginIDType:      nonPhoneLoginIDType,
		LoginIDContextualType:    loginIDContextualType,
	}
}
