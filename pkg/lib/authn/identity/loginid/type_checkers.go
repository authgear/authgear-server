package loginid

import (
	"context"
	"errors"
	"regexp"
	"strings"

	confusable "github.com/skygeario/go-confusable-homoglyphs"
	"golang.org/x/text/secure/precis"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/matchlist"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const usernameFormat = `^[a-zA-Z0-9_\-.]*$`

var usernameRegex = regexp.MustCompile(usernameFormat)

type TypeChecker interface {
	Validate(ctx context.Context, validationCtx *validation.Context, loginID string)
}

type TypeCheckerFactory struct {
	UIConfig      *config.UIConfig
	LoginIDConfig *config.LoginIDConfig
	Resources     ResourceManager
}

func (f *TypeCheckerFactory) NewChecker(ctx context.Context, loginIDKeyType model.LoginIDKeyType, options CheckerOptions) TypeChecker {
	switch loginIDKeyType {
	case model.LoginIDKeyTypeEmail:
		return f.makeEmailChecker(ctx, options)
	case model.LoginIDKeyTypeUsername:
		return f.makeUsernameChecker(ctx, options)
	case model.LoginIDKeyTypePhone:
		return f.makePhoneNumberChecker()
	}

	return &NullChecker{}
}

func (f *TypeCheckerFactory) loadMatchlist(ctx context.Context, desc resource.Descriptor) (*matchlist.MatchList, error) {
	// Load matchlist for validation, (e.g. doamin blocklist, allowlist, username exclude keywords...etc.)
	var list *matchlist.MatchList
	result, err := f.Resources.Read(ctx, desc, resource.EffectiveResource{})
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

func (f *TypeCheckerFactory) makeEmailChecker(ctx context.Context, options CheckerOptions) *EmailChecker {
	loginIDEmailConfig := f.LoginIDConfig.Types.Email

	checker := &EmailChecker{
		Config: loginIDEmailConfig,
	}

	if options.EmailByPassBlocklistAllowlist {
		return checker
	}

	// blocklist and allowlist are mutually exclusive
	// block free email providers domain require blocklist enabled
	if *loginIDEmailConfig.DomainBlocklistEnabled {
		domainsList, err := f.loadMatchlist(ctx, EmailDomainBlockListTXT)
		if err != nil {
			checker.Error = err
			return checker
		}
		checker.DomainBlockList = domainsList
		if *loginIDEmailConfig.BlockFreeEmailProviderDomains {
			domainsList, err := f.loadMatchlist(ctx, FreeEmailProviderDomainsTXT)
			if err != nil {
				checker.Error = err
				return checker
			}
			checker.BlockFreeEmailProviderDomains = domainsList
		}
	} else if *loginIDEmailConfig.DomainAllowlistEnabled {
		domainsList, err := f.loadMatchlist(ctx, EmailDomainAllowListTXT)
		if err != nil {
			checker.Error = err
			return checker
		}
		checker.DomainAllowList = domainsList
	}
	return checker
}

func (f *TypeCheckerFactory) makeUsernameChecker(ctx context.Context, options CheckerOptions) *UsernameChecker {
	loginIDUsernameConfig := f.LoginIDConfig.Types.Username

	checker := &UsernameChecker{
		Config: loginIDUsernameConfig,
	}

	if *loginIDUsernameConfig.BlockReservedUsernames {
		var list *blocklist.Blocklist
		result, err := f.Resources.Read(ctx, ReservedNameTXT, resource.EffectiveResource{})
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
		excludedKeywords, err := f.loadMatchlist(ctx, UsernameExcludedKeywordsTXT)
		if err != nil {
			checker.Error = err
			return checker
		}

		checker.ExcludedKeywords = excludedKeywords
	}

	return checker
}

func (f *TypeCheckerFactory) makePhoneNumberChecker() *PhoneChecker {
	var allowlist []string
	if f.UIConfig.PhoneInput != nil {
		allowlist = f.UIConfig.PhoneInput.AllowList
	}

	return &PhoneChecker{
		Alpha2AllowList: allowlist,
	}
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

func (c *EmailChecker) Validate(ctx context.Context, validationCtx *validation.Context, loginID string) {
	if c.Error != nil {
		validationCtx.AddError(c.Error)
		return
	}

	validationCtx = validationCtx.Child("login_id")

	err := validation.FormatEmail{}.CheckFormat(ctx, loginID)
	if err != nil {
		validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
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
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "BlockPlusSign"})
			return
		}
	}

	if c.DomainBlockList != nil {
		matched, err := c.DomainBlockList.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if matched {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainBlocklist"})
			return
		}
	}

	if c.BlockFreeEmailProviderDomains != nil {
		matched, err := c.BlockFreeEmailProviderDomains.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if matched {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainBlocklist"})
			return
		}
	}

	if c.DomainAllowList != nil {
		matched, err := c.DomainAllowList.Matched(domain)
		if err != nil {
			// email that the domain cannot be fold case
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
			return
		}
		if !matched {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "EmailDomainAllowlist"})
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

func (c *UsernameChecker) Validate(ctx context.Context, validationCtx *validation.Context, loginID string) {
	if c.Error != nil {
		validationCtx.AddError(c.Error)
		return
	}

	validationCtx = validationCtx.Child("login_id")

	// Ensure the login id is valid for Identifier profile
	// and use the casefolded value for checking blacklist
	// https://godoc.org/golang.org/x/text/secure/precis#NewIdentifier
	p := precis.NewIdentifier(precis.FoldCase())
	cfLoginID, err := p.String(loginID)
	if err != nil {
		validationCtx.EmitError("format", map[string]interface{}{"format": "username"})
		return
	}

	if c.ReservedNames != nil {
		if c.ReservedNames.IsBlocked(cfLoginID) {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "UsernameReserved"})
			return
		}
	}

	if c.ExcludedKeywords != nil {
		matched, err := c.ExcludedKeywords.Matched(cfLoginID)
		if err != nil {
			// username cannot be fold case
			validationCtx.EmitError("format", map[string]interface{}{"format": "username"})
			return
		}
		if matched {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "UsernameExcludedKeywords"})
			return
		}
	}

	if *c.Config.ASCIIOnly {
		if !usernameRegex.MatchString(loginID) {
			validationCtx.EmitError("format", map[string]interface{}{"format": "username"})
			return
		}
	}

	confusables := confusable.IsConfusable(loginID, false, []string{"LATIN", "COMMON"})
	if len(confusables) > 0 {
		validationCtx.EmitErrorMessage("username contains confusable characters")
	}
}

type PhoneChecker struct {
	Alpha2AllowList []string
}

func (c *PhoneChecker) Validate(ctx context.Context, validationCtx *validation.Context, loginID string) {
	validationCtx = validationCtx.Child("login_id")

	parsed, err := phone.ParsePhoneNumberWithUserInput(loginID)
	if err != nil {
		validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		return
	}

	err = config.FormatPhone{}.CheckFormat(ctx, parsed.E164)
	if err != nil {
		validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		return
	}

	if len(c.Alpha2AllowList) > 0 {
		isAllowed := false
		for _, allow := range c.Alpha2AllowList {
			// Allow the phone number if any of the possible region code is in allow list
			for _, alpha2 := range parsed.Alpha2 {
				if allow == alpha2 {
					isAllowed = true
					break
				}
			}
		}
		if !isAllowed {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "PhoneNumberCountryCodeAllowlist"})
			return
		}
	}
}

type NullChecker struct{}

func (c *NullChecker) Validate(ctx context.Context, valicationCtx *validation.Context, loginID string) {
}
