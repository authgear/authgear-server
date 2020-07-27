package loginid

import (
	"regexp"
	"strings"

	confusable "github.com/skygeario/go-confusable-homoglyphs"
	"golang.org/x/text/secure/precis"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/validation"
)

const usernameFormat = `^[a-zA-Z0-9_\-.]*$`

var usernameRegex = regexp.MustCompile(usernameFormat)

type TypeChecker interface {
	Validate(ctx *validation.Context, loginID string)
}

type TypeCheckerFactory struct {
	Config              *config.LoginIDConfig
	ReservedNameChecker *ReservedNameChecker
}

func (f *TypeCheckerFactory) NewChecker(loginIDKeyType config.LoginIDKeyType) TypeChecker {
	switch loginIDKeyType {
	case config.LoginIDKeyTypeEmail:
		return &EmailChecker{
			Config: f.Config.Types.Email,
		}
	case config.LoginIDKeyTypeUsername:
		return &UsernameChecker{
			Config:              f.Config.Types.Username,
			ReservedNameChecker: f.ReservedNameChecker,
		}
	case config.LoginIDKeyTypePhone:
		return &PhoneChecker{}
	}

	return &NullChecker{}
}

type EmailChecker struct {
	Config *config.LoginIDEmailConfig
}

func (c *EmailChecker) Validate(ctx *validation.Context, loginID string) {
	err := validation.FormatEmail{}.CheckFormat(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "email"})
		return
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
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
	}
}

type UsernameChecker struct {
	Config              *config.LoginIDUsernameConfig
	ReservedNameChecker *ReservedNameChecker
}

func (c *UsernameChecker) Validate(ctx *validation.Context, loginID string) {
	// Ensure the login id is valid for Identifier profile
	// and use the casefolded value for checking blacklist
	// https://godoc.org/golang.org/x/text/secure/precis#NewIdentifier
	p := precis.NewIdentifier(precis.FoldCase())
	cfLoginID, err := p.String(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "username"})
		return
	}

	if *c.Config.BlockReservedUsernames {
		reserved, err := c.ReservedNameChecker.IsReserved(cfLoginID)
		if err != nil {
			ctx.AddError(err)
			return
		}
		if reserved {
			ctx.EmitErrorMessage("username is not allowed")
			return
		}
	}

	for _, item := range c.Config.ExcludedKeywords {
		cfItem, err := p.String(item)
		if err != nil {
			panic(errors.Newf("password: invalid exclude keywords: %s", item))
		}

		if strings.Contains(cfLoginID, cfItem) {
			ctx.EmitErrorMessage("username is not allowed")
			return
		}
	}

	if *c.Config.ASCIIOnly {
		if !usernameRegex.MatchString(loginID) {
			ctx.EmitError("format", map[string]interface{}{"format": "username"})
			return
		}
	}

	confusables := confusable.IsConfusable(loginID, false, []string{"LATIN", "COMMON"})
	if len(confusables) > 0 {
		ctx.EmitErrorMessage("username contains confusable characters")
	}
}

type PhoneChecker struct{}

func (c *PhoneChecker) Validate(ctx *validation.Context, loginID string) {
	err := validation.FormatPhone{}.CheckFormat(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "phone"})
	}
}

type NullChecker struct{}

func (c *NullChecker) Validate(ctx *validation.Context, loginID string) {
}
