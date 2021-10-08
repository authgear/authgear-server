package loginid

import (
	"errors"
	"regexp"
	"strings"

	confusable "github.com/skygeario/go-confusable-homoglyphs"
	"golang.org/x/text/secure/precis"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/matchlist"
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

func (f *TypeCheckerFactory) NewChecker(loginIDKeyType config.LoginIDKeyType, options CheckerOptions) TypeChecker {
	// Load matchlist for validation, (e.g. doamin blocklist, allowlist, username exclude keywords...etc.)
	loadMatchlist := func(desc resource.Descriptor) (*matchlist.MatchList, error) {
		var list *matchlist.MatchList
		result, err := f.Resources.Read(desc, resource.EffectiveResource{})
		if errors.Is(err, resource.ErrResourceNotFound) {
			// No domain list resources
			list = &matchlist.MatchList{}
		} else if err != nil {
			return nil, err
		} else {
			list = result.(*matchlist.MatchList)
		}
		return list, nil
	}

	switch loginIDKeyType {
	case config.LoginIDKeyTypeEmail:

		loginIDEmailConfig := f.Config.Types.Email

		checker := &EmailChecker{
			Config: loginIDEmailConfig,
		}

		if options.EmailByPassBlocklistAllowlist {
			return checker
		}

		// blocklist and allowlist are mutually exclusive
		// block free email providers domain require blocklist enabled
		if *loginIDEmailConfig.DomainBlocklistEnabled {
			domainsList, err := loadMatchlist(EmailDomainBlockListTXT)
			if err != nil {
				checker.Error = err
				return checker
			}
			checker.DomainBlockList = domainsList
			if *loginIDEmailConfig.BlockFreeEmailProviderDomains {
				domainsList, err := loadMatchlist(FreeEmailProviderDomainsTXT)
				if err != nil {
					checker.Error = err
					return checker
				}
				checker.BlockFreeEmailProviderDomains = domainsList
			}
		} else if *loginIDEmailConfig.DomainAllowlistEnabled {
			domainsList, err := loadMatchlist(EmailDomainAllowListTXT)
			if err != nil {
				checker.Error = err
				return checker
			}
			checker.DomainAllowList = domainsList
		}
		return checker

	case config.LoginIDKeyTypeUsername:
		loginIDUsernameConfig := f.Config.Types.Username

		checker := &UsernameChecker{
			Config: loginIDUsernameConfig,
		}

		if *loginIDUsernameConfig.BlockReservedUsernames {
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
		}

		if *loginIDUsernameConfig.ExcludeKeywordsEnabled {
			excludedKeywords, err := loadMatchlist(UsernameExcludedKeywordsTXT)
			if err != nil {
				checker.Error = err
				return checker
			}

			checker.ExcludedKeywords = excludedKeywords
		}

		return checker
	case config.LoginIDKeyTypePhone:
		return &PhoneChecker{}
	}

	return &NullChecker{}
}

type EmailChecker struct {
	Config *config.LoginIDEmailConfig
	// DomainBlockList, DomainAllowList and BlockFreeEmailProviderDomains
	// are provided by TypeCheckerFactory based on config, so the related
	// resources will only be loaded when it is enabled
	// EmailChecker will not further check the config before performing
	// validation
	DomainBlockList               *matchlist.MatchList
	DomainAllowList               *matchlist.MatchList
	BlockFreeEmailProviderDomains *matchlist.MatchList
	Error                         error
}

func (c *EmailChecker) Validate(ctx *validation.Context, loginID string) {
	if c.Error != nil {
		ctx.AddError(c.Error)
		return
	}

	ctx = ctx.Child("login_id")

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

	if c.DomainBlockList != nil {
		matched, err := c.DomainBlockList.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if matched {
			ctx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainBlocklist"})
			return
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
			ctx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainBlocklist"})
			return
		}
	}

	if c.DomainAllowList != nil {
		matched, err := c.DomainAllowList.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if !matched {
			ctx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainAllowlist"})
			return
		}
	}
}

type UsernameChecker struct {
	Config *config.LoginIDUsernameConfig
	// ReservedNames and ExcludedKeywords
	// are provided by TypeCheckerFactory based on config, so the related
	// resources will only be loaded when it is enabled
	// UsernameChecker will not further check the config before performing
	// validation
	ReservedNames    *blocklist.Blocklist
	ExcludedKeywords *matchlist.MatchList
	Error            error
}

func (c *UsernameChecker) Validate(ctx *validation.Context, loginID string) {
	if c.Error != nil {
		ctx.AddError(c.Error)
		return
	}

	ctx = ctx.Child("login_id")

	// Ensure the login id is valid for Identifier profile
	// and use the casefolded value for checking blacklist
	// https://godoc.org/golang.org/x/text/secure/precis#NewIdentifier
	p := precis.NewIdentifier(precis.FoldCase())
	cfLoginID, err := p.String(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "username"})
		return
	}

	if c.ReservedNames != nil {
		if c.ReservedNames.IsBlocked(cfLoginID) {
			ctx.EmitError("blocked", map[string]interface{}{"reason": "UsernameReserved"})
			return
		}
	}

	if c.ExcludedKeywords != nil {
		matched, err := c.ExcludedKeywords.Matched(cfLoginID)
		if err != nil {
			// username cannot be fold case
			ctx.EmitError("format", map[string]interface{}{"format": "username"})
			return
		}
		if matched {
			ctx.EmitError("blocked", map[string]interface{}{"reason": "UsernameExcludedKeywords"})
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
	ctx = ctx.Child("login_id")

	err := validation.FormatPhone{}.CheckFormat(loginID)
	if err != nil {
		ctx.EmitError("format", map[string]interface{}{"format": "phone"})
	}
}

type NullChecker struct{}

func (c *NullChecker) Validate(ctx *validation.Context, loginID string) {
}
