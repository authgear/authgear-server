package elasticsearch

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type Stats struct {
	TotalCount int
}

func makeStringFlatMapper[T any](stringExtractor func(T) *string) func(item T) []string {
	return func(item T) []string {
		str := stringExtractor(item)
		if str != nil {
			return []string{*str}
		}
		return []string{}
	}
}

func RawToSource(raw *model.ElasticsearchUserRaw) *model.ElasticsearchUserSource {
	extractString := func(attrs map[string]interface{}, key string) string {
		if attrs == nil {
			return ""
		}
		if v, ok := attrs[key].(string); ok {
			return v
		}
		return ""
	}

	extractAddressString := func(attrs map[string]interface{}, key string) string {
		if attrs == nil {
			return ""
		}
		address, ok := attrs[stdattrs.Address].(map[string]interface{})
		if !ok {
			return ""
		}
		if v, ok := address[key].(string); ok {
			return v
		}
		return ""
	}

	source := &model.ElasticsearchUserSource{
		ID:                    raw.ID,
		AppID:                 raw.AppID,
		CreatedAt:             raw.CreatedAt,
		UpdatedAt:             raw.UpdatedAt,
		LastLoginAt:           raw.LastLoginAt,
		IsDisabled:            raw.IsDisabled,
		Email:                 raw.Email,
		EmailText:             raw.Email,
		PreferredUsername:     raw.PreferredUsername,
		PreferredUsernameText: raw.PreferredUsername,
		PhoneNumber:           raw.PhoneNumber,
		PhoneNumberText:       raw.PhoneNumber,
		OAuthSubjectID:        raw.OAuthSubjectID,
		OAuthSubjectIDText:    raw.OAuthSubjectID,
		FamilyName:            extractString(raw.StandardAttributes, stdattrs.FamilyName),
		GivenName:             extractString(raw.StandardAttributes, stdattrs.GivenName),
		MiddleName:            extractString(raw.StandardAttributes, stdattrs.MiddleName),
		Name:                  extractString(raw.StandardAttributes, stdattrs.Name),
		Nickname:              extractString(raw.StandardAttributes, stdattrs.Nickname),
		Gender:                extractString(raw.StandardAttributes, stdattrs.Gender),
		Zoneinfo:              extractString(raw.StandardAttributes, stdattrs.Zoneinfo),
		Locale:                extractString(raw.StandardAttributes, stdattrs.Locale),
		Formatted:             extractAddressString(raw.StandardAttributes, stdattrs.Formatted),
		StreetAddress:         extractAddressString(raw.StandardAttributes, stdattrs.StreetAddress),
		Locality:              extractAddressString(raw.StandardAttributes, stdattrs.Locality),
		Region:                extractAddressString(raw.StandardAttributes, stdattrs.Region),
		PostalCode:            extractAddressString(raw.StandardAttributes, stdattrs.PostalCode),
		Country:               extractAddressString(raw.StandardAttributes, stdattrs.Country),
		RoleKey:               slice.Map(raw.EffectiveRoles, func(r *model.Role) string { return r.Key }),
		RoleName:              slice.FlatMap(raw.EffectiveRoles, makeStringFlatMapper(func(r *model.Role) *string { return r.Name })),
		GroupKey:              slice.Map(raw.Groups, func(g *model.Group) string { return g.Key }),
		GroupName:             slice.FlatMap(raw.Groups, makeStringFlatMapper(func(g *model.Group) *string { return g.Name })),
	}

	var emailLocalPart []string
	var emailDomain []string
	for _, email := range raw.Email {
		local, domain := mail.SplitAddress(email)
		emailLocalPart = append(emailLocalPart, local)
		emailDomain = append(emailDomain, domain)
	}

	var phoneNumberCountryCode []string
	var phoneNumberNationalNumber []string
	for _, phoneNumber := range raw.PhoneNumber {
		// For indexing into elasticsearch, we do not need to phone number to be IsPossibleNumber or IsValidNumber.
		parsed, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
		if err == nil {
			phoneNumberCountryCode = append(phoneNumberCountryCode, parsed.CountryCallingCodeWithoutPlusSign)
			phoneNumberNationalNumber = append(phoneNumberNationalNumber, parsed.NationalNumberWithoutFormatting)
		}
	}

	source.EmailLocalPart = emailLocalPart
	source.EmailLocalPartText = emailLocalPart

	source.EmailDomain = emailDomain
	source.EmailDomainText = emailDomain

	source.PhoneNumberCountryCode = phoneNumberCountryCode

	source.PhoneNumberNationalNumber = phoneNumberNationalNumber
	source.PhoneNumberNationalNumberText = phoneNumberNationalNumber

	return source
}
