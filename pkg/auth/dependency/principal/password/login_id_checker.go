package password

import (
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
) LoginIDTypeCheckerFactory {
	return &checkerFactoryImpl{
		loginIDsKeys: loginIDsKeys,
		loginIDTypes: loginIDTypes,
	}
}

type checkerFactoryImpl struct {
	loginIDsKeys []config.LoginIDKeyConfiguration
	loginIDTypes *config.LoginIDTypesConfiguration
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
		return &LoginIDEmailChecker{}
	case metadata.Username:
		return &LoginIDUsernameChecker{}
	case metadata.Phone:
		return &LoginIDPhoneChecker{}
	}

	return &LoginIDNullChecker{}
}

type LoginIDEmailChecker struct{}

func (c *LoginIDEmailChecker) Validate(loginID string) error {
	ok := validation.Email{}.IsFormat(loginID)
	if ok {
		return nil
	}

	return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
		Kind:    validation.ErrorStringFormat,
		Pointer: "/value",
		Message: "invalid login ID format",
		Details: map[string]interface{}{"format": "email"},
	}})
}

type LoginIDUsernameChecker struct{}

func (c *LoginIDUsernameChecker) Validate(loginID string) error {
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
