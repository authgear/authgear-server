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
	// IsDisabled tells if the account is disabled for whatever reason.
	IsDisabled bool
	// AccountStatusStaleFrom tells if IsDisabled is accurate.
	// If AccountStatusStaleFrom is null, then IsDisabled is accurate.
	// If now < AccountStatusStaleFrom, then IsDisabled is accurate, else IsDisabled is stale.
	AccountStatusStaleFrom *time.Time
	// IsIndefinitelyDisabled tells if the account is disabled indefinitely.
	// If IsIndefinitelyDisabled is nullable, then an algorithm is used to set it to non-null.
	IsIndefinitelyDisabled *bool
	// IsDeactivated tells if the account is disabled via Admin API, or is disabled by the end-user.
	// If IsDeactivated is true, then the account is disabled by the end-user.
	IsDeactivated *bool
	// DisableReason is an optional string to specify the reason.
	// It can be specified via Admin API.
	DisableReason *string
	// TemporarilyDisabledFrom and TemporarilyDisabledUntil forms a temporarily disabled period.
	// Temporarily Disabled is mutually exclusive with Indefinitely Disabled.
	TemporarilyDisabledFrom  *time.Time
	TemporarilyDisabledUntil *time.Time
	// AccountValidFrom and AccountValidUntil forms account valid period.
	AccountValidFrom  *time.Time
	AccountValidUntil *time.Time
	// DeleteAt is the scheduled time when the account will be deleted.
	DeleteAt *time.Time
	// AnonymizeAt is the scheduled time when the account will be anonymized.
	AnonymizeAt *time.Time
	// AnonymizedAt is the actual time when the account was anonymized.
	AnonymizedAt *time.Time
	// IsAnonymized tells if the account is anonymized.
	IsAnonymized *bool
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
		IsDisabled:               u.IsDisabled,
		AccountStatusStaleFrom:   u.AccountStatusStaleFrom,
		IsIndefinitelyDisabled:   u.IsIndefinitelyDisabled,
		IsDeactivated:            u.IsDeactivated,
		DisableReason:            u.DisableReason,
		TemporarilyDisabledFrom:  u.TemporarilyDisabledFrom,
		TemporarilyDisabledUntil: u.TemporarilyDisabledUntil,
		AccountValidFrom:         u.AccountValidFrom,
		AccountValidUntil:        u.AccountValidUntil,
		DeleteAt:                 u.DeleteAt,
		AnonymizeAt:              u.AnonymizeAt,
		AnonymizedAt:             u.AnonymizedAt,
		IsAnonymized:             u.IsAnonymized,
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
	if user.IsDeactivated != nil && *user.IsDeactivated {
		isDeactivated = true
	}

	isAnonymized := false
	if user.IsAnonymized != nil && *user.IsAnonymized {
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
		IsDisabled:         user.IsDisabled,
		DisableReason:      user.DisableReason,
		IsDeactivated:      isDeactivated,
		DeleteAt:           user.DeleteAt,
		IsAnonymized:       isAnonymized,
		AnonymizeAt:        user.AnonymizeAt,
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
