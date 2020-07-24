package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

// Ideally we should use type alias to present
// LoginPageTextLoginIDVariant and LoginPageTextLoginIDInputType
// But they may be passed to localize which does not support type alias of builtin types.

const (
	LoginPageTextLoginIDVariantNone            = "none"
	LoginPageTextLoginIDVariantEamilOrUsername = "email_or_username"
	LoginPageTextLoginIDVariantEmail           = "email"
	LoginPageTextLoginIDVariantUsername        = "username"
)

const (
	LoginPageTextLoginIDInputTypeText  = "text"
	LoginPageTextLoginIDInputTypeEmail = "email"
)

type IdentityCandidateGetter interface {
	GetIdentityCandidate() identity.Candidate
}

type AuthenticationViewModel struct {
	IdentityCandidates            []identity.Candidate
	LoginPageLoginIDHasPhone      bool
	LoginPageTextLoginIDVariant   string
	LoginPageTextLoginIDInputType string
}

func NewAuthenticationViewModel(edges []newinteraction.Edge) AuthenticationViewModel {
	var candidates []identity.Candidate
	hasEmail := false
	hasUsername := false
	hasPhone := false

	for _, edge := range edges {
		if a, ok := edge.(IdentityCandidateGetter); ok {
			candidate := a.GetIdentityCandidate()
			candidates = append(candidates, candidate)
			typ, _ := candidate[identity.CandidateKeyType].(string)
			if typ == string(authn.IdentityTypeLoginID) {
				loginIDType, _ := candidate[identity.CandidateKeyLoginIDType].(string)
				switch loginIDType {
				case "phone":
					candidate["login_id_input_type"] = "phone"
					hasPhone = true
				case "email":
					candidate["login_id_input_type"] = "email"
					hasEmail = true
				default:
					candidate["login_id_input_type"] = "text"
					hasUsername = true
				}
			}
		}
	}

	var loginPageTextLoginIDVariant string
	var loginPageTextLoginIDInputType string
	if hasEmail {
		if hasUsername {
			loginPageTextLoginIDVariant = LoginPageTextLoginIDVariantEamilOrUsername
			loginPageTextLoginIDInputType = LoginPageTextLoginIDInputTypeText
		} else {
			loginPageTextLoginIDVariant = LoginPageTextLoginIDVariantEmail
			loginPageTextLoginIDInputType = LoginPageTextLoginIDInputTypeEmail
		}
	} else {
		if hasUsername {
			loginPageTextLoginIDVariant = LoginPageTextLoginIDVariantUsername
			loginPageTextLoginIDInputType = LoginPageTextLoginIDInputTypeText
		} else {
			loginPageTextLoginIDVariant = LoginPageTextLoginIDVariantNone
			loginPageTextLoginIDInputType = LoginPageTextLoginIDInputTypeText
		}
	}

	return AuthenticationViewModel{
		IdentityCandidates:            candidates,
		LoginPageLoginIDHasPhone:      hasPhone,
		LoginPageTextLoginIDVariant:   loginPageTextLoginIDVariant,
		LoginPageTextLoginIDInputType: loginPageTextLoginIDInputType,
	}
}
