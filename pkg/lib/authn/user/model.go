package user

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

func (o SortOption) Apply(builder db.SelectBuilder) db.SelectBuilder {
	sortBy := o.SortBy
	if sortBy == SortByDefault {
		sortBy = SortByCreatedAt
	}

	sortDirection := o.SortDirection
	if sortDirection == model.SortDirectionDefault {
		sortDirection = model.SortDirectionDesc
	}

	return builder.OrderBy(fmt.Sprintf("%s %s NULLS LAST", sortBy, sortDirection))
}

var InvalidAccountStatusTransition = apierrors.Invalid.WithReason("InvalidAccountStatusTransition")

type AccountStatusType string

const (
	AccountStatusTypeNormal                         AccountStatusType = "normal"
	AccountStatusTypeDisabled                       AccountStatusType = "disabled"
	AccountStatusTypeDeactivated                    AccountStatusType = "deactivated"
	AccountStatusTypeScheduledDeletionDisabled      AccountStatusType = "scheduled_deletion_disabled"
	AccountStatusTypeScheduledDeletionDeactivated   AccountStatusType = "scheduled_deletion_deactivated"
	AccountStatusTypeAnonymized                     AccountStatusType = "anonymized"
	AccountStatusTypeScheduledAnonymizationDisabled AccountStatusType = "scheduled_anonymization_disabled"
)

// AccountStatus represents disabled, deactivated, or scheduled deletion state.
// The zero value means normal.
type AccountStatus struct {
	IsDisabled    bool
	IsDeactivated bool
	DisableReason *string
	DeleteAt      *time.Time
	IsAnonymized  bool
	AnonymizeAt   *time.Time
}

func (s AccountStatus) Type() AccountStatusType {
	if !s.IsDisabled {
		return AccountStatusTypeNormal
	}
	if s.DeleteAt != nil {
		if s.IsDeactivated {
			return AccountStatusTypeScheduledDeletionDeactivated
		}
		return AccountStatusTypeScheduledDeletionDisabled
	}
	if s.IsAnonymized {
		return AccountStatusTypeAnonymized
	}
	if s.AnonymizeAt != nil {
		return AccountStatusTypeScheduledAnonymizationDisabled
	}
	if s.IsDeactivated {
		return AccountStatusTypeDeactivated
	}
	return AccountStatusTypeDisabled
}

func (s AccountStatus) Check() error {
	// This method must be in sync with IsAccountStatusError.
	typ := s.Type()
	switch typ {
	case AccountStatusTypeNormal:
		return nil
	case AccountStatusTypeDisabled:
		return NewErrDisabledUser(s.DisableReason)
	case AccountStatusTypeDeactivated:
		return ErrDeactivatedUser
	case AccountStatusTypeAnonymized:
		return ErrAnonymizedUser
	case AccountStatusTypeScheduledDeletionDisabled:
		return NewErrScheduledDeletionByAdmin(*s.DeleteAt)
	case AccountStatusTypeScheduledDeletionDeactivated:
		return NewErrScheduledDeletionByEndUser(*s.DeleteAt)
	case AccountStatusTypeScheduledAnonymizationDisabled:
		return NewErrScheduledAnonymizationByAdmin(*s.AnonymizeAt)
	default:
		panic(fmt.Errorf("unknown account status type: %v", typ))
	}
}

func (s AccountStatus) Reenable() (*AccountStatus, error) {
	target := AccountStatus{}
	if s.Type() == AccountStatusTypeDisabled {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) Disable(reason *string) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:    true,
		DisableReason: reason,
	}
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleDeletionByEndUser(deleteAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:    true,
		IsDeactivated: true,
		DeleteAt:      &deleteAt,
	}
	if s.Type() != AccountStatusTypeNormal {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) ScheduleDeletionByAdmin(deleteAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   true,
		IsAnonymized: s.IsAnonymized,
		DeleteAt:     &deleteAt,
	}
	if s.DeleteAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleDeletionByAdmin() (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   s.IsAnonymized,
		IsAnonymized: s.IsAnonymized,
		DeleteAt:     nil,
	}
	if s.DeleteAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) Anonymize() (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:   true,
		IsAnonymized: true,
		AnonymizeAt:  s.AnonymizeAt,
	}
	if s.Type() == AccountStatusTypeNormal {
		return &target, nil
	}
	return nil, s.makeTransitionError(target.Type())
}

func (s AccountStatus) ScheduleAnonymizationByAdmin(anonymizeAt time.Time) (*AccountStatus, error) {
	target := AccountStatus{
		IsDisabled:  true,
		AnonymizeAt: &anonymizeAt,
	}
	if s.AnonymizeAt != nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) UnscheduleAnonymizationByAdmin() (*AccountStatus, error) {
	var target AccountStatus
	if s.AnonymizeAt == nil {
		return nil, s.makeTransitionError(target.Type())
	}
	return &target, nil
}

func (s AccountStatus) makeTransitionError(targetType AccountStatusType) error {
	return InvalidAccountStatusTransition.NewWithInfo(
		fmt.Sprintf("invalid account status transition: %v -> %v", s.Type(), targetType),
		map[string]interface{}{
			"from": s.Type(),
			"to":   targetType,
		},
	)
}

type User struct {
	ID                   string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	MostRecentLoginAt    *time.Time
	LessRecentLoginAt    *time.Time
	IsDisabled           bool
	DisableReason        *string
	IsDeactivated        bool
	DeleteAt             *time.Time
	IsAnonymized         bool
	AnonymizeAt          *time.Time
	StandardAttributes   map[string]interface{}
	CustomAttributes     map[string]interface{}
	LastIndexedAt        *time.Time
	RequireReindexAfter  *time.Time
	MFAGracePeriodtEndAt *time.Time
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
		IsDisabled:    u.IsDisabled,
		DisableReason: u.DisableReason,
		IsDeactivated: u.IsDeactivated,
		DeleteAt:      u.DeleteAt,
		IsAnonymized:  u.IsAnonymized,
		AnonymizeAt:   u.AnonymizeAt,
	}
}

func newUserModel(
	user *User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
	isVerified bool,
	derivedStandardAttributes map[string]interface{},
	customAttributes map[string]interface{},
	web3Info *model.UserWeb3Info,
	roles []string,
	groups []string,
) *model.User {
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

	return &model.User{
		Meta: model.Meta{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		LastLoginAt:          user.MostRecentLoginAt,
		IsAnonymous:          isAnonymous,
		IsVerified:           isVerified,
		IsDisabled:           user.IsDisabled,
		DisableReason:        user.DisableReason,
		IsDeactivated:        user.IsDeactivated,
		DeleteAt:             user.DeleteAt,
		IsAnonymized:         user.IsAnonymized,
		AnonymizeAt:          user.AnonymizeAt,
		CanReauthenticate:    canReauthenticate,
		StandardAttributes:   derivedStandardAttributes,
		CustomAttributes:     customAttributes,
		Web3:                 web3Info,
		Roles:                roles,
		Groups:               groups,
		MFAGracePeriodtEndAt: user.MFAGracePeriodtEndAt,
	}
}
