package loginid

import (
	"fmt"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/text/secure/precis"
	"golang.org/x/text/unicode/norm"

	"github.com/authgear/authgear-server/pkg/api/internalinterface"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type NormalizerFactory struct {
	Config *config.LoginIDConfig
}

func (f *NormalizerFactory) NormalizerWithLoginIDType(loginIDKeyType model.LoginIDKeyType) internalinterface.LoginIDNormalizer {
	switch loginIDKeyType {
	case model.LoginIDKeyTypeEmail:
		return &EmailNormalizer{
			Config: f.Config.Types.Email,
		}
	case model.LoginIDKeyTypeUsername:
		return &UsernameNormalizer{
			Config: f.Config.Types.Username,
		}
	case model.LoginIDKeyTypePhone:
		return &PhoneNumberNormalizer{}
	}

	return &NullNormalizer{}
}

type EmailNormalizer struct {
	Config *config.LoginIDEmailConfig
}

var _ internalinterface.LoginIDNormalizer = &EmailNormalizer{}

func (n *EmailNormalizer) Normalize(loginID string) (string, error) {
	// refs from stdlib
	// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
	at := strings.LastIndex(loginID, "@")
	if at < 0 {
		panic("loginid: malformed address, should be rejected by the email format checker")
	}
	local, domain := loginID[:at], loginID[at+1:]

	// convert the domain part
	var err error
	p := precis.NewFreeform(precis.FoldCase())
	domain, err = p.String(domain)
	if err != nil {
		return "", fmt.Errorf("failed to case fold email: %w", err)
	}

	// convert the local part
	local = norm.NFKC.String(local)

	if !*n.Config.CaseSensitive {
		local, err = p.String(local)
		if err != nil {
			return "", fmt.Errorf("failed to case fold email: %w", err)
		}
	}

	if *n.Config.IgnoreDotSign {
		local = strings.Replace(local, ".", "", -1)
	}

	return local + "@" + domain, nil
}

func (n *EmailNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	at := strings.LastIndex(normalizeLoginID, "@")
	if at < 0 {
		panic("loginid: malformed address, should be rejected by the email format checker")
	}
	local, domain := normalizeLoginID[:at], normalizeLoginID[at+1:]
	punycode, err := idna.ToASCII(domain)
	if err != nil {
		return "", err
	}
	domain = punycode
	return local + "@" + domain, nil
}

type UsernameNormalizer struct {
	Config *config.LoginIDUsernameConfig
}

var _ internalinterface.LoginIDNormalizer = &UsernameNormalizer{}

func (n *UsernameNormalizer) Normalize(loginID string) (string, error) {
	loginID = norm.NFKC.String(loginID)

	var err error
	if !*n.Config.CaseSensitive {
		p := precis.NewIdentifier(precis.FoldCase())
		loginID, err = p.String(loginID)
		if err != nil {
			return "", fmt.Errorf("failed to case fold username: %w", err)
		}
	}

	return loginID, nil
}

func (n *UsernameNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

type PhoneNumberNormalizer struct {
}

var _ internalinterface.LoginIDNormalizer = &PhoneNumberNormalizer{}

func (n *PhoneNumberNormalizer) Normalize(loginID string) (string, error) {
	e164, err := phone.LegalParser.ParseInputPhoneNumber(loginID)
	if err != nil {
		return "", err
	}

	return e164, nil
}

func (n *PhoneNumberNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

type NullNormalizer struct{}

var _ internalinterface.LoginIDNormalizer = &NullNormalizer{}

func (n *NullNormalizer) Normalize(loginID string) (string, error) {
	return loginID, nil
}

func (n *NullNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}
