package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type IdentityProvider interface {
	ListCandidates(userID string) ([]identity.Candidate, error)
}

type LoginPageTextLoginIDVariant string

const (
	LoginPageTextLoginIDVariantNone            = "none"
	LoginPageTextLoginIDVariantEamilOrUsername = "email_or_username"
	LoginPageTextLoginIDVariantEmail           = "email"
	LoginPageTextLoginIDVariantUsername        = "username"
)

type LoginPageTextLoginIDInputType string

const (
	LoginPageTextLoginIDInputTypeText  = "text"
	LoginPageTextLoginIDInputTypeEmail = "email"
)

type AuthenticationViewModel struct {
	IdentityCandidates            []identity.Candidate
	LoginPageLoginIDHasPhone      bool
	LoginPageTextLoginIDVariant   LoginPageTextLoginIDVariant
	LoginPageTextLoginIDInputType LoginPageTextLoginIDInputType
	PasswordAuthenticatorEnabled  bool
}

type AuthenticationViewModeler struct {
	Identity       IdentityProvider
	Authentication *config.AuthenticationConfig
}

func (m *AuthenticationViewModeler) ViewModel(r *http.Request) AuthenticationViewModel {
	userID := ""
	if sess := authn.GetSession(r.Context()); sess != nil {
		userID = sess.AuthnAttrs().UserID
	}

	identityCandidates, err := m.Identity.ListCandidates(userID)
	if err != nil {
		panic(err)
	}

	hasEmail := false
	hasUsername := false
	hasPhone := false
	for _, c := range identityCandidates {
		if c[identity.CandidateKeyType] == string(authn.IdentityTypeLoginID) {
			if c[identity.CandidateKeyLoginIDType] == "phone" {
				c["login_id_input_type"] = "phone"
				hasPhone = true
			} else if c[identity.CandidateKeyLoginIDType] == "email" {
				c["login_id_input_type"] = "email"
				hasEmail = true
			} else {
				c["login_id_input_type"] = "text"
				hasUsername = true
			}
		}
	}

	var loginPageTextLoginIDVariant LoginPageTextLoginIDVariant
	var loginPageTextLoginIDInputType LoginPageTextLoginIDInputType
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

	passwordAuthenticatorEnabled := false
	for _, s := range m.Authentication.PrimaryAuthenticators {
		if s == authn.AuthenticatorTypePassword {
			passwordAuthenticatorEnabled = true
		}
	}

	return AuthenticationViewModel{
		IdentityCandidates:            identityCandidates,
		LoginPageLoginIDHasPhone:      hasPhone,
		LoginPageTextLoginIDVariant:   loginPageTextLoginIDVariant,
		LoginPageTextLoginIDInputType: loginPageTextLoginIDInputType,
		PasswordAuthenticatorEnabled:  passwordAuthenticatorEnabled,
	}
}
