package password

import (
	"strings"
	"unicode"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type LoginIDTypeChecker interface {
	Validate(loginID string) error
}

type LoginIDTypeCheckerFactory interface {
	NewChecker(loginIDKey string) LoginIDTypeChecker
}

func NewLoginIDTypeCheckerFactory(
	loginIDsKeys []config.LoginIDKeyConfiguration,
	loginIDTypes *config.LoginIDTypesConfiguration,
	reservedNameSourceFile string,
) LoginIDTypeCheckerFactory {
	return &checkerFactoryImpl{
		loginIDsKeys:           loginIDsKeys,
		loginIDTypes:           loginIDTypes,
		reservedNameSourceFile: reservedNameSourceFile,
	}
}

type checkerFactoryImpl struct {
	loginIDsKeys           []config.LoginIDKeyConfiguration
	loginIDTypes           *config.LoginIDTypesConfiguration
	reservedNameSourceFile string
}

func (f *checkerFactoryImpl) NewChecker(loginIDKey string) LoginIDTypeChecker {
	for _, c := range f.loginIDsKeys {
		if c.Key == loginIDKey {
			return f.newChecker(c.Type)
		}
	}

	panic("password: invalid login id key: " + loginIDKey)
}

func (f *checkerFactoryImpl) newChecker(loginIDKeyType config.LoginIDKeyType) LoginIDTypeChecker {
	metadataKey, _ := loginIDKeyType.MetadataKey()
	switch metadataKey {
	case metadata.Email:
		return &LoginIDEmailChecker{
			config: f.loginIDTypes.Email,
		}
	case metadata.Username:
		return &LoginIDUsernameChecker{
			config:                 f.loginIDTypes.Username,
			reservedNameSourceFile: f.reservedNameSourceFile,
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
		parts := strings.Split(loginID, "@")
		local := parts[0]
		if strings.Contains(local, "+") {
			return invalidFormatError
		}
	}

	return nil
}

type LoginIDUsernameChecker struct {
	config                 *config.LoginIDTypeUsernameConfiguration
	reservedNameSourceFile string
}

func (c *LoginIDUsernameChecker) Validate(loginID string) error {
	if *c.config.BlockReservedKeywords {
		checker := ReservedNameChecker{c.reservedNameSourceFile}
		reserved, err := checker.isReserved(loginID)
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
		if item == loginID {
			return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: "/value",
				Message: "username is not allowed",
			}})
		}
	}

	if *c.config.ASCIIOnly {
		for _, c := range loginID {
			if c > unicode.MaxASCII {
				return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
					Kind:    validation.ErrorStringFormat,
					Pointer: "/value",
					Message: "invalid login ID format",
					Details: map[string]interface{}{"format": "username"},
				}})
			}
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
