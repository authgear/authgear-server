package elasticsearch

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type Stats struct {
	TotalCount int
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
		nationalNumber, callingCode, err := phone.ParseE164ToCallingCodeAndNumber(phoneNumber)
		if err == nil {
			phoneNumberCountryCode = append(phoneNumberCountryCode, callingCode)
			phoneNumberNationalNumber = append(phoneNumberNationalNumber, nationalNumber)
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
