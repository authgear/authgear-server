package webapp

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

var phoneRegexp = regexp.MustCompile(`^\+[0-9]*$`)

func GetIdentificationOptions(f *authflow.FlowResponse) []declarative.IdentificationOption {
	var options []declarative.IdentificationOption
	switch data := f.Action.Data.(type) {
	case declarative.IdentificationData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}

// As IntentAccountRecoveryFlowStepIdentify has it's own IdentificationData type to narrow down Identification as {"email", "phone"},
// we imitate the same logic in GetIdentificationOptions here
func GetAccountRecoveryIdentificationOptions(f *authflow.FlowResponse) []declarative.AccountRecoveryIdentificationOption {
	var options []declarative.AccountRecoveryIdentificationOption
	switch data := f.Action.Data.(type) {
	case declarative.IntentAccountRecoveryFlowStepIdentifyData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}

func GetMostAppropriateIdentification(ctx context.Context, f *authflow.FlowResponse, loginID string, loginIDInputType string) model.AuthenticationFlowIdentification {
	// If loginIDInputType already tell us the login id type, return the corresponding type
	switch loginIDInputType {
	case "email":
		return model.AuthenticationFlowIdentificationEmail
	case "phone":
		return model.AuthenticationFlowIdentificationPhone
	}

	// Else, guess the type

	lookLikeAPhoneNumber := func(loginID string) bool {
		err := config.FormatPhone{}.CheckFormat(ctx, loginID)
		if err == nil {
			return true
		}

		if phoneRegexp.MatchString(loginID) {
			return true
		}

		return false
	}

	lookLikeAnEmailAddress := func(loginID string) bool {
		_, err := mail.ParseAddress(loginID)
		if err == nil {
			return true
		}

		if strings.Contains(loginID, "@") {
			return true
		}

		return false
	}

	isPhoneLike := lookLikeAPhoneNumber(loginID)
	isEmailLike := lookLikeAnEmailAddress(loginID)

	options := GetIdentificationOptions(f)
	var iden model.AuthenticationFlowIdentification
	for _, o := range options {
		switch o.Identification {
		case model.AuthenticationFlowIdentificationEmail:
			// If it is a email like login id, and there is an email option, it must be email
			if isEmailLike {
				iden = model.AuthenticationFlowIdentificationEmail
				break
			}
		case model.AuthenticationFlowIdentificationPhone:
			// If it is a phone like login id, and there is an phone option, it must be phone
			if isPhoneLike {
				iden = model.AuthenticationFlowIdentificationPhone
				break
			}
		case model.AuthenticationFlowIdentificationUsername:
			// The login id is not phone or email, then it can only be username
			if !isPhoneLike && !isEmailLike {
				iden = model.AuthenticationFlowIdentificationUsername
				break
			}
			// If it is like a email or phone, it can be username,
			// but we should continue the loop to see if there are better options
			if iden == "" {
				iden = model.AuthenticationFlowIdentificationUsername
			}
		}
	}

	if iden == "" {
		panic(fmt.Errorf("expected the authflow to allow login ID as identification"))
	}

	return iden
}

func GetAuthenticationOptions(f *authflow.FlowResponse) []declarative.AuthenticateOptionForOutput {
	var options []declarative.AuthenticateOptionForOutput
	switch data := f.Action.Data.(type) {
	case declarative.StepAuthenticateData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}

func GetCreateAuthenticatorOptions(f *authflow.FlowResponse) []declarative.CreateAuthenticatorOptionForOutput {
	var options []declarative.CreateAuthenticatorOptionForOutput
	switch data := f.Action.Data.(type) {
	case declarative.CreateAuthenticatorData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}
