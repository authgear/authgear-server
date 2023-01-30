package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(userID string) ([]*identity.Info, error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(info *identity.Info) error
	Delete(is *identity.Info) error
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
}

type AuthenticatorService interface {
	Get(id string) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	WithSpec(authenticatorInfo *authenticator.Info, spec *authenticator.Spec) (changed bool, info *authenticator.Info, err error)
	Create(authenticatorInfo *authenticator.Info) error
	Update(authenticatorInfo *authenticator.Info) error
	Delete(authenticatorInfo *authenticator.Info) error
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec) (requireUpdate bool, err error)
	RemoveOrphans(identities []*identity.Info) error
}

type VerificationService interface {
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	RemoveOrphanedClaims(identities []*identity.Info, authenticators []*authenticator.Info) error
	ResetVerificationStatus(userID string) error
}

type MFAService interface {
	InvalidateAllRecoveryCode(userID string) error
}

type UserQueries interface {
	GetRaw(userID string) (*user.User, error)
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type UserCommands interface {
	UpdateAccountStatus(userID string, accountStatus user.AccountStatus) error
	Delete(userID string) error
	Anonymize(userID string) error
}

type PasswordHistoryStore interface {
	ResetPasswordHistory(userID string) error
}

type OAuthService interface {
	ResetAll(userID string) error
}

type SessionManager interface {
	Delete(session session.Session) error
	List(userID string) ([]session.Session, error)
}

type StdAttrsService interface {
	PopulateIdentityAwareStandardAttributes(userID string) error
}

type IDPSessionManager SessionManager
type OAuthSessionManager SessionManager

// Coordinator represents interaction between identities, authenticators, and
// other high-level features (such as verification).
// FIXME(interaction): This is used to avoid circular dependency between
//
//	feature implementations. We should investigate a proper
//	resolution, as the interactions between features will
//	get complicated fast.
//
// FIXME(mfa): remove all MFA recovery code when last secondary authenticator is
//
//	removed, so that recovery codes are re-generated when setup again.
type Coordinator struct {
	Events                     EventService
	Identities                 IdentityService
	Authenticators             AuthenticatorService
	Verification               VerificationService
	MFA                        MFAService
	UserCommands               UserCommands
	UserQueries                UserQueries
	StdAttrsService            StdAttrsService
	PasswordHistory            PasswordHistoryStore
	OAuth                      OAuthService
	IDPSessions                IDPSessionManager
	OAuthSessions              OAuthSessionManager
	IdentityConfig             *config.IdentityConfig
	AccountDeletionConfig      *config.AccountDeletionConfig
	AccountAnonymizationConfig *config.AccountAnonymizationConfig
	Clock                      clock.Clock
}

func (c *Coordinator) IdentityGet(id string) (*identity.Info, error) {
	return c.Identities.Get(id)
}

func (c *Coordinator) IdentitySearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	return c.Identities.SearchBySpec(spec)
}

func (c *Coordinator) IdentityListByUser(userID string) ([]*identity.Info, error) {
	return c.Identities.ListByUser(userID)
}

func (c *Coordinator) IdentityListByClaim(name string, value string) ([]*identity.Info, error) {
	return c.Identities.ListByClaim(name, value)
}

func (c *Coordinator) IdentityNew(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return c.Identities.New(userID, spec, options)
}

func (c *Coordinator) IdentityUpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return c.Identities.UpdateWithSpec(is, spec, options)
}

func (c *Coordinator) IdentityCreate(is *identity.Info) error {
	err := c.Identities.Create(is)
	if err != nil {
		return err
	}

	err = c.markOAuthEmailAsVerified(is)
	if err != nil {
		return err
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(is.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityUpdate(info *identity.Info) error {
	err := c.Identities.Update(info)
	if err != nil {
		return err
	}

	err = c.markOAuthEmailAsVerified(info)
	if err != nil {
		return err
	}

	err = c.removeOrphans(info.UserID)
	if err != nil {
		return err
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(info.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityDelete(is *identity.Info) error {
	userID := is.UserID

	err := c.Identities.Delete(is)
	if err != nil {
		return err
	}

	err = c.removeOrphans(is.UserID)
	if err != nil {
		return err
	}

	identities, err := c.Identities.ListByUser(userID)
	if err != nil {
		return err
	}

	authenticators, err := c.Authenticators.List(userID)
	if err != nil {
		return err
	}

	// Ensure the user has at least one identifiable identity after deletion.
	remaining := identity.ApplyFilters(
		identities,
		identity.KeepIdentifiable,
	)
	if len(remaining) <= 0 {
		return NewInvariantViolated(
			"RemoveLastIdentity",
			"cannot remove last identity",
			nil,
		)
	}

	// Allow deleting anonymous identity, even if the remaining identities don't have an applicable authenticator
	// In promotion flow, we will
	// 1. Create a new identity
	// 2. Remove the anonymous identity
	// 3. Add the new authenticator
	if is.Type != model.IdentityTypeAnonymous {
		// And for each identifiable identity, if it requires primary authenticator,
		// there is at least one applicable primary authenticator remaining.
		for _, i := range remaining {
			types := i.PrimaryAuthenticatorTypes()
			if len(types) <= 0 {
				continue
			}

			ok := false
			for _, a := range authenticators {
				if a.IsApplicableTo(i) {
					ok = true
				}
			}
			if !ok {
				return NewInvariantViolated(
					"RemoveLastPrimaryAuthenticator",
					"cannot remove last primary authenticator for identity",
					map[string]interface{}{"identity_id": i.ID},
				)
			}
		}
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(is.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityCheckDuplicated(info *identity.Info) (*identity.Info, error) {
	return c.Identities.CheckDuplicated(info)
}

func (c *Coordinator) AuthenticatorGet(id string) (*authenticator.Info, error) {
	return c.Authenticators.Get(id)
}

func (c *Coordinator) AuthenticatorList(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return c.Authenticators.List(userID, filters...)
}

func (c *Coordinator) AuthenticatorNew(spec *authenticator.Spec) (*authenticator.Info, error) {
	return c.Authenticators.New(spec)
}

func (c *Coordinator) AuthenticatorNewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	return c.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
}

func (c *Coordinator) AuthenticatorWithSpec(authenticatorInfo *authenticator.Info, spec *authenticator.Spec) (changed bool, info *authenticator.Info, err error) {
	return c.Authenticators.WithSpec(authenticatorInfo, spec)
}

func (c *Coordinator) AuthenticatorCreate(authenticatorInfo *authenticator.Info, markVerified bool) error {
	err := c.Authenticators.Create(authenticatorInfo)
	if err != nil {
		return err
	}

	// Mark as verified for authenticators created by user.
	if markVerified {
		err = c.markVerified(authenticatorInfo.UserID, authenticatorInfo.StandardClaims())
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) AuthenticatorUpdate(authenticatorInfo *authenticator.Info) error {
	return c.Authenticators.Update(authenticatorInfo)
}

func (c *Coordinator) AuthenticatorDelete(authenticatorInfo *authenticator.Info) error {
	err := c.Authenticators.Delete(authenticatorInfo)
	if err != nil {
		return err
	}

	err = c.removeOrphans(authenticatorInfo.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) AuthenticatorVerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec) (requireUpdate bool, err error) {
	return c.Authenticators.VerifyWithSpec(info, spec)
}

func (c *Coordinator) UserDelete(userID string, isScheduledDeletion bool) error {
	// Delete dependents of user entity.

	// Identities:
	identities, err := c.Identities.ListByUser(userID)
	if err != nil {
		return err
	}
	for _, i := range identities {
		if err = c.Identities.Delete(i); err != nil {
			return err
		}
	}

	// Authenticators:
	authenticators, err := c.Authenticators.List(userID)
	if err != nil {
		return err
	}
	for _, a := range authenticators {
		if err = c.Authenticators.Delete(a); err != nil {
			return err
		}
	}

	// MFA recovery codes:
	if err = c.MFA.InvalidateAllRecoveryCode(userID); err != nil {
		return err
	}

	// OAuth authorizations:
	if err = c.OAuth.ResetAll(userID); err != nil {
		return err
	}

	// Verified claims:
	if err = c.Verification.ResetVerificationStatus(userID); err != nil {
		return err
	}

	// Password history:
	if err = c.PasswordHistory.ResetPasswordHistory(userID); err != nil {
		return err
	}

	// Sessions:
	if err = c.terminateAllSessions(userID); err != nil {
		return err
	}

	userModel, err := c.UserQueries.Get(userID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}

	// Finally, delete the user.
	err = c.UserCommands.Delete(userID)
	if err != nil {
		return err
	}

	err = c.Events.DispatchEvent(&nonblocking.UserDeletedEventPayload{
		UserModel:           *userModel,
		IsScheduledDeletion: isScheduledDeletion,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) removeOrphans(userID string) error {
	identities, err := c.Identities.ListByUser(userID)
	if err != nil {
		return err
	}

	err = c.Authenticators.RemoveOrphans(identities)
	if err != nil {
		return err
	}

	authenticators, err := c.Authenticators.List(userID)
	if err != nil {
		return err
	}

	err = c.Verification.RemoveOrphanedClaims(identities, authenticators)
	if err != nil {
		return err
	}

	hasSecondaryAuth := false
	for _, a := range authenticators {
		if a.Kind == authenticator.KindSecondary {
			hasSecondaryAuth = true
			break
		}
	}
	if !hasSecondaryAuth {
		err = c.MFA.InvalidateAllRecoveryCode(userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) markVerified(userID string, claims map[model.ClaimName]string) error {
	for name, value := range claims {
		name := string(name)

		claim := c.Verification.NewVerifiedClaim(userID, name, value)
		err := c.Verification.MarkClaimVerified(claim)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) markOAuthEmailAsVerified(info *identity.Info) error {
	if info.Type != model.IdentityTypeOAuth {
		return nil
	}

	providerID := info.OAuth.ProviderID

	var cfg *config.OAuthSSOProviderConfig
	for _, c := range c.IdentityConfig.OAuth.Providers {
		if c.ProviderID().Equal(&providerID) {
			c := c
			cfg = &c
			break
		}
	}

	standardClaims := info.IdentityAwareStandardClaims()

	email, ok := standardClaims[model.ClaimEmail]
	if ok && cfg != nil && *cfg.Claims.Email.AssumeVerified {
		// Mark as verified if OAuth email is assumed to be verified
		err := c.markVerified(info.UserID, map[model.ClaimName]string{
			model.ClaimEmail: email,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) terminateAllSessions(userID string) error {
	idpSessions, err := c.IDPSessions.List(userID)
	if err != nil {
		return err
	}
	for _, s := range idpSessions {
		if err = c.IDPSessions.Delete(s); err != nil {
			return err
		}
	}

	oauthSessions, err := c.OAuthSessions.List(userID)
	if err != nil {
		return err
	}
	for _, s := range oauthSessions {
		if err = c.OAuthSessions.Delete(s); err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserReenable(userID string) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	accountStatus, err := u.AccountStatus().Reenable()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	e := &nonblocking.UserReenabledEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
	}

	err = c.Events.DispatchEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserDisable(userID string, reason *string) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	accountStatus, err := u.AccountStatus().Disable(reason)
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	err = c.terminateAllSessions(userID)
	if err != nil {
		return err
	}

	e := &nonblocking.UserDisabledEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
	}

	err = c.Events.DispatchEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserScheduleDeletionByAdmin(userID string) error {
	return c.userScheduleDeletion(userID, true)
}

func (c *Coordinator) UserScheduleDeletionByEndUser(userID string) error {
	return c.userScheduleDeletion(userID, false)
}

func (c *Coordinator) userScheduleDeletion(userID string, byAdmin bool) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	deleteAt := now.Add(c.AccountDeletionConfig.GracePeriod.Duration())

	var accountStatus *user.AccountStatus
	if byAdmin {
		accountStatus, err = u.AccountStatus().ScheduleDeletionByAdmin(deleteAt)
	} else {
		accountStatus, err = u.AccountStatus().ScheduleDeletionByEndUser(deleteAt)
	}
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	err = c.terminateAllSessions(userID)
	if err != nil {
		return err
	}

	userRef := model.UserRef{
		Meta: model.Meta{
			ID: userID,
		},
	}

	events := []event.Payload{
		&blocking.UserPreScheduleDeletionBlockingEventPayload{
			UserRef:  userRef,
			AdminAPI: byAdmin,
		},
		&nonblocking.UserDeletionScheduledEventPayload{
			UserRef:  userRef,
			AdminAPI: byAdmin,
		},
	}

	for _, e := range events {
		err := c.Events.DispatchEvent(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserUnscheduleDeletionByAdmin(userID string) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	accountStatus, err := u.AccountStatus().UnscheduleDeletionByAdmin()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	e := &nonblocking.UserDeletionUnscheduledEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
		AdminAPI: true,
	}

	err = c.Events.DispatchEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserAnonymize(userID string, IsScheduledAnonymization bool) error {
	// Delete dependents of user entity.

	// Identities:
	identities, err := c.Identities.ListByUser(userID)
	if err != nil {
		return err
	}
	for _, i := range identities {
		if err = c.Identities.Delete(i); err != nil {
			return err
		}
	}

	// Authenticators:
	authenticators, err := c.Authenticators.List(userID)
	if err != nil {
		return err
	}
	for _, a := range authenticators {
		if err = c.Authenticators.Delete(a); err != nil {
			return err
		}
	}

	// MFA recovery codes:
	if err = c.MFA.InvalidateAllRecoveryCode(userID); err != nil {
		return err
	}

	// OAuth authorizations:
	if err = c.OAuth.ResetAll(userID); err != nil {
		return err
	}

	// Verified claims:
	if err = c.Verification.ResetVerificationStatus(userID); err != nil {
		return err
	}

	// Password history:
	if err = c.PasswordHistory.ResetPasswordHistory(userID); err != nil {
		return err
	}

	// Sessions:
	if err = c.terminateAllSessions(userID); err != nil {
		return err
	}

	userModel, err := c.UserQueries.Get(userID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}

	// Anonymize user record:
	err = c.UserCommands.Anonymize(userID)
	if err != nil {
		return err
	}

	err = c.Events.DispatchEvent(&nonblocking.UserAnonymizedEventPayload{
		UserModel:                *userModel,
		IsScheduledAnonymization: IsScheduledAnonymization,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserScheduleAnonymizationByAdmin(userID string) error {
	return c.userScheduleAnonymization(userID, true)
}

func (c *Coordinator) userScheduleAnonymization(userID string, byAdmin bool) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	anonymizeAt := now.Add(c.AccountAnonymizationConfig.GracePeriod.Duration())

	var accountStatus *user.AccountStatus
	if byAdmin {
		accountStatus, err = u.AccountStatus().ScheduleAnonymizationByAdmin(anonymizeAt)
	} else {
		err = errors.New("not implemented")
	}
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	err = c.terminateAllSessions(userID)
	if err != nil {
		return err
	}

	userRef := model.UserRef{
		Meta: model.Meta{
			ID: userID,
		},
	}

	events := []event.Payload{
		&blocking.UserPreScheduleAnonymizationBlockingEventPayload{
			UserRef:  userRef,
			AdminAPI: byAdmin,
		},
		&nonblocking.UserAnonymizationScheduledEventPayload{
			UserRef:  userRef,
			AdminAPI: byAdmin,
		},
	}

	for _, e := range events {
		err := c.Events.DispatchEvent(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserUnscheduleAnonymizationByAdmin(userID string) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	accountStatus, err := u.AccountStatus().UnscheduleAnonymizationByAdmin()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(userID, *accountStatus)
	if err != nil {
		return err
	}

	e := &nonblocking.UserAnonymizationUnscheduledEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
		AdminAPI: true,
	}

	err = c.Events.DispatchEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserCheckAnonymized(userID string) error {
	u, err := c.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	if u.IsAnonymized {
		return ErrUserIsAnonymized
	}

	return nil
}
