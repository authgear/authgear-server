package webapp

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

var phoneRegexp = regexp.MustCompile(`^\+[0-9]*$`)

func GetIdentificationOptions(f *authflowclient.FlowResponse) []authflowclient.DataIdentifyOption {
	var data authflowclient.DataIdentify
	err := authflowclient.Cast(f.Action.Data, &data)
	if err != nil {
		panic(err)
	}
	return data.Options
}

func GetMostAppropriateIdentification(f *authflowclient.FlowResponse, loginID string) authflowclient.Identification {
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
	var first authflowclient.Identification
	for _, o := range options {
		switch o.Identification {
		case authflowclient.IdentificationEmail:
			if first == "" {
				first = authflowclient.IdentificationEmail
			}
			if isEmailLike {
				return authflowclient.IdentificationEmail
			}
		case authflowclient.IdentificationPhone:
			if first == "" {
				first = authflowclient.IdentificationEmail
			}
			if isPhoneLike {
				return authflowclient.IdentificationPhone
			}
		case authflowclient.IdentificationUsername:
			if first == "" {
				first = authflowclient.IdentificationEmail
			}
		}
	}

	if first == "" {
		panic(fmt.Errorf("expected the authflow to allow login ID as identification"))
	}

	return first
}
