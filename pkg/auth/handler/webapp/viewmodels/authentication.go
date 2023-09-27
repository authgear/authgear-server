package viewmodels

import (
	"encoding/json"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type IdentityCandidatesGetter interface {
	GetIdentityCandidates() []identity.Candidate
}

type AuthenticationViewModel struct {
	IdentityCandidates     []identity.Candidate
	IdentityCount          int
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

	// q_login_id_input_type is the input the end-user has chosen.
	// It is "email", "phone" or "text".

	// LoginIDContextualType is the type the end-user thinks they should enter.
	// It depends on q_login_id_input_type.
	// It is "email", "phone", "username", or "email_or_username".
	LoginIDContextualType string

	PasskeyRequestOptionsJSON string
}

type AuthenticationViewModeler struct {
	Authentication *config.AuthenticationConfig
	LoginID        *config.LoginIDConfig
}

func (m *AuthenticationViewModeler) NewWithGraph(graph *interaction.Graph, form url.Values) AuthenticationViewModel {
	var node IdentityCandidatesGetter
	if !graph.FindLastNode(&node) {
		panic("webapp: no node with identity candidates found")
	}

	return m.NewWithCandidates(node.GetIdentityCandidates(), form)
}

func (m *AuthenticationViewModeler) NewWithCandidates(candidates []identity.Candidate, form url.Values) AuthenticationViewModel {
	hasEmail := false
	hasUsername := false
	hasPhone := false
	identityCount := 0

	// In the first loop, we first find out what type of login ID are available.
	for _, c := range candidates {
		typ, _ := c[identity.CandidateKeyType].(string)
		if typ == string(model.IdentityTypeLoginID) {
			loginIDType, _ := c[identity.CandidateKeyLoginIDType].(string)
			switch loginIDType {
			case "phone":
				hasPhone = true
			case "email":
				hasEmail = true
			default:
				hasUsername = true
			}
		}

		identityID := c[identity.CandidateKeyIdentityID].(string)
		if identityID != "" {
			identityCount++
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

	// Then we loop again and assign login_id_input_type.
	for _, c := range candidates {
		typ, _ := c[identity.CandidateKeyType].(string)
		if typ == string(model.IdentityTypeLoginID) {
			loginIDType, _ := c[identity.CandidateKeyLoginIDType].(string)
			switch loginIDType {
			case "phone":
				c["login_id_input_type"] = "phone"
			default:
				c["login_id_input_type"] = nonPhoneLoginIDInputType
			}
		}
	}

	// Then we determine q_login_id_input_type.
	xLoginIDInputType := "text"
	if _, ok := form["q_login_id_input_type"]; ok {
		xLoginIDInputType = form.Get("q_login_id_input_type")
	} else {
		if len(m.LoginID.Keys) > 0 {
			if m.LoginID.Keys[0].Type == model.LoginIDKeyTypePhone {
				xLoginIDInputType = "phone"
			} else {
				xLoginIDInputType = nonPhoneLoginIDInputType
			}
		}
	}

	var loginIDContextualType string
	switch {
	case xLoginIDInputType == "phone":
		loginIDContextualType = "phone"
	default:
		loginIDContextualType = nonPhoneLoginIDType
	}

	loginIDDisabled := !hasEmail && !hasUsername && !hasPhone

	passkeyEnabled := false
	for _, typ := range m.Authentication.Identities {
		if typ == model.IdentityTypePasskey {
			passkeyEnabled = true
		}
	}

	return AuthenticationViewModel{
		IdentityCandidates:     candidates,
		IdentityCount:          identityCount,
		LoginIDDisabled:        loginIDDisabled,
		PhoneLoginIDEnabled:    hasPhone,
		EmailLoginIDEnabled:    hasEmail,
		UsernameLoginIDEnabled: hasUsername,
		PasskeyEnabled:         passkeyEnabled,

		NonPhoneLoginIDInputType: nonPhoneLoginIDInputType,
		NonPhoneLoginIDType:      nonPhoneLoginIDType,
		LoginIDContextualType:    loginIDContextualType,
	}
}

func (m *AuthenticationViewModeler) NewWithAuthflow(f *authflow.FlowResponse, form url.Values) AuthenticationViewModel {
	options := webapp.GetIdentificationOptions(f)

	var firstLoginIDIdentification config.AuthenticationFlowIdentification
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

	xLoginIDInputType := "text"
	if _, ok := form["q_login_id_input_type"]; ok {
		xLoginIDInputType = form.Get("q_login_id_input_type")
	} else {
		if firstLoginIDIdentification != "" {
			if firstLoginIDIdentification == config.AuthenticationFlowIdentificationPhone {
				xLoginIDInputType = "phone"
			} else {
				xLoginIDInputType = nonPhoneLoginIDInputType
			}
		}
	}

	var loginIDContextualType string
	switch {
	case xLoginIDInputType == "phone":
		loginIDContextualType = "phone"
	default:
		loginIDContextualType = nonPhoneLoginIDType
	}

	loginIDDisabled := !hasEmail && !hasUsername && !hasPhone

	makeLoginIDCandidate := func(t model.LoginIDKeyType) identity.Candidate {
		candidate := identity.Candidate{
			identity.CandidateKeyIdentityID:   "",
			identity.CandidateKeyType:         string(model.IdentityTypeLoginID),
			identity.CandidateKeyLoginIDType:  string(t),
			identity.CandidateKeyLoginIDKey:   string(t),
			identity.CandidateKeyLoginIDValue: "",
			identity.CandidateKeyDisplayID:    "",
			// This is irrelevant.
			identity.CandidateKeyModifyDisabled: false,
		}
		return candidate
	}

	var candidates []identity.Candidate
	for _, o := range options {
		switch o.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeEmail))
		case config.AuthenticationFlowIdentificationPhone:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypePhone))
		case config.AuthenticationFlowIdentificationUsername:
			candidates = append(candidates, makeLoginIDCandidate(model.LoginIDKeyTypeUsername))
		case config.AuthenticationFlowIdentificationOAuth:
			candidate := identity.Candidate{
				identity.CandidateKeyIdentityID:        "",
				identity.CandidateKeyType:              string(model.IdentityTypeOAuth),
				identity.CandidateKeyProviderType:      string(o.ProviderType),
				identity.CandidateKeyProviderAlias:     o.Alias,
				identity.CandidateKeyProviderSubjectID: "",
				identity.CandidateKeyProviderAppType:   string(o.WechatAppType),
				identity.CandidateKeyDisplayID:         "",
				// This is irrelevant.
				identity.CandidateKeyModifyDisabled: false,
			}
			candidates = append(candidates, candidate)
		case config.AuthenticationFlowIdentificationPasskey:
			// Passkey was not handled by candidates.
			break
		}
	}

	return AuthenticationViewModel{
		IdentityCandidates: candidates,
		// IdentityCount is relevant only in settings.
		IdentityCount: 0,

		LoginIDDisabled:           loginIDDisabled,
		PhoneLoginIDEnabled:       hasPhone,
		EmailLoginIDEnabled:       hasEmail,
		UsernameLoginIDEnabled:    hasUsername,
		PasskeyEnabled:            passkeyEnabled,
		PasskeyRequestOptionsJSON: passkeyRequestOptionsJSON,

		NonPhoneLoginIDInputType: nonPhoneLoginIDInputType,
		NonPhoneLoginIDType:      nonPhoneLoginIDType,
		LoginIDContextualType:    loginIDContextualType,
	}
}
