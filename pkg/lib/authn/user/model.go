package user

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type ListOptions struct {
	SortOption SortOption
}

type FilterOptions struct {
	GroupKeys []string
	RoleKeys  []string
}

func (o FilterOptions) IsFilterEnabled() bool {
	return len(o.GroupKeys) > 0 || len(o.RoleKeys) > 0
}

type SortBy string

const (
	SortByDefault     SortBy = ""
	SortByCreatedAt   SortBy = "created_at"
	SortByLastLoginAt SortBy = "last_login_at"
)

type SortOption struct {
	SortBy        SortBy
	SortDirection model.SortDirection
}

func (o SortOption) GetSortBy() SortBy {
	switch o.SortBy {
	case SortByCreatedAt:
		fallthrough
	case SortByLastLoginAt:
		return o.SortBy
	}
	return SortByCreatedAt
}

func (o SortOption) GetSortDirection() model.SortDirection {
	switch o.SortDirection {
	case model.SortDirectionAsc:
		fallthrough
	case model.SortDirectionDesc:
		return o.SortDirection
	}
	return model.SortDirectionDesc
}

func (o SortOption) Apply(builder db.SelectBuilder, after string) db.SelectBuilder {
	sortBy := o.GetSortBy()

	sortDirection := o.GetSortDirection()

	q := builder.OrderBy(fmt.Sprintf("%s %s NULLS LAST", sortBy, sortDirection))

	if after != "" {
		switch sortDirection {
		case model.SortDirectionDesc:
			q = q.Where(fmt.Sprintf("%s < ?", sortBy), after)
		case model.SortDirectionAsc:
			q = q.Where(fmt.Sprintf("%s > ?", sortBy), after)
		}
	}

	return q
}

type User struct {
	ID                   string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	MostRecentLoginAt    *time.Time
	LessRecentLoginAt    *time.Time
	StandardAttributes   map[string]interface{}
	CustomAttributes     map[string]interface{}
	LastIndexedAt        *time.Time
	RequireReindexAfter  *time.Time
	MFAGracePeriodtEndAt *time.Time
	OptOutPasskeyUpsell  bool

	// Account Status columns
	//
	// isDisabled tells if the account is disabled for whatever reason.
	isDisabled bool
	// accountStatusStaleFrom tells if IsDisabled is accurate.
	// If accountStatusStaleFrom is null, then IsDisabled is accurate.
	// If now < accountStatusStaleFrom, then IsDisabled is accurate, else IsDisabled is stale.
	accountStatusStaleFrom *time.Time
	// isIndefinitelyDisabled tells if the account is disabled indefinitely.
	// If isIndefinitelyDisabled is nullable, then an algorithm is used to set it to non-null.
	isIndefinitelyDisabled *bool
	// isDeactivated tells if the account is disabled via Admin API, or is disabled by the end-user.
	// If isDeactivated is true, then the account is disabled by the end-user.
	isDeactivated *bool
	// disableReason is an optional string to specify the reason.
	// It can be specified via Admin API.
	disableReason *string
	// temporarilyDisabledFrom and TemporarilyDisabledUntil forms a temporarily disabled period.
	// Temporarily Disabled is mutually exclusive with Indefinitely Disabled.
	temporarilyDisabledFrom  *time.Time
	temporarilyDisabledUntil *time.Time
	// accountValidFrom and AccountValidUntil forms account valid period.
	accountValidFrom  *time.Time
	accountValidUntil *time.Time
	// deleteAt is the scheduled time when the account will be deleted.
	deleteAt *time.Time
	// anonymizeAt is the scheduled time when the account will be anonymized.
	anonymizeAt *time.Time
	// anonymizedAt is the actual time when the account was anonymized.
	anonymizedAt *time.Time
	// isAnonymized tells if the account is anonymized.
	isAnonymized *bool
}

func (u *User) GetMeta() model.Meta {
	return model.Meta{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (u *User) ToRef() *model.UserRef {
	return &model.UserRef{
		Meta: u.GetMeta(),
	}
}

func (u *User) AccountStatus() AccountStatus {
	return AccountStatus{
		isDisabled:               u.isDisabled,
		accountStatusStaleFrom:   u.accountStatusStaleFrom,
		isIndefinitelyDisabled:   u.isIndefinitelyDisabled,
		isDeactivated:            u.isDeactivated,
		disableReason:            u.disableReason,
		temporarilyDisabledFrom:  u.temporarilyDisabledFrom,
		temporarilyDisabledUntil: u.temporarilyDisabledUntil,
		accountValidFrom:         u.accountValidFrom,
		accountValidUntil:        u.accountValidUntil,
		deleteAt:                 u.deleteAt,
		anonymizeAt:              u.anonymizeAt,
		anonymizedAt:             u.anonymizedAt,
		isAnonymized:             u.isAnonymized,
	}
}

func newUserModel(
	user *User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
	isVerified bool,
	derivedStandardAttributes map[string]interface{},
	customAttributes map[string]interface{},
	roles []string,
	groups []string,
) *model.User {
	if derivedStandardAttributes == nil {
		derivedStandardAttributes = make(map[string]interface{})
	}
	if customAttributes == nil {
		customAttributes = make(map[string]interface{})
	}
	if roles == nil {
		roles = make([]string, 0)
	}
	if groups == nil {
		groups = make([]string, 0)
	}

	isAnonymous := false
	for _, i := range identities {
		if i.Type == model.IdentityTypeAnonymous {
			isAnonymous = true
			break
		}
	}

	canReauthenticate := false
	for _, i := range authenticators {
		if i.Kind == authenticator.KindPrimary {
			canReauthenticate = true
		}
		if i.Kind == authenticator.KindSecondary {
			canReauthenticate = true
		}
	}

	isDeactivated := false
	if user.isDeactivated != nil && *user.isDeactivated {
		isDeactivated = true
	}

	isAnonymized := false
	if user.isAnonymized != nil && *user.isAnonymized {
		isAnonymized = true
	}

	return &model.User{
		Meta: model.Meta{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		LastLoginAt:        user.MostRecentLoginAt,
		IsAnonymous:        isAnonymous,
		IsVerified:         isVerified,
		IsDisabled:         user.isDisabled,
		DisableReason:      user.disableReason,
		IsDeactivated:      isDeactivated,
		DeleteAt:           user.deleteAt,
		IsAnonymized:       isAnonymized,
		AnonymizeAt:        user.anonymizeAt,
		CanReauthenticate:  canReauthenticate,
		StandardAttributes: derivedStandardAttributes,
		CustomAttributes:   customAttributes,
		// For backwards compatibility, we always output an empty object here.
		// The AdminAPI has marked this field non-null, so it MUST BE a map.
		Web3:                 make(map[string]interface{}),
		Roles:                roles,
		Groups:               groups,
		MFAGracePeriodtEndAt: user.MFAGracePeriodtEndAt,

		EndUserAccountID: computeEndUserAccountID(derivedStandardAttributes, identities),
	}
}

type UserForExport struct {
	model.User

	Identities     []*identity.Info
	Authenticators []*authenticator.Info
}

func computeEndUserAccountID(derivedStandardAttributes map[string]interface{}, identities []*identity.Info) string {
	var endUserAccountID string

	var ldapDisplayID string
	for _, iden := range identities {
		if iden.Type == model.IdentityTypeLDAP {
			ldapDisplayID = iden.DisplayID()
			break
		}
	}

	if s, ok := derivedStandardAttributes[string(model.ClaimEmail)].(string); ok && s != "" {
		endUserAccountID = s
	} else if s, ok := derivedStandardAttributes[string(model.ClaimPreferredUsername)].(string); ok && s != "" {
		endUserAccountID = s
	} else if s, ok := derivedStandardAttributes[string(model.ClaimPhoneNumber)].(string); ok && s != "" {
		endUserAccountID = s
	} else if ldapDisplayID != "" {
		endUserAccountID = ldapDisplayID
	}

	return endUserAccountID
}
