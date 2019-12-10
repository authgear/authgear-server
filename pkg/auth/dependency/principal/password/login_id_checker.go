package password

import (
	"regexp"
	"strings"

	"golang.org/x/text/secure/precis"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

const usernameFormat = `^[a-zA-Z0-9_\-.]*$`

var usernameRegex = regexp.MustCompile(usernameFormat)

type LoginIDTypeChecker interface {
	Validate(loginID string) error
}

type LoginIDTypeCheckerFactory interface {
	NewChecker(loginIDKey config.LoginIDKeyType) LoginIDTypeChecker
}

func NewLoginIDTypeCheckerFactory(
	loginIDsKeys []config.LoginIDKeyConfiguration,
	loginIDTypes *config.LoginIDTypesConfiguration,
	reservedNameChecker *ReservedNameChecker,
) LoginIDTypeCheckerFactory {
	return &checkerFactoryImpl{
		loginIDsKeys:        loginIDsKeys,
		loginIDTypes:        loginIDTypes,
		reservedNameChecker: reservedNameChecker,
	}
}

type checkerFactoryImpl struct {
	loginIDsKeys        []config.LoginIDKeyConfiguration
	loginIDTypes        *config.LoginIDTypesConfiguration
	reservedNameChecker *ReservedNameChecker
}

func (f *checkerFactoryImpl) NewChecker(loginIDKeyType config.LoginIDKeyType) LoginIDTypeChecker {
	metadataKey, _ := loginIDKeyType.MetadataKey()
	switch metadataKey {
	case metadata.Email:
		return &LoginIDEmailChecker{
			config: f.loginIDTypes.Email,
		}
	case metadata.Username:
		return &LoginIDUsernameChecker{
			config:              f.loginIDTypes.Username,
			reservedNameChecker: f.reservedNameChecker,
		}
	case metadata.Phone:
		return &LoginIDPhoneChecker{}
	}

	return &LoginIDNullChecker{}
}

type LoginIDEmailChecker struct {
	config *config.LoginIDTypeEmailConfiguration
}

func (c *LoginIDEmailChecker) Validate(loginID string) error {
	invalidFormatError := validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
		Kind:    validation.ErrorStringFormat,
		Pointer: "/value",
		Message: "invalid login ID format",
		Details: map[string]interface{}{"format": "email"},
	}})
	ok := validation.Email{}.IsFormat(loginID)
	if !ok {
		return invalidFormatError
	}

	if *c.config.BlockPlusSign {
		// refs from stdlib
		// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
		at := strings.LastIndex(loginID, "@")
		if at < 0 {
			panic("password: malformed address, should be rejected by the email format checker")
		}

		local := loginID[:at]
		if strings.Contains(local, "+") {
			return invalidFormatError
		}
	}

	return nil
}

type LoginIDUsernameChecker struct {
	config              *config.LoginIDTypeUsernameConfiguration
	reservedNameChecker *ReservedNameChecker
}

func (c *LoginIDUsernameChecker) Validate(loginID string) error {
	invalidFormatError := validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
		Kind:    validation.ErrorStringFormat,
		Pointer: "/value",
		Message: "invalid login ID format",
		Details: map[string]interface{}{"format": "username"},
	}})

	// Ensure the login id is valid for Identifier profile
	// and use the casefolded value for checking blacklist
	// https://godoc.org/golang.org/x/text/secure/precis#NewIdentifier
	p := precis.NewIdentifier(precis.FoldCase())
	cfLoginID, err := p.String(loginID)
	if err != nil {
		return invalidFormatError
	}

	if *c.config.BlockReservedKeywords {
		reserved, err := c.reservedNameChecker.isReserved(cfLoginID)
		if err != nil {
			return err
		}
		if reserved {
			return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: "/value",
				Message: "username is not allowed",
			}})
		}
	}

	for _, item := range c.config.ExcludedKeywords {
		cfItem, err := p.String(item)
		if err != nil {
			panic(errors.Newf("password: invalid exclude keywords: %s", item))
		}

		if strings.Contains(cfLoginID, cfItem) {
			return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: "/value",
				Message: "username is not allowed",
			}})
		}
	}

	if *c.config.ASCIIOnly {
		if !usernameRegex.MatchString(loginID) {
			return invalidFormatError
		}
	}

	return nil
}

type LoginIDPhoneChecker struct{}

func (c *LoginIDPhoneChecker) Validate(loginID string) error {
	ok := validation.E164Phone{}.IsFormat(loginID)
	if ok {
		return nil
	}
	return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
		Kind:    validation.ErrorStringFormat,
		Pointer: "/value",
		Message: "invalid login ID format",
		Details: map[string]interface{}{"format": "phone"},
	}})
}

type LoginIDNullChecker struct{}

func (c *LoginIDNullChecker) Validate(loginID string) error {
	return nil
}
