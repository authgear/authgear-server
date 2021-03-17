package loginid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	confusable "github.com/skygeario/go-confusable-homoglyphs"
	"golang.org/x/text/secure/precis"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/exactmatchlist"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const usernameFormat = `^[a-zA-Z0-9_\-.]*$`

var usernameRegex = regexp.MustCompile(usernameFormat)

type TypeChecker interface {
	Validate(ctx *validation.Context, loginID string)
}

type TypeCheckerFactory struct {
	Config    *config.LoginIDConfig
	Resources ResourceManager
}

func (f *TypeCheckerFactory) NewChecker(loginIDKeyType config.LoginIDKeyType) TypeChecker {
	switch loginIDKeyType {
	case config.LoginIDKeyTypeEmail:

		loginIDEmailConfig := f.Config.Types.Email

		checker := &EmailChecker{
			Config: loginIDEmailConfig,
		}

		loadDomainsList := func(desc resource.Descriptor) (*exactmatchlist.ExactMatchList, error) {
			var list *exactmatchlist.ExactMatchList
			result, err := f.Resources.Read(desc, resource.EffectiveResource{})
			if errors.Is(err, resource.ErrResourceNotFound) {
				// No domain list resources
				list = &exactmatchlist.ExactMatchList{}
			} else if err != nil {
				return nil, err
			} else {
				list = result.(*exactmatchlist.ExactMatchList)
			}
			return list, nil
		}
		if *loginIDEmailConfig.DomainBlacklistEnabled {
			domainsList, err := loadDomainsList(EmailBlacklistedDomainsTXT)
			if err != nil {
				checker.Error = err
				return checker
			}
			checker.BlacklistedDomains = domainsList
			if *loginIDEmailConfig.BlockFreeEmailProviderDomains {
				domainsList, err := loadDomainsList(FreeEmailProviderDomainsTXT)
				if err != nil {
					checker.Error = err
					return checker
				}
				checker.BlockFreeEmailProviderDomains = domainsList
			}
		} else if *loginIDEmailConfig.DomainWhitelistEnabled {
			domainsList, err := loadDomainsList(EmailWhiteListedDomainsTXT)
			if err != nil {
				checker.Error = err
				return checker
			}
			checker.WhiteListedDomains = domainsList
		}
		return checker

	case config.LoginIDKeyTypeUsername:
		checker := &UsernameChecker{
			Config: f.Config.Types.Username,
		}

		var list *blocklist.Blocklist
		result, err := f.Resources.Read(ReservedNameTXT, resource.EffectiveResource{})
		if errors.Is(err, resource.ErrResourceNotFound) {
			// No reserved usernames
			list = &blocklist.Blocklist{}
		} else if err != nil {
			checker.Error = err
			return checker
		} else {
			list = result.(*blocklist.Blocklist)
		}

		checker.ReservedNames = list
		return checker
	case config.LoginIDKeyTypePhone:
		return &PhoneChecker{}
	}

	return &NullChecker{}
}

type EmailChecker struct {
	Config *config.LoginIDEmailConfig
	// BlacklistedDomains, WhiteListedDomains and BlockFreeEmailProviderDomains
	// are provided by TypeCheckerFactory based on config, so the related
	// resources will only be loaded when it is enabled
	// EmailChecker will not further check the config before performing
	// validation
	BlacklistedDomains            *exactmatchlist.ExactMatchList
	WhiteListedDomains            *exactmatchlist.ExactMatchList
	BlockFreeEmailProviderDomains *exactmatchlist.ExactMatchList
	Error                         error
}

func (c *EmailChecker) Validate(ctx *validation.Context, loginID string) {
	if c.Error != nil {
		ctx.AddError(c.Error)
		return
	}

	err := validation.FormatEmail{}.CheckFormat(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "email"})
		return
	}

	// refs from stdlib
	// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
	at := strings.LastIndex(loginID, "@")
	if at < 0 {
		panic("password: malformed address, should be rejected by the email format checker")
	}

	local, domain := loginID[:at], loginID[at+1:]

	if *c.Config.BlockPlusSign {
		if strings.Contains(local, "+") {
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
	}

	if c.BlacklistedDomains != nil {
		matched, err := c.BlacklistedDomains.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if matched {
			ctx.EmitErrorMessage("email domain is not allowed")
		}
	}

	if c.BlockFreeEmailProviderDomains != nil {
		matched, err := c.BlockFreeEmailProviderDomains.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if matched {
			ctx.EmitErrorMessage("email domain is not allowed")
		}
	}

	if c.WhiteListedDomains != nil {
		matched, err := c.WhiteListedDomains.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if !matched {
			ctx.EmitErrorMessage("email domain is not allowed")
		}
	}
}

type UsernameChecker struct {
	Config        *config.LoginIDUsernameConfig
	ReservedNames *blocklist.Blocklist
	Error         error
}

func (c *UsernameChecker) Validate(ctx *validation.Context, loginID string) {
	if c.Error != nil {
		ctx.AddError(c.Error)
		return
	}

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
		if c.ReservedNames.IsBlocked(cfLoginID) {
			ctx.EmitErrorMessage("username is not allowed")
			return
		}
	}

	for _, item := range c.Config.ExcludedKeywords {
		cfItem, err := p.String(item)
		if err != nil {
			panic(fmt.Sprintf("password: invalid exclude keywords: %s", item))
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
