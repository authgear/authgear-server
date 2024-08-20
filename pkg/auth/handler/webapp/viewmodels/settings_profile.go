package viewmodels

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/labelutil"
	"github.com/authgear/authgear-server/pkg/util/territoryutil"
	"github.com/authgear/authgear-server/pkg/util/tzutil"
)

type CustomAttribute struct {
	Value          interface{}
	Label          string
	EnumValueLabel string
	Pointer        string
	Type           string
	IsEditable     bool
	Minimum        *float64
	Maximum        *float64
	Enum           []CustomAttributeEnum
}

type CustomAttributeEnum struct {
	Value string
	Label string
}

type SettingsProfileViewModel struct {
	FormattedName    string
	EndUserAccountID string
	FormattedNames   string
	Today            string

	Timezones          []tzutil.Timezone
	Alpha2             []string
	Languages          []string
	Emails             []string
	PhoneNumbers       []string
	PreferredUsernames []string

	IsReadable                    func(jsonpointer string) bool
	IsEditable                    func(jsonpointer string) bool
	GetCustomAttributeByPointer   func(jsonpointer string) *CustomAttribute
	IsStandardAttributesAllHidden bool

	Name                 string
	GivenName            string
	FamilyName           string
	MiddleName           string
	Nickname             string
	Picture              string
	Profile              string
	Website              string
	Email                string
	PhoneNumber          string
	PreferredUsername    string
	Gender               string
	Birthdate            string
	Zoneinfo             string
	ZoneinfoTimezone     *tzutil.Timezone
	Locale               string
	AddressStreetAddress string
	AddressLocality      string
	AddressRegion        string
	AddressPostalCode    string
	AddressCountry       string

	CustomAttributes []CustomAttribute
}

type SettingsProfileUserService interface {
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type SettingsProfileIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type SettingsProfileViewModeler struct {
	Localization      *config.LocalizationConfig
	UserProfileConfig *config.UserProfileConfig
	Users             SettingsProfileUserService
	Identities        SettingsProfileIdentityService
	Clock             clock.Clock
}

// nolint: gocognit
func (m *SettingsProfileViewModeler) ViewModel(userID string) (*SettingsProfileViewModel, error) {
	var emails []string
	var phoneNumbers []string
	var preferredUsernames []string
	identities, err := m.Identities.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	for _, iden := range identities {
		standardClaims := iden.IdentityAwareStandardClaims()
		if email, ok := standardClaims[model.ClaimEmail]; ok && email != "" {
			emails = append(emails, email)
		}
		if phoneNumber, ok := standardClaims[model.ClaimPhoneNumber]; ok && phoneNumber != "" {
			phoneNumbers = append(phoneNumbers, phoneNumber)
		}
		if preferredUsername, ok := standardClaims[model.ClaimPreferredUsername]; ok && preferredUsername != "" {
			preferredUsernames = append(preferredUsernames, preferredUsername)
		}
	}

	user, err := m.Users.Get(userID, config.RoleEndUser)
	if err != nil {
		return nil, err
	}
	stdAttrs := user.StandardAttributes
	str := func(key string) string {
		value, _ := stdAttrs[key].(string)
		return value
	}
	addressStr := func(key string) string {
		address, ok := stdAttrs[stdattrs.Address].(map[string]interface{})
		if !ok {
			return ""
		}

		value, _ := address[key].(string)
		return value
	}

	now := m.Clock.NowUTC()
	timezones, err := tzutil.List(now)
	if err != nil {
		return nil, err
	}

	accessControl := m.UserProfileConfig.StandardAttributes.GetAccessControl().MergedWith(
		m.UserProfileConfig.CustomAttributes.GetAccessControl(),
	)

	isReadable := func(jsonpointer string) bool {
		level := accessControl.GetLevel(
			accesscontrol.Subject(jsonpointer),
			config.RoleEndUser,
			config.AccessControlLevelHidden,
		)
		return level >= config.AccessControlLevelReadonly
	}

	isEditable := func(jsonpointer string) bool {
		level := accessControl.GetLevel(
			accesscontrol.Subject(jsonpointer),
			config.RoleEndUser,
			config.AccessControlLevelHidden,
		)
		return level == config.AccessControlLevelReadwrite
	}

	zoneinfo := str(stdattrs.Zoneinfo)
	var zoneinfoTimezone *tzutil.Timezone
	if zoneinfo != "" {
		zoneinfoTimezone, err = tzutil.AsTimezone(zoneinfo, now)
		if err != nil {
			return nil, err
		}
	}

	var customAttrs []CustomAttribute
	for _, c := range m.UserProfileConfig.CustomAttributes.Attributes {
		level := accessControl.GetLevel(
			accesscontrol.Subject(c.Pointer),
			config.RoleEndUser,
			config.AccessControlLevelHidden,
		)

		if level >= config.AccessControlLevelReadonly {
			ptr, err := jsonpointer.Parse(c.Pointer)
			if err != nil {
				return nil, err
			}

			var value interface{}
			if v, err := ptr.Traverse(user.CustomAttributes); err == nil {
				value = v
			}

			var enumValueLabel string
			var enum []CustomAttributeEnum
			if c.Type == config.CustomAttributeTypeEnum {
				if str, ok := value.(string); ok {
					enumValueLabel = labelutil.Label(str)
				}

				for _, variant := range c.Enum {
					enum = append(enum, CustomAttributeEnum{
						Value: variant,
						Label: labelutil.Label(variant),
					})
				}
			}

			customAttrs = append(customAttrs, CustomAttribute{
				Value:          value,
				Label:          labelutil.Label(ptr[0]),
				EnumValueLabel: enumValueLabel,
				Pointer:        c.Pointer,
				Type:           string(c.Type),
				IsEditable:     level >= config.AccessControlLevelReadwrite,
				Minimum:        c.Minimum,
				Maximum:        c.Maximum,
				Enum:           enum,
			})
		}
	}

	getCustomAttributeByPointer := func(jsonpointer string) *CustomAttribute {
		for _, ca := range customAttrs {
			if ca.Pointer == jsonpointer {
				out := ca
				return &out
			}
		}
		return nil
	}

	viewModel := &SettingsProfileViewModel{
		FormattedName:    stdattrs.T(stdAttrs).FormattedName(),
		EndUserAccountID: user.EndUserAccountID,
		FormattedNames:   stdattrs.T(stdAttrs).FormattedNames(),
		Today:            now.Format("2006-01-02"),

		Timezones:          timezones,
		Alpha2:             territoryutil.Alpha2,
		Languages:          m.Localization.SupportedLanguages,
		Emails:             emails,
		PhoneNumbers:       phoneNumbers,
		PreferredUsernames: preferredUsernames,

		IsReadable:                    isReadable,
		IsEditable:                    isEditable,
		GetCustomAttributeByPointer:   getCustomAttributeByPointer,
		IsStandardAttributesAllHidden: m.UserProfileConfig.StandardAttributes.IsEndUserAllHidden(),

		Name:                 str(stdattrs.Name),
		GivenName:            str(stdattrs.GivenName),
		FamilyName:           str(stdattrs.FamilyName),
		MiddleName:           str(stdattrs.MiddleName),
		Nickname:             str(stdattrs.Nickname),
		Picture:              str(stdattrs.Picture),
		Profile:              str(stdattrs.Profile),
		Website:              str(stdattrs.Website),
		Email:                str(stdattrs.Email),
		PhoneNumber:          str(stdattrs.PhoneNumber),
		PreferredUsername:    str(stdattrs.PreferredUsername),
		Gender:               str(stdattrs.Gender),
		Birthdate:            str(stdattrs.Birthdate),
		Zoneinfo:             zoneinfo,
		ZoneinfoTimezone:     zoneinfoTimezone,
		Locale:               str(stdattrs.Locale),
		AddressStreetAddress: addressStr(stdattrs.StreetAddress),
		AddressLocality:      addressStr(stdattrs.Locality),
		AddressRegion:        addressStr(stdattrs.Region),
		AddressPostalCode:    addressStr(stdattrs.PostalCode),
		AddressCountry:       addressStr(stdattrs.Country),

		CustomAttributes: customAttrs,
	}

	return viewModel, nil
}
