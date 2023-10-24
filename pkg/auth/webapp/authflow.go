package webapp

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

var phoneRegexp = regexp.MustCompile(`^\+[0-9]*$`)

func GetIdentificationOptions(f *authflow.FlowResponse) []declarative.IdentificationOption {
	var options []declarative.IdentificationOption
	switch data := f.Action.Data.(type) {
	case declarative.IntentLoginFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentSignupFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentPromoteFlowStepIdentifyData:
		options = data.Options
	case declarative.IntentSignupLoginFlowStepIdentifyData:
		options = data.Options
	default:
		panic(fmt.Errorf("unexpected type of data: %T", f.Action.Data))
	}
	return options
}

func GetMostAppropriateIdentification(f *authflow.FlowResponse, loginID string) config.AuthenticationFlowIdentification {
	lookLikeAPhoneNumber := func(loginID string) bool {
		err := phone.EnsureE164(loginID)
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
	var first config.AuthenticationFlowIdentification
	for _, o := range options {
		switch o.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			if first == "" {
				first = config.AuthenticationFlowIdentificationEmail
			}
			if isEmailLike {
				return config.AuthenticationFlowIdentificationEmail
			}
		case config.AuthenticationFlowIdentificationPhone:
			if first == "" {
				first = config.AuthenticationFlowIdentificationEmail
			}
			if isPhoneLike {
				return config.AuthenticationFlowIdentificationPhone
			}
		case config.AuthenticationFlowIdentificationUsername:
			if first == "" {
				first = config.AuthenticationFlowIdentificationEmail
			}
		}
	}

	if first == "" {
		panic(fmt.Errorf("expected the authflow to allow login ID as identification"))
	}

	return first
}
