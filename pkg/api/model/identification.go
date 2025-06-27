package model

import "fmt"

type AuthenticationFlowIdentification string

const (
	AuthenticationFlowIdentificationEmail    AuthenticationFlowIdentification = "email"
	AuthenticationFlowIdentificationPhone    AuthenticationFlowIdentification = "phone"
	AuthenticationFlowIdentificationUsername AuthenticationFlowIdentification = "username"
	AuthenticationFlowIdentificationOAuth    AuthenticationFlowIdentification = "oauth"
	AuthenticationFlowIdentificationPasskey  AuthenticationFlowIdentification = "passkey"
	AuthenticationFlowIdentificationIDToken  AuthenticationFlowIdentification = "id_token"
	AuthenticationFlowIdentificationLDAP     AuthenticationFlowIdentification = "ldap"
)

func (m AuthenticationFlowIdentification) PrimaryAuthentications() []AuthenticationFlowAuthentication {
	switch m {
	case AuthenticationFlowIdentificationEmail:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationPhone:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationUsername:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationOAuth:
		// OAuth does not require primary authentication.
		return nil
	case AuthenticationFlowIdentificationPasskey:
		// Passkey does not require primary authentication.
		return nil
	case AuthenticationFlowIdentificationLDAP:
		// LDAP does not require primary authentication.
		return nil
	default:
		panic(fmt.Errorf("unknown identification: %v", m))
	}
}

func (m AuthenticationFlowIdentification) SecondaryAuthentications() []AuthenticationFlowAuthentication {
	all := []AuthenticationFlowAuthentication{
		AuthenticationFlowAuthenticationSecondaryPassword,
		AuthenticationFlowAuthenticationSecondaryTOTP,
		AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
		AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
	}
	switch m {
	case AuthenticationFlowIdentificationEmail:
		return all
	case AuthenticationFlowIdentificationPhone:
		return all
	case AuthenticationFlowIdentificationUsername:
		return all
	case AuthenticationFlowIdentificationOAuth:
		// OAuth does not require secondary authentication.
		return nil
	case AuthenticationFlowIdentificationPasskey:
		// Passkey does not require secondary authentication.
		return nil
	case AuthenticationFlowIdentificationLDAP:
		return all
	default:
		panic(fmt.Errorf("unknown identification: %v", m))
	}
}

type Identification struct {
	Identification AuthenticationFlowIdentification `json:"identification"`
	Identity       *Identity                        `json:"identity"`
	IDToken        *string                          `json:"id_token"`
}
