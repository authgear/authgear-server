package viewmodels

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

type AuthflowViewModel struct {
	IdentityCandidates []map[string]interface{}

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
	LoginIDInputType string

	// LoginIDContextualType is the type the end-user thinks they should enter.
	// It depends on LoginIDInputType.
	// It is "email", "phone", "username", or "email_or_username".
	LoginIDContextualType string

	PasskeyRequestOptionsJSON string
}

type AuthflowViewModeler struct {
	Authentication *config.AuthenticationConfig
	LoginID        *config.LoginIDConfig
	Identity       *config.IdentityConfig
}

// nolint: gocognit
func (m *AuthflowViewModeler) NewWithAuthflow(f *authflow.FlowResponse, r *http.Request) AuthflowViewModel {
	options := webapp.GetIdentificationOptions(f)

	var firstLoginIDIdentification config.AuthenticationFlowIdentification
	var firstNonPhoneLoginIDIdentification config.AuthenticationFlowIdentification
	hasEmail := false
	hasUsername := false
	hasPhone := false
	passkeyEnabled := false
	passkeyRequestOptionsJSON := ""

	for _, o := range options {
		switch o.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationEmail
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = config.AuthenticationFlowIdentificationEmail
			}
			hasEmail = true
		case config.AuthenticationFlowIdentificationPhone:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationPhone
			}
			hasPhone = true
		case config.AuthenticationFlowIdentificationUsername:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationUsername
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = config.AuthenticationFlowIdentificationUsername
			}
			hasUsername = true
		case config.AuthenticationFlowIdentificationPasskey:
			passkeyEnabled = true
			bytes, err := json.Marshal(o.RequestOptions)
			if err != nil {
				panic(err)
			}
			passkeyRequestOptionsJSON = string(bytes)
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
			if firstLoginIDIdentification == config.AuthenticationFlowIdentificationPhone {
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
		case config.AuthenticationFlowIdentificationEmail:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeEmail))
		case config.AuthenticationFlowIdentificationPhone:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypePhone))
		case config.AuthenticationFlowIdentificationUsername:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeUsername))
		case config.AuthenticationFlowIdentificationOAuth:
			candidate := map[string]interface{}{
				"type":              string(model.IdentityTypeOAuth),
				"provider_type":     string(o.ProviderType),
				"provider_alias":    o.Alias,
				"provider_app_type": string(o.WechatAppType),
			}
			candidates = append(candidates, candidate)
		case config.AuthenticationFlowIdentificationPasskey:
			// Passkey was not handled by candidates.
			break
		}
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

func (m *AuthflowViewModeler) NewWithConfig() AuthflowViewModel {
	var firstLoginIDIdentification config.AuthenticationFlowIdentification
	var firstNonPhoneLoginIDIdentification config.AuthenticationFlowIdentification
	hasEmail := false
	hasUsername := false
	hasPhone := false
	passkeyEnabled := false
	passkeyRequestOptionsJSON := ""

	for _, loginIDKey := range m.Identity.LoginID.Keys {
		switch loginIDKey.Type {
		case model.LoginIDKeyTypeEmail:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationEmail
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = config.AuthenticationFlowIdentificationEmail
			}
			hasEmail = true
		case model.LoginIDKeyTypePhone:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationPhone
			}
			hasPhone = true
		case model.LoginIDKeyTypeUsername:
			if firstLoginIDIdentification == "" {
				firstLoginIDIdentification = config.AuthenticationFlowIdentificationUsername
			}
			if firstNonPhoneLoginIDIdentification == "" {
				firstNonPhoneLoginIDIdentification = config.AuthenticationFlowIdentificationUsername
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
		if firstLoginIDIdentification == config.AuthenticationFlowIdentificationPhone {
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
