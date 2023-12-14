package reindex

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type UserQueries interface {
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type Reindexer struct {
	AppID     config.AppID
	Users     UserQueries
	OAuth     *identityoauth.Store
	LoginID   *identityloginid.Store
	TaskQueue task.Queue
}

func (s *Reindexer) ReindexUser(impl config.SearchImplementation, userID string, isDelete bool) error {
	if isDelete {
		s.TaskQueue.Enqueue(&tasks.ReindexUserParam{
			Implementation:  impl,
			DeleteUserAppID: string(s.AppID),
			DeleteUserID:    userID,
		})
		return nil
	}

	u, err := s.Users.Get(userID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}
	oauthIdentities, err := s.OAuth.List(u.ID)
	if err != nil {
		return err
	}
	loginIDIdentities, err := s.LoginID.List(u.ID)
	if err != nil {
		return err
	}

	raw := &model.SearchUserRaw{
		ID:                 u.ID,
		AppID:              string(s.AppID),
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
		LastLoginAt:        u.LastLoginAt,
		IsDisabled:         u.IsDisabled,
		StandardAttributes: u.StandardAttributes,
	}

	var arrClaims []map[model.ClaimName]string
	for _, oauthI := range oauthIdentities {
		arrClaims = append(arrClaims, oauthI.ToInfo().IdentityAwareStandardClaims())
		raw.OAuthSubjectID = append(raw.OAuthSubjectID, oauthI.ProviderSubjectID)
	}
	for _, loginIDI := range loginIDIdentities {
		arrClaims = append(arrClaims, loginIDI.ToInfo().IdentityAwareStandardClaims())
	}

	for _, claims := range arrClaims {
		if email, ok := claims[model.ClaimEmail]; ok {
			raw.Email = append(raw.Email, email)
		}
		if phoneNumber, ok := claims[model.ClaimPhoneNumber]; ok {
			raw.PhoneNumber = append(raw.PhoneNumber, phoneNumber)
		}
		if preferredUsername, ok := claims[model.ClaimPreferredUsername]; ok {
			raw.PreferredUsername = append(raw.PreferredUsername, preferredUsername)
		}
	}

	s.TaskQueue.Enqueue(&tasks.ReindexUserParam{
		Implementation: impl,
		User:           RawToSource(raw),
	})

	return nil
}

func RawToSource(raw *model.SearchUserRaw) *model.SearchUserSource {
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

	source := &model.SearchUserSource{
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
