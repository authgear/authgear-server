package reindex

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type ReindexItem struct {
	Value  *model.SearchUserSource
	Cursor model.PageCursor
}

type UserQueries interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type SourceProvider struct {
	AppID           config.AppID
	Users           UserQueries
	UserStore       *user.Store
	IdentityService *identityservice.Service
	RolesGroups     *rolesgroups.Store
}

func (s *SourceProvider) QueryPage(ctx context.Context, after model.PageCursor, first uint64) ([]ReindexItem, error) {
	users, offset, err := s.UserStore.QueryPage(ctx, user.ListOptions{}, graphqlutil.PageArgs{
		First: &first,
		After: graphqlutil.Cursor(after),
	})
	if err != nil {
		return nil, err
	}
	models := make([]ReindexItem, len(users))
	for i, u := range users {
		//nolint:gosec // G115
		i_uint64 := uint64(i)
		pageKey := db.PageKey{Offset: offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}
		source, err := s.getSource(ctx, u)
		if err != nil {
			return nil, err
		}
		models[i] = ReindexItem{Value: source, Cursor: cursor}
	}

	return models, nil
}

func (s *SourceProvider) getSource(ctx context.Context, user *user.User) (*model.SearchUserSource, error) {
	u, err := s.Users.Get(ctx, user.ID, accesscontrol.RoleGreatest)
	if err != nil {
		return nil, err
	}

	effectiveRoles, err := s.RolesGroups.ListEffectiveRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	groups, err := s.RolesGroups.ListGroupsByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	raw := &model.SearchUserRaw{
		ID:                 u.ID,
		AppID:              string(s.AppID),
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
		LastLoginAt:        u.LastLoginAt,
		IsDisabled:         u.IsDisabled,
		StandardAttributes: u.StandardAttributes,
		EffectiveRoles:     slice.Map(effectiveRoles, func(r *rolesgroups.Role) *model.Role { return r.ToModel() }),
		Groups:             slice.Map(groups, func(g *rolesgroups.Group) *model.Group { return g.ToModel() }),
	}

	arrIdentityInfo, err := s.IdentityService.ListByUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	for _, identityInfo := range arrIdentityInfo {
		claims := identityInfo.IdentityAwareStandardClaims()
		if email, ok := claims[model.ClaimEmail]; ok {
			raw.Email = append(raw.Email, email)
		}
		if phoneNumber, ok := claims[model.ClaimPhoneNumber]; ok {
			raw.PhoneNumber = append(raw.PhoneNumber, phoneNumber)
		}
		if preferredUsername, ok := claims[model.ClaimPreferredUsername]; ok {
			raw.PreferredUsername = append(raw.PreferredUsername, preferredUsername)
		}
		switch identityInfo.Type {
		case model.IdentityTypeOAuth:
			raw.OAuthSubjectID = append(raw.OAuthSubjectID, identityInfo.OAuth.ProviderSubjectID)
		case model.IdentityTypeLoginID:
			// No additional fields
		case model.IdentityTypeAnonymous:
			// No additional fields
		case model.IdentityTypeBiometric:
			// No additional fields
		case model.IdentityTypePasskey:
			// No additional fields
		case model.IdentityTypeSIWE:
			// No additional fields
		case model.IdentityTypeLDAP:
			// No additional fields
		default:
			panic(fmt.Errorf("search: unknown identity type %s", identityInfo.Type))
		}
	}

	return RawToSource(raw), nil
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

func makeStringFlatMapper[T any](stringExtractor func(T) *string) func(item T) []string {
	return func(item T) []string {
		str := stringExtractor(item)
		if str != nil {
			return []string{*str}
		}
		return []string{}
	}
}
