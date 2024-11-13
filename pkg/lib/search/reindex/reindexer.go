package reindex

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/sirupsen/logrus"
)

type ReindexRequest struct {
	UserID string `json:"user_id"`
}

type ReindexResult struct {
	UserID       string `json:"user_id"`
	IsSuccess    bool   `json:"is_success"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type UserQueries interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type ReindexerLogger struct{ *log.Logger }

func NewReindexerLogger(lf *log.Factory) *ReindexerLogger {
	return &ReindexerLogger{lf.New("search-reindexer")}
}

type ElasticsearchReindexer interface {
	ReindexUser(user *model.SearchUserSource) error
	DeleteUser(userID string) error
}

type PostgresqlReindexer interface {
	ReindexUser(ctx context.Context, user *model.SearchUserSource) error
	DeleteUser(ctx context.Context, userID string) error
}

type Reindexer struct {
	AppID           config.AppID
	SearchConfig    *config.SearchConfig
	Clock           clock.Clock
	Database        *appdb.Handle
	Logger          *ReindexerLogger
	Users           UserQueries
	UserStore       *user.Store
	IdentityService *identityservice.Service
	RolesGroups     *rolesgroups.Store

	ElasticsearchReindexer ElasticsearchReindexer
	PostgresqlReindexer    PostgresqlReindexer
}

type action string

const (
	actionReindex action = "reindex"
	actionDelete  action = "delete"
	actionSkip    action = "skip"
)

func (s *Reindexer) getSource(ctx context.Context, userID string) (*model.SearchUserSource, action, error) {
	rawUser, err := s.UserStore.Get(ctx, userID)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, actionDelete, nil
	}
	if rawUser.LastIndexedAt != nil && rawUser.RequireReindexAfter != nil && rawUser.LastIndexedAt.After(*rawUser.RequireReindexAfter) {
		// Already latest state, skip the update
		return nil, actionSkip, nil
	}

	u, err := s.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, actionDelete, nil
	}
	if err != nil {
		return nil, "", err
	}

	effectiveRoles, err := s.RolesGroups.ListEffectiveRolesByUserID(ctx, u.ID)
	if err != nil {
		return nil, "", err
	}

	groups, err := s.RolesGroups.ListGroupsByUserID(ctx, u.ID)
	if err != nil {
		return nil, "", err
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
		return nil, "", err
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
			panic(fmt.Errorf("elasticsearch: unknown identity type %s", identityInfo.Type))
		}
	}

	return RawToSource(raw), actionReindex, nil
}

func (s *Reindexer) ExecReindexUser(ctx context.Context, request ReindexRequest) (result ReindexResult) {
	failure := func(err error) ReindexResult {
		s.Logger.WithFields(map[string]interface{}{"user_id": request.UserID}).
			WithError(err).
			Error("unknown error on reindexing user")
		return ReindexResult{
			UserID:       request.UserID,
			IsSuccess:    false,
			ErrorMessage: fmt.Sprintf("%v", err),
		}
	}

	startedAt := s.Clock.NowUTC()
	var source *model.SearchUserSource = nil
	var actionToExec action
	err := s.Database.ReadOnly(ctx, func(ctx context.Context) error {
		s, a, err := s.getSource(ctx, request.UserID)
		if err != nil {
			return err
		}
		source = s
		actionToExec = a
		return nil
	})

	if err != nil {
		return failure(err)
	}

	switch actionToExec {
	case actionDelete:
		err = s.deleteUser(ctx, request.UserID)
		if err != nil {
			return failure(err)
		}

	case actionReindex:
		err = s.reindexUser(ctx, source)
		if err != nil {
			return failure(err)
		}
		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			return s.UserStore.UpdateLastIndexedAt(ctx, []string{request.UserID}, startedAt)
		})
		if err != nil {
			return failure(err)
		}

	case actionSkip:
		s.Logger.WithFields(logrus.Fields{
			"app_id":  s.AppID,
			"user_id": request.UserID,
		}).Info("skipping reindexing user because it is already up to date")
	default:
		panic(fmt.Errorf("search: unknown action %s", actionToExec))
	}

	return ReindexResult{
		UserID:    request.UserID,
		IsSuccess: true,
	}

}

func (s *Reindexer) reindexUser(ctx context.Context, source *model.SearchUserSource) error {
	switch s.SearchConfig.GetImplementation() {
	case config.SearchImplementationElasticsearch:
		return s.ElasticsearchReindexer.ReindexUser(source)
	case config.SearchImplementationPostgresql:
		return s.PostgresqlReindexer.ReindexUser(ctx, source)
	}

	return fmt.Errorf("unknown search implementation %s", s.SearchConfig.GetImplementation())
}

func (s *Reindexer) deleteUser(ctx context.Context, userID string) error {
	switch s.SearchConfig.GetImplementation() {
	case config.SearchImplementationElasticsearch:
		return s.ElasticsearchReindexer.DeleteUser(userID)
	case config.SearchImplementationPostgresql:
		return s.PostgresqlReindexer.DeleteUser(ctx, userID)
	}

	return fmt.Errorf("unknown search implementation %s", s.SearchConfig.GetImplementation())
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
