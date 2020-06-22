package loginid

import (
	"regexp"
	"strings"

	confusable "github.com/skygeario/go-confusable-homoglyphs"
	"golang.org/x/text/secure/precis"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

const usernameFormat = `^[a-zA-Z0-9_\-.]*$`

var usernameRegex = regexp.MustCompile(usernameFormat)

type TypeChecker interface {
	Validate(loginID string) error
}

type TypeCheckerFactory struct {
	Config              *config.LoginIDConfig
	ReservedNameChecker *ReservedNameChecker
}

func (f *TypeCheckerFactory) NewChecker(loginIDKeyType config.LoginIDKeyType) TypeChecker {
	metadataKey, _ := loginIDKeyType.MetadataKey()
	switch metadataKey {
	case metadata.Email:
		return &EmailChecker{
			Config: f.Config.Types.Email,
		}
	case metadata.Username:
		return &UsernameChecker{
			Config:              f.Config.Types.Username,
			ReservedNameChecker: f.ReservedNameChecker,
		}
	case metadata.Phone:
		return &PhoneChecker{}
	}

	return &NullChecker{}
}

type EmailChecker struct {
	Config *config.LoginIDEmailConfig
}

func (c *EmailChecker) Validate(loginID string) error {
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

	if *c.Config.BlockPlusSign {
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

type UsernameChecker struct {
	Config              *config.LoginIDUsernameConfig
	ReservedNameChecker *ReservedNameChecker
}

func (c *UsernameChecker) Validate(loginID string) error {
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

	if *c.Config.BlockReservedUsernames {
		reserved, err := c.ReservedNameChecker.IsReserved(cfLoginID)
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

	for _, item := range c.Config.ExcludedKeywords {
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

	if *c.Config.ASCIIOnly {
		if !usernameRegex.MatchString(loginID) {
			return invalidFormatError
		}
	}

	confusables := confusable.IsConfusable(loginID, false, []string{"LATIN", "COMMON"})
	if len(confusables) > 0 {
		return validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/value",
			Message: "username contains confusable characters",
		}})
	}

	return nil
}

type PhoneChecker struct{}

func (c *PhoneChecker) Validate(loginID string) error {
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

type NullChecker struct{}

func (c *NullChecker) Validate(loginID string) error {
	return nil
}
