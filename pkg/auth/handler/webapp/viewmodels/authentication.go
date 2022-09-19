package viewmodels

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
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

	// x_login_id_input_type is the input the end-user has chosen.
	// It is "email", "phone" or "text".

	// LoginIDContextualType is the type the end-user thinks they should enter.
	// It depends on x_login_id_input_type.
	// It is "email", "phone", "username", or "email_or_username".
	LoginIDContextualType string
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

	// Then we determine x_login_id_input_type.
	xLoginIDInputType := "text"
	if _, ok := form["x_login_id_input_type"]; ok {
		xLoginIDInputType = form.Get("x_login_id_input_type")
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
