package facade

import (
	"errors"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
	DispatchEventImmediately(payload event.NonBlockingPayload) error
}

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(userID string) ([]*identity.Info, error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
	ListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(info *identity.Info) error
	Delete(is *identity.Info) error
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
	CheckDuplicatedByUniqueKey(info *identity.Info) (*identity.Info, error)
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
	VerifyOneWithSpec(userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *service.VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error)
	UpdateOrphans(oldInfo *identity.Info, newInfo *identity.Info) error
	RemoveOrphans(identities []*identity.Info) error
	ClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error
}

type VerificationService interface {
	GetClaims(userID string) ([]*verification.Claim, error)
	GetClaimStatus(userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error)
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	DeleteClaim(claim *verification.Claim) error
	RemoveOrphanedClaims(userID string, identities []*identity.Info, authenticators []*authenticator.Info) error
	ResetVerificationStatus(userID string) error
}

type MFAService interface {
	InvalidateAllRecoveryCode(userID string) error
	GenerateDeviceToken() string
	CreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(userID string, token string) error
	InvalidateAllDeviceTokens(userID string) error
	VerifyRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(rc *mfa.RecoveryCode) error
	GenerateRecoveryCodes() []string
	ReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
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

type RolesGroupsCommands interface {
	DeleteUserGroup(userID string) error
	DeleteUserRole(userID string) error
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
	RolesGroupsCommands        RolesGroupsCommands
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

func (c *Coordinator) IdentityListByClaimJSONPointer(pointer jsonpointer.T, value string) ([]*identity.Info, error) {
	return c.Identities.ListByClaimJSONPointer(pointer, value)
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

func (c *Coordinator) IdentityUpdate(oldInfo *identity.Info, newInfo *identity.Info) error {
	err := c.Identities.Update(newInfo)
	if err != nil {
		return err
	}

	err = c.markOAuthEmailAsVerified(newInfo)
	if err != nil {
		return err
	}

	err = c.Authenticators.UpdateOrphans(oldInfo, newInfo)
	if err != nil {
		return err
	}

	err = c.removeOrphans(newInfo.UserID)
	if err != nil {
		return err
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(newInfo.UserID)
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
		return api.NewInvariantViolated(
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
				return api.NewInvariantViolated(
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

func (c *Coordinator) IdentityCheckDuplicatedByUniqueKey(info *identity.Info) (*identity.Info, error) {
	return c.Identities.CheckDuplicatedByUniqueKey(info)
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

func (c *Coordinator) AuthenticatorVerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (verifyResult *service.VerifyResult, err error) {
	_, verifyResult, err = c.AuthenticatorVerifyOneWithSpec(
		info.UserID,
		info.Type,
		[]*authenticator.Info{info},
		spec,
		options,
	)
	return
}

func (c *Coordinator) dispatchAuthenticationFailedEvent(userID string,
	stage authn.AuthenticationStage,
	authenticationType authn.AuthenticationType) error {
	return c.Events.DispatchEventImmediately(&nonblocking.AuthenticationFailedEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
		AuthenticationStage: string(stage),
		AuthenticationType:  string(authenticationType),
	})
}

func (c *Coordinator) addAuthenticationTypeToError(err error, authnType authn.AuthenticationType) error {
	d := errorutil.Details{
		"AuthenticationType": apierrors.APIErrorDetail.Value(authnType),
	}
	newe := errorutil.WithDetails(err, d)
	return newe
}

func (c *Coordinator) AuthenticatorVerifyOneWithSpec(userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error) {
	if options == nil {
		options = &VerifyOptions{}
	}
	info, verifyResult, err = c.Authenticators.VerifyOneWithSpec(userID, authenticatorType, infos, spec, options.toServiceOptions())
	if err != nil && errors.Is(err, api.ErrInvalidCredentials) && options.AuthenticationDetails != nil {
		eventerr := c.dispatchAuthenticationFailedEvent(
			options.AuthenticationDetails.UserID,
			options.AuthenticationDetails.Stage,
			options.AuthenticationDetails.AuthenticationType,
		)
		if eventerr != nil {
			return nil, verifyResult, eventerr
		}
		err = c.addAuthenticationTypeToError(err, options.AuthenticationDetails.AuthenticationType)
	}
	return
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

	// Groups:
	if err = c.RolesGroupsCommands.DeleteUserGroup(userID); err != nil {
		return err
	}

	// Roles:
	if err = c.RolesGroupsCommands.DeleteUserRole(userID); err != nil {
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

	err = c.Events.DispatchEventOnCommit(&nonblocking.UserDeletedEventPayload{
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

	err = c.Verification.RemoveOrphanedClaims(userID, identities, authenticators)
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

func (c *Coordinator) MarkOOBIdentityVerified(info *authenticator.Info) error {
	claim := map[model.ClaimName]string{}
	switch info.Type {
	case model.AuthenticatorTypeOOBEmail:
		identities, err := c.Identities.ListByClaim(string(model.ClaimEmail), info.OOBOTP.Email)
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			return nil
		}
		claim[model.ClaimEmail] = info.OOBOTP.Email
	case model.AuthenticatorTypeOOBSMS:
		identities, err := c.Identities.ListByClaim(string(model.ClaimPhoneNumber), info.OOBOTP.Phone)
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			return nil
		}
		claim[model.ClaimPhoneNumber] = info.OOBOTP.Phone
	default:
		return nil
	}

	err := c.markVerified(info.UserID, claim)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) markOAuthEmailAsVerified(info *identity.Info) error {
	if info.Type != model.IdentityTypeOAuth {
		return nil
	}

	providerID := info.OAuth.ProviderID

	var cfg oauthrelyingparty.ProviderConfig
	for _, c := range c.IdentityConfig.OAuth.Providers {
		c := c
		if c.AsProviderConfig().ProviderID().Equal(providerID) {
			cfg = c.AsProviderConfig()
			break
		}
	}

	standardClaims := info.IdentityAwareStandardClaims()

	email, ok := standardClaims[model.ClaimEmail]
	if ok && cfg != nil {
		assumedVerified := cfg.EmailClaimConfig().AssumeVerified()
		if assumedVerified {
			// Mark as verified if OAuth email is assumed to be verified
			err := c.markVerified(info.UserID, map[model.ClaimName]string{
				model.ClaimEmail: email,
			})
			if err != nil {
				return err
			}
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

	err = c.Events.DispatchEventOnCommit(e)
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

	err = c.Events.DispatchEventOnCommit(e)
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
		err := c.Events.DispatchEventOnCommit(e)
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

	err = c.Events.DispatchEventOnCommit(e)
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

	// Groups:
	if err = c.RolesGroupsCommands.DeleteUserGroup(userID); err != nil {
		return err
	}

	// Roles:
	if err = c.RolesGroupsCommands.DeleteUserRole(userID); err != nil {
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

	err = c.Events.DispatchEventOnCommit(&nonblocking.UserAnonymizedEventPayload{
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
		err := c.Events.DispatchEventOnCommit(e)
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

	err = c.Events.DispatchEventOnCommit(e)
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

func (c *Coordinator) MarkClaimVerifiedByAdmin(claim *verification.Claim) error {
	is, err := c.IdentityListByClaim(claim.Name, claim.Value)
	if err != nil {
		return err
	}

	var userIdentity *identity.Info
	for _, i := range is {
		if i.UserID == claim.UserID {
			userIdentity = i
			break
		}
	}
	if userIdentity == nil {
		return identity.ErrIdentityNotFound
	}

	if err := c.Verification.MarkClaimVerified(claim); err != nil {
		return err
	}

	user := model.UserRef{
		Meta: model.Meta{
			ID: claim.UserID,
		},
	}
	if e, ok := nonblocking.NewIdentityVerifiedEventPayload(
		user,
		userIdentity.ToModel(),
		claim.Name,
		true,
	); ok {
		err = c.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) DeleteVerifiedClaimByAdmin(claim *verification.Claim) error {
	is, err := c.IdentityListByClaim(claim.Name, claim.Value)
	if err != nil {
		return err
	}

	var userIdentity *identity.Info
	for _, i := range is {
		if i.UserID == claim.UserID {
			userIdentity = i
			break
		}
	}
	if userIdentity == nil {
		return identity.ErrIdentityNotFound
	}

	if err := c.Verification.DeleteClaim(claim); err != nil {
		return err
	}

	user := model.UserRef{
		Meta: model.Meta{
			ID: claim.UserID,
		},
	}
	if e, ok := nonblocking.NewIdentityUnverifiedEventPayload(
		user,
		userIdentity.ToModel(),
		claim.Name,
		true,
	); ok {
		err = c.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) AuthenticatorClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	return c.Authenticators.ClearLockoutAttempts(userID, usedMethods)
}

func (c *Coordinator) MFAGenerateDeviceToken() string {
	return c.MFA.GenerateDeviceToken()
}

func (c *Coordinator) MFACreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error) {
	return c.MFA.CreateDeviceToken(userID, token)
}

func (c *Coordinator) MFAVerifyDeviceToken(userID string, token string) error {
	return c.MFA.VerifyDeviceToken(userID, token)
}

func (c *Coordinator) MFAInvalidateAllDeviceTokens(userID string) error {
	return c.MFA.InvalidateAllDeviceTokens(userID)
}

func (c *Coordinator) MFAVerifyRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error) {
	rc, err := c.MFA.VerifyRecoveryCode(userID, code)
	authnType := authn.AuthenticationTypeRecoveryCode
	if errors.Is(err, mfa.ErrRecoveryCodeNotFound) || errors.Is(err, mfa.ErrRecoveryCodeConsumed) {
		err = c.dispatchAuthenticationFailedEvent(
			userID,
			authn.AuthenticationStageSecondary,
			authnType,
		)
		if err != nil {
			return nil, err
		}
		err = c.addAuthenticationTypeToError(errors.Join(api.ErrInvalidCredentials, err), authnType)
	}
	return rc, err
}

func (c *Coordinator) MFAConsumeRecoveryCode(rc *mfa.RecoveryCode) error {
	return c.MFA.ConsumeRecoveryCode(rc)
}

func (c *Coordinator) MFAGenerateRecoveryCodes() []string {
	return c.MFA.GenerateRecoveryCodes()
}

func (c *Coordinator) MFAReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error) {
	return c.MFA.ReplaceRecoveryCodes(userID, codes)
}

func (c *Coordinator) MFAListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error) {
	return c.MFA.ListRecoveryCodes(userID)
}
