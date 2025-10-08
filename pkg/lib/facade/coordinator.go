package facade

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type IdentityService interface {
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Normalize(ctx context.Context, typ model.LoginIDKeyType, value string) (normalized string, uniqueKey string, err error)

	Get(ctx context.Context, id string) (*identity.Info, error)
	SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error)
	ListRefsByUsers(ctx context.Context, userIDs []string, identityType *model.IdentityType) ([]*model.IdentityRef, error)
	Create(ctx context.Context, is *identity.Info) error
	Update(ctx context.Context, info *identity.Info) error
	Delete(ctx context.Context, is *identity.Info) error
	CheckDuplicated(ctx context.Context, info *identity.Info) (*identity.Info, error)
	CheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (*identity.Info, error)
	AdminAPIGetByLoginIDKeyAndLoginIDValue(ctx context.Context, loginIDKey string, loginIDValue string) (*identity.Info, error)
	AdminAPIGetByOAuthAliasAndSubject(ctx context.Context, alias string, subjectID string) (*identity.Info, error)
}

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)

	Get(ctx context.Context, id string) (*authenticator.Info, error)
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Create(ctx context.Context, authenticatorInfo *authenticator.Info) error
	Update(ctx context.Context, authenticatorInfo *authenticator.Info) error
	Delete(ctx context.Context, authenticatorInfo *authenticator.Info) error
	VerifyOneWithSpec(ctx context.Context, userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *service.VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error)
	UpdateOrphans(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error
	RemoveOrphans(ctx context.Context, identities []*identity.Info) error
	ClearLockoutAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error
}

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim

	GetClaims(ctx context.Context, userID string) ([]*verification.Claim, error)
	GetClaimStatus(ctx context.Context, userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error)
	GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error)
	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
	DeleteClaim(ctx context.Context, claim *verification.Claim) error
	RemoveOrphanedClaims(ctx context.Context, userID string, identities []*identity.Info, authenticators []*authenticator.Info) error
	ResetVerificationStatus(ctx context.Context, userID string) error
}

type MFAService interface {
	GenerateDeviceToken(ctx context.Context) string
	GenerateRecoveryCodes(ctx context.Context) []string

	InvalidateAllRecoveryCode(ctx context.Context, userID string) error
	CreateDeviceToken(ctx context.Context, userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(ctx context.Context, userID string, token string) error
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
	VerifyRecoveryCode(ctx context.Context, userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(ctx context.Context, rc *mfa.RecoveryCode) error
	ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
}

type SendPasswordService interface {
	Send(ctx context.Context, userID string, password string, msgType translation.MessageType) error
}

type UserQueries interface {
	GetRaw(ctx context.Context, userID string) (*user.User, error)
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type UserCommands interface {
	Create(ctx context.Context, userID string) (*user.User, error)
	UpdateAccountStatus(ctx context.Context, userID string, accountStatus user.AccountStatusWithRefTime) error
	UpdateMFAEnrollment(ctx context.Context, userID string, gracePeriodEndAt *time.Time) error
	Delete(ctx context.Context, userID string) error
	Anonymize(ctx context.Context, userID string) error
	AfterCreate(
		ctx context.Context,
		user *user.User,
		identities []*identity.Info,
		authenticators []*authenticator.Info,
		isAdminAPI bool,
	) error
}

type RolesGroupsCommands interface {
	DeleteUserGroup(ctx context.Context, userID string) error
	DeleteUserRole(ctx context.Context, userID string) error
}

type PasswordHistoryStore interface {
	ResetPasswordHistory(ctx context.Context, userID string) error
}

type OAuthService interface {
	ResetAll(ctx context.Context, userID string) error
}

type SessionManager interface {
	Delete(ctx context.Context, session session.ListableSession) error
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	CleanUpForDeletingUserID(ctx context.Context, userID string) error
}

type StdAttrsService interface {
	PopulateIdentityAwareStandardAttributes(ctx context.Context, userID string) error
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
	SendPassword               SendPasswordService
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
	AuthenticationConfig       *config.AuthenticationConfig
	Clock                      clock.Clock
	PasswordGenerator          *password.Generator
}

func (c *Coordinator) IdentityGet(ctx context.Context, id string) (*identity.Info, error) {
	return c.Identities.Get(ctx, id)
}

func (c *Coordinator) IdentitySearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	return c.Identities.SearchBySpec(ctx, spec)
}

func (c *Coordinator) IdentityListByUser(ctx context.Context, userID string) ([]*identity.Info, error) {
	return c.Identities.ListByUser(ctx, userID)
}

func (c *Coordinator) IdentityListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error) {
	return c.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, userID)
}

func (c *Coordinator) IdentityListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error) {
	return c.Identities.ListByClaim(ctx, name, value)
}

func (c *Coordinator) IdentityListRefsByUsers(ctx context.Context, userIDs []string, identityType *model.IdentityType) ([]*model.IdentityRef, error) {
	return c.Identities.ListRefsByUsers(ctx, userIDs, identityType)
}

func (c *Coordinator) IdentityNew(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return c.Identities.New(ctx, userID, spec, options)
}

func (c *Coordinator) IdentityUpdateWithSpec(ctx context.Context, is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	return c.Identities.UpdateWithSpec(ctx, is, spec, options)
}

func (c *Coordinator) IdentityCreate(ctx context.Context, is *identity.Info) error {
	err := c.Identities.Create(ctx, is)
	if err != nil {
		return err
	}

	err = c.markOAuthEmailAsVerified(ctx, is)
	if err != nil {
		return err
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(ctx, is.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityCreateByAdmin(ctx context.Context, userID string, spec *identity.Spec, password string) (*identity.Info, error) {
	iden, err := c.Identities.New(ctx, userID, spec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	if _, err := c.Identities.CheckDuplicated(ctx, iden); err != nil {
		return nil, err
	}

	if err := c.Identities.Create(ctx, iden); err != nil {
		return nil, err
	}

	// Remove any anonymous identity this user has.
	if err := c.removeAnonymousIdentitiesOfUser(ctx, userID); err != nil {
		return nil, err
	}

	if err := c.StdAttrsService.PopulateIdentityAwareStandardAttributes(ctx, userID); err != nil {
		return nil, err
	}

	if err := c.createPrimaryAuthenticatorsForAdminAPICreateIdentity(ctx, iden, password); err != nil {
		return nil, err
	}

	// Dispatch event
	isAdminAPI := true
	var e event.Payload
	switch iden.Type {
	case model.IdentityTypeLoginID:
		loginIDType := iden.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDAddedEventPayload(
			model.UserRef{
				Meta: model.Meta{
					ID: iden.UserID,
				},
			},
			iden.ToModel(),
			string(loginIDType),
			isAdminAPI,
		); ok {
			e = payload
		}
	default:
		panic(fmt.Errorf("unexpected identity type: %v", iden.Type))
	}

	if e != nil {
		err = c.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return nil, err
		}
	}

	return iden, nil
}

func (c *Coordinator) removeAnonymousIdentitiesOfUser(ctx context.Context, userID string) (err error) {
	idens, err := c.Identities.ListByUser(ctx, userID)
	if err != nil {
		return
	}

	anonymousIdentities := identity.ApplyFilters(
		idens,
		identity.KeepType(model.IdentityTypeAnonymous),
	)
	for _, iden := range anonymousIdentities {
		err = c.Identities.Delete(ctx, iden)
		if err != nil {
			return
		}
	}

	return
}

// nolint: gocognit
func (c *Coordinator) createPrimaryAuthenticatorsForAdminAPICreateIdentity(ctx context.Context, iden *identity.Info, password string) (err error) {
	authenticatorTypes := *c.AuthenticationConfig.PrimaryAuthenticators
	var authenticatorSpecs []*authenticator.Spec
	for _, t := range authenticatorTypes {
		switch t {
		case model.AuthenticatorTypePassword:
			if password != "" {
				var passwords []*authenticator.Info
				passwords, err = c.Authenticators.List(ctx,
					iden.UserID,
					authenticator.KeepKind(authenticator.KindPrimary),
					authenticator.KeepType(model.AuthenticatorTypePassword),
				)
				if err != nil {
					return
				}
				// We only create the password if the user does not have one already.
				if len(passwords) == 0 {
					var passwordSpec *authenticator.Spec
					opts := CreatePasswordOptions{
						Password:           &password,
						SendPassword:       false,
						SetPasswordExpired: false,
					}
					passwordSpec, err = c.createPasswordAuthenticatorSpec(iden.UserID, opts)
					if err != nil {
						return
					}
					if passwordSpec != nil {
						authenticatorSpecs = append(authenticatorSpecs, passwordSpec)
					}
				}
			}
		case model.AuthenticatorTypeOOBSMS:
			if iden.Type == model.IdentityTypeLoginID && iden.LoginID.LoginIDType == model.LoginIDKeyTypePhone {
				spec, ok := c.createOOBSMSAuthenticatorSpec(iden, iden.UserID)
				if ok {
					authenticatorSpecs = append(authenticatorSpecs, spec)
				}
			}
		case model.AuthenticatorTypeOOBEmail:
			if iden.Type == model.IdentityTypeLoginID && iden.LoginID.LoginIDType == model.LoginIDKeyTypeEmail {
				spec, ok := c.createOOBEmailAuthenticatorSpec(iden, iden.UserID)
				if ok {
					authenticatorSpecs = append(authenticatorSpecs, spec)
				}
			}
		}
	}

	_, err = c.createAuthenticatorInfos(ctx, authenticatorSpecs)
	if err != nil {
		return
	}

	return nil
}

func (c *Coordinator) IdentityUpdate(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error {
	err := c.Identities.Update(ctx, newInfo)
	if err != nil {
		return err
	}

	err = c.markOAuthEmailAsVerified(ctx, newInfo)
	if err != nil {
		return err
	}

	err = c.Authenticators.UpdateOrphans(ctx, oldInfo, newInfo)
	if err != nil {
		return err
	}

	err = c.removeOrphans(ctx, newInfo.UserID)
	if err != nil {
		return err
	}

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(ctx, newInfo.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityDelete(ctx context.Context, is *identity.Info) error {
	userID := is.UserID

	err := c.Identities.Delete(ctx, is)
	if err != nil {
		return err
	}

	err = c.removeOrphans(ctx, is.UserID)
	if err != nil {
		return err
	}

	identities, err := c.Identities.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	authenticators, err := c.Authenticators.List(ctx, userID)
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

	err = c.StdAttrsService.PopulateIdentityAwareStandardAttributes(ctx, is.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityCheckDuplicated(ctx context.Context, info *identity.Info) (*identity.Info, error) {
	return c.Identities.CheckDuplicated(ctx, info)
}

func (c *Coordinator) IdentityCheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (*identity.Info, error) {
	return c.Identities.CheckDuplicatedByUniqueKey(ctx, info)
}

func (c *Coordinator) AuthenticatorGet(ctx context.Context, id string) (*authenticator.Info, error) {
	return c.Authenticators.Get(ctx, id)
}

func (c *Coordinator) AuthenticatorList(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return c.Authenticators.List(ctx, userID, filters...)
}

func (c *Coordinator) AuthenticatorNew(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error) {
	return c.Authenticators.New(ctx, spec)
}

func (c *Coordinator) AuthenticatorNewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	return c.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, spec)
}

func (c *Coordinator) AuthenticatorCreate(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error {
	err := c.Authenticators.Create(ctx, authenticatorInfo)
	if err != nil {
		return err
	}

	// Mark as verified for authenticators created by user.
	if markVerified {
		err = c.markVerified(ctx, authenticatorInfo.UserID, authenticatorInfo.StandardClaims())
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) AuthenticatorUpdate(ctx context.Context, authenticatorInfo *authenticator.Info) error {
	return c.Authenticators.Update(ctx, authenticatorInfo)
}

func (c *Coordinator) AuthenticatorUpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error) {
	return c.Authenticators.UpdatePassword(ctx, authenticatorInfo, options)
}

func (c *Coordinator) AuthenticatorDelete(ctx context.Context, authenticatorInfo *authenticator.Info) error {
	err := c.Authenticators.Delete(ctx, authenticatorInfo)
	if err != nil {
		return err
	}

	err = c.removeOrphans(ctx, authenticatorInfo.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) AuthenticatorVerifyWithSpec(ctx context.Context, info *authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (verifyResult *service.VerifyResult, err error) {
	_, verifyResult, err = c.AuthenticatorVerifyOneWithSpec(
		ctx,
		info.UserID,
		info.Type,
		[]*authenticator.Info{info},
		spec,
		options,
	)
	return
}

func (c *Coordinator) dispatchAuthenticationFailedEvent(ctx context.Context, userID string,
	stage authn.AuthenticationStage,
	authenticationType authn.AuthenticationType) error {
	return c.Events.DispatchEventImmediately(ctx, &nonblocking.AuthenticationFailedEventPayload{
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

func (c *Coordinator) AuthenticatorVerifyOneWithSpec(ctx context.Context, userID string, authenticatorType model.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error) {
	if options == nil {
		options = &VerifyOptions{}
	}
	info, verifyResult, err = c.Authenticators.VerifyOneWithSpec(ctx, userID, authenticatorType, infos, spec, options.toServiceOptions())
	if err != nil && errors.Is(err, api.ErrInvalidCredentials) && options.AuthenticationDetails != nil {
		eventerr := c.dispatchAuthenticationFailedEvent(
			ctx,
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

type CreatePasswordOptions struct {
	Password           *string
	SendPassword       bool
	SetPasswordExpired bool
}

func (c *Coordinator) UserCreatebyAdmin(ctx context.Context, identitySpec *identity.Spec, opts CreatePasswordOptions) (*user.User, error) {
	// 1. Create user
	userID := uuid.New()
	user, err := c.UserCommands.Create(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Create identity
	identityInfo, err := c.Identities.New(ctx, userID, identitySpec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	if _, err := c.Identities.CheckDuplicated(ctx, identityInfo); err != nil {
		return nil, err
	}

	if err := c.Identities.Create(ctx, identityInfo); err != nil {
		return nil, err
	}

	if err := c.StdAttrsService.PopulateIdentityAwareStandardAttributes(ctx, userID); err != nil {
		return nil, err
	}

	// 3. Create primary authenticators
	authenticatorInfos, err := c.createPrimaryAuthenticatorsForAdminAPICreateUser(ctx, identityInfo, userID, opts)
	if err != nil {
		return nil, err
	}

	if err := c.UserCommands.AfterCreate(ctx,
		user,
		[]*identity.Info{identityInfo},
		authenticatorInfos,
		true,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Coordinator) createPrimaryAuthenticatorsForAdminAPICreateUser(ctx context.Context, identityInfo *identity.Info, userID string, opts CreatePasswordOptions) ([]*authenticator.Info, error) {
	authenticatorTypes := *c.AuthenticationConfig.PrimaryAuthenticators
	if len(authenticatorTypes) == 0 {
		return nil, api.InvalidConfiguration.New("identity requires primary authenticator but none is enabled")
	}

	var passwordSpec *authenticator.Spec
	var authenticatorSpecs []*authenticator.Spec
	for _, t := range authenticatorTypes {
		switch t {
		case model.AuthenticatorTypePassword:
			var err error
			passwordSpec, err = c.createPasswordAuthenticatorSpec(userID, opts)
			if err != nil {
				return nil, err
			}
			if passwordSpec != nil {
				authenticatorSpecs = append(authenticatorSpecs, passwordSpec)
			}
		case model.AuthenticatorTypeOOBSMS:
			spec, ok := c.createOOBSMSAuthenticatorSpec(identityInfo, userID)
			if ok {
				authenticatorSpecs = append(authenticatorSpecs, spec)
			}
		case model.AuthenticatorTypeOOBEmail:
			spec, ok := c.createOOBEmailAuthenticatorSpec(identityInfo, userID)
			if ok {
				authenticatorSpecs = append(authenticatorSpecs, spec)
			}
		}
	}

	authenticatorInfos, err := c.createAuthenticatorInfos(ctx, authenticatorSpecs)
	if err != nil {
		return nil, err
	}

	if len(authenticatorInfos) == 0 {
		for _, t := range authenticatorTypes {
			if t == model.AuthenticatorTypePassword {
				return nil, api.NewInvariantViolated(
					"PasswordRequired",
					"password is required",
					nil,
				)
			}
		}
		return nil, api.InvalidConfiguration.New("no primary authenticator can be created for identity")
	}

	if passwordSpec != nil && opts.SendPassword {
		err := c.SendPassword.Send(ctx, userID, passwordSpec.Password.PlainPassword, translation.MessageTypeSendPasswordToNewUser)
		if err != nil {
			return nil, err
		}
	}

	return authenticatorInfos, nil
}

func (c *Coordinator) createPasswordAuthenticatorSpec(userID string, opts CreatePasswordOptions) (*authenticator.Spec, error) {
	if opts.Password == nil {
		return nil, nil
	}

	var expireAfter *time.Time
	if opts.SetPasswordExpired {
		expireAt := c.Clock.NowUTC()
		expireAfter = &expireAt
	}

	password := *opts.Password
	var err error
	if password == "" {
		password, err = c.PasswordGenerator.Generate()
		if err != nil {
			return nil, err
		}
	}

	return &authenticator.Spec{
		UserID:    userID,
		IsDefault: true,
		Kind:      model.AuthenticatorKindPrimary,
		Type:      model.AuthenticatorTypePassword,
		Password: &authenticator.PasswordSpec{
			PlainPassword: password,
			ExpireAfter:   expireAfter,
		},
	}, nil
}

func (c *Coordinator) createOOBSMSAuthenticatorSpec(identityInfo *identity.Info, userID string) (*authenticator.Spec, bool) {
	if identityInfo.LoginID.LoginIDType != model.LoginIDKeyTypePhone {
		return nil, false
	}

	return &authenticator.Spec{
		UserID:    userID,
		IsDefault: true,
		Kind:      model.AuthenticatorKindPrimary,
		Type:      model.AuthenticatorTypeOOBSMS,
		OOBOTP: &authenticator.OOBOTPSpec{
			Phone: identityInfo.LoginID.LoginID,
		},
	}, true
}

func (c *Coordinator) createOOBEmailAuthenticatorSpec(identityInfo *identity.Info, userID string) (*authenticator.Spec, bool) {
	if identityInfo.LoginID.LoginIDType != model.LoginIDKeyTypeEmail {
		return nil, false
	}

	return &authenticator.Spec{
		UserID:    userID,
		IsDefault: true,
		Kind:      model.AuthenticatorKindPrimary,
		Type:      model.AuthenticatorTypeOOBEmail,
		OOBOTP: &authenticator.OOBOTPSpec{
			Email: identityInfo.LoginID.LoginID,
		},
	}, true
}

func (c *Coordinator) createAuthenticatorInfos(ctx context.Context, authenticatorSpecs []*authenticator.Spec) ([]*authenticator.Info, error) {
	var authenticatorInfos []*authenticator.Info
	for _, authenticatorSpec := range authenticatorSpecs {
		authenticatorID := uuid.New()
		authenticatorInfo, err := c.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, authenticatorSpec)
		if err != nil {
			return nil, err
		}

		if err := c.Authenticators.Create(ctx, authenticatorInfo); err != nil {
			return nil, err
		}

		authenticatorInfos = append(authenticatorInfos, authenticatorInfo)
	}
	return authenticatorInfos, nil
}

func (c *Coordinator) UserDelete(ctx context.Context, userID string, isScheduledDeletion bool) error {
	// Delete dependents of user entity.

	// Identities:
	identities, err := c.Identities.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, i := range identities {
		if err = c.Identities.Delete(ctx, i); err != nil {
			return err
		}
	}

	// Authenticators:
	authenticators, err := c.Authenticators.List(ctx, userID)
	if err != nil {
		return err
	}
	for _, a := range authenticators {
		if err = c.Authenticators.Delete(ctx, a); err != nil {
			return err
		}
	}

	// MFA recovery codes:
	if err = c.MFA.InvalidateAllRecoveryCode(ctx, userID); err != nil {
		return err
	}

	// OAuth authorizations:
	if err = c.OAuth.ResetAll(ctx, userID); err != nil {
		return err
	}

	// Verified claims:
	if err = c.Verification.ResetVerificationStatus(ctx, userID); err != nil {
		return err
	}

	// Password history:
	if err = c.PasswordHistory.ResetPasswordHistory(ctx, userID); err != nil {
		return err
	}

	// Sessions:
	if err = c.terminateAllSessionsAndCleanUp(ctx, userID); err != nil {
		return err
	}

	// Groups:
	if err = c.RolesGroupsCommands.DeleteUserGroup(ctx, userID); err != nil {
		return err
	}

	// Roles:
	if err = c.RolesGroupsCommands.DeleteUserRole(ctx, userID); err != nil {
		return err
	}

	userModel, err := c.UserQueries.Get(ctx, userID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}

	// Finally, delete the user.
	err = c.UserCommands.Delete(ctx, userID)
	if err != nil {
		return err
	}

	err = c.Events.DispatchEventOnCommit(ctx, &nonblocking.UserDeletedEventPayload{
		UserModel:           *userModel,
		IsScheduledDeletion: isScheduledDeletion,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) removeOrphans(ctx context.Context, userID string) error {
	identities, err := c.Identities.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	err = c.Authenticators.RemoveOrphans(ctx, identities)
	if err != nil {
		return err
	}

	authenticators, err := c.Authenticators.List(ctx, userID)
	if err != nil {
		return err
	}

	err = c.Verification.RemoveOrphanedClaims(ctx, userID, identities, authenticators)
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
		err = c.MFA.InvalidateAllRecoveryCode(ctx, userID)
		if err != nil {
			return err
		}
		err = c.MFA.InvalidateAllDeviceTokens(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) markVerified(ctx context.Context, userID string, claims map[model.ClaimName]string) error {
	for name, value := range claims {
		name := string(name)

		claim := c.Verification.NewVerifiedClaim(ctx, userID, name, value)
		err := c.Verification.MarkClaimVerified(ctx, claim)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) MarkOOBIdentityVerified(ctx context.Context, info *authenticator.Info) error {
	claim := map[model.ClaimName]string{}
	switch info.Type {
	case model.AuthenticatorTypeOOBEmail:
		identities, err := c.Identities.ListByClaim(ctx, string(model.ClaimEmail), info.OOBOTP.Email)
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			return nil
		}
		claim[model.ClaimEmail] = info.OOBOTP.Email
	case model.AuthenticatorTypeOOBSMS:
		identities, err := c.Identities.ListByClaim(ctx, string(model.ClaimPhoneNumber), info.OOBOTP.Phone)
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

	err := c.markVerified(ctx, info.UserID, claim)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) markOAuthEmailAsVerified(ctx context.Context, info *identity.Info) error {
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
			err := c.markVerified(ctx, info.UserID, map[model.ClaimName]string{
				model.ClaimEmail: email,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Coordinator) terminateAllSessionsAndCleanUp(ctx context.Context, userID string) error {
	err := c.terminateAllSessions(ctx, userID)
	if err != nil {
		return err
	}

	// According to https://redis.io/docs/latest/commands/hdel/
	// When the last field of a Redis hash is deleted, the hash is deleted.
	// So in most case, we do not actually need this.
	// However, List() does not actually return every key that the Redis hash has,
	// so having CleanUpForDeletingUserID ensures we delete the hash.

	err = c.IDPSessions.CleanUpForDeletingUserID(ctx, userID)
	if err != nil {
		return err
	}

	err = c.OAuthSessions.CleanUpForDeletingUserID(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) terminateAllSessions(ctx context.Context, userID string) error {
	idpSessions, err := c.IDPSessions.List(ctx, userID)
	if err != nil {
		return err
	}
	for _, s := range idpSessions {
		if err = c.IDPSessions.Delete(ctx, s); err != nil {
			return err
		}
	}

	oauthSessions, err := c.OAuthSessions.List(ctx, userID)
	if err != nil {
		return err
	}
	for _, s := range oauthSessions {
		if err = c.OAuthSessions.Delete(ctx, s); err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserReenable(ctx context.Context, userID string) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	accountStatus, err := u.AccountStatus(now).Reenable()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
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

	err = c.Events.DispatchEventOnCommit(ctx, e)
	if err != nil {
		return err
	}

	return nil
}

type SetDisabledOptions struct {
	UserID                   string
	IsDisabled               bool
	Reason                   *string
	TemporarilyDisabledFrom  *time.Time
	TemporarilyDisabledUntil *time.Time
}

func (c *Coordinator) UserDisable(ctx context.Context, options SetDisabledOptions) error {
	u, err := c.UserQueries.GetRaw(ctx, options.UserID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()

	var accountStatus *user.AccountStatusWithRefTime
	if options.TemporarilyDisabledFrom != nil || options.TemporarilyDisabledUntil != nil {
		accountStatus, err = u.AccountStatus(now).DisableTemporarily(
			options.TemporarilyDisabledFrom,
			options.TemporarilyDisabledUntil,
			options.Reason,
		)
	} else {
		accountStatus, err = u.AccountStatus(now).DisableIndefinitely(options.Reason)
	}
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, options.UserID, *accountStatus)
	if err != nil {
		return err
	}

	// If the account is disabled now, terminate all sessions.
	if accountStatus.IsDisabled() {
		err = c.terminateAllSessions(ctx, options.UserID)
		if err != nil {
			return err
		}
	}

	// Always fire this event, even the account is not considered as disabled now.
	e := &nonblocking.UserDisabledEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: options.UserID,
			},
		},
		TemporarilyDisabledFrom:  options.TemporarilyDisabledFrom,
		TemporarilyDisabledUntil: options.TemporarilyDisabledUntil,
	}

	err = c.Events.DispatchEventOnCommit(ctx, e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserSetAccountValidFrom(ctx context.Context, userID string, from *time.Time) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()

	accountStatus, err := u.AccountStatus(now).SetAccountValidFrom(from)
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
	if err != nil {
		return err
	}

	// If the account is disabled now, terminate all sessions.
	if accountStatus.IsDisabled() {
		err = c.terminateAllSessions(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserSetAccountValidUntil(ctx context.Context, userID string, until *time.Time) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()

	accountStatus, err := u.AccountStatus(now).SetAccountValidUntil(until)
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
	if err != nil {
		return err
	}

	// If the account is disabled now, terminate all sessions.
	if accountStatus.IsDisabled() {
		err = c.terminateAllSessions(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserSetAccountValidPeriod(ctx context.Context, userID string, from *time.Time, until *time.Time) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()

	accountStatus, err := u.AccountStatus(now).SetAccountValidPeriod(from, until)
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
	if err != nil {
		return err
	}

	// If the account is disabled now, terminate all sessions.
	if accountStatus.IsDisabled() {
		err = c.terminateAllSessions(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserScheduleDeletionByAdmin(ctx context.Context, userID string) error {
	return c.userScheduleDeletion(ctx, userID, true)
}

func (c *Coordinator) UserScheduleDeletionByEndUser(ctx context.Context, userID string) error {
	return c.userScheduleDeletion(ctx, userID, false)
}

func (c *Coordinator) userScheduleDeletion(ctx context.Context, userID string, byAdmin bool) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	deleteAt := now.Add(c.AccountDeletionConfig.GracePeriod.Duration())

	var accountStatus *user.AccountStatusWithRefTime
	if byAdmin {
		accountStatus, err = u.AccountStatus(now).ScheduleDeletionByAdmin(deleteAt)
	} else {
		accountStatus, err = u.AccountStatus(now).ScheduleDeletionByEndUser(deleteAt)
	}
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
	if err != nil {
		return err
	}

	err = c.terminateAllSessions(ctx, userID)
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
		err := c.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserUnscheduleDeletionByAdmin(ctx context.Context, userID string) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	accountStatus, err := u.AccountStatus(now).UnscheduleDeletionByAdmin()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
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

	err = c.Events.DispatchEventOnCommit(ctx, e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserAnonymize(ctx context.Context, userID string, IsScheduledAnonymization bool) error {
	// Delete dependents of user entity.

	// Identities:
	identities, err := c.Identities.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, i := range identities {
		if err = c.Identities.Delete(ctx, i); err != nil {
			return err
		}
	}

	// Authenticators:
	authenticators, err := c.Authenticators.List(ctx, userID)
	if err != nil {
		return err
	}
	for _, a := range authenticators {
		if err = c.Authenticators.Delete(ctx, a); err != nil {
			return err
		}
	}

	// MFA recovery codes:
	if err = c.MFA.InvalidateAllRecoveryCode(ctx, userID); err != nil {
		return err
	}

	// OAuth authorizations:
	if err = c.OAuth.ResetAll(ctx, userID); err != nil {
		return err
	}

	// Verified claims:
	if err = c.Verification.ResetVerificationStatus(ctx, userID); err != nil {
		return err
	}

	// Password history:
	if err = c.PasswordHistory.ResetPasswordHistory(ctx, userID); err != nil {
		return err
	}

	// Sessions:
	if err = c.terminateAllSessionsAndCleanUp(ctx, userID); err != nil {
		return err
	}

	// Groups:
	if err = c.RolesGroupsCommands.DeleteUserGroup(ctx, userID); err != nil {
		return err
	}

	// Roles:
	if err = c.RolesGroupsCommands.DeleteUserRole(ctx, userID); err != nil {
		return err
	}

	userModel, err := c.UserQueries.Get(ctx, userID, accesscontrol.RoleGreatest)
	if err != nil {
		return err
	}

	// Anonymize user record:
	err = c.UserCommands.Anonymize(ctx, userID)
	if err != nil {
		return err
	}

	err = c.Events.DispatchEventOnCommit(ctx, &nonblocking.UserAnonymizedEventPayload{
		UserModel:                *userModel,
		IsScheduledAnonymization: IsScheduledAnonymization,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserScheduleAnonymizationByAdmin(ctx context.Context, userID string) error {
	return c.userScheduleAnonymization(ctx, userID, true)
}

func (c *Coordinator) userScheduleAnonymization(ctx context.Context, userID string, byAdmin bool) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	anonymizeAt := now.Add(c.AccountAnonymizationConfig.GracePeriod.Duration())

	var accountStatus *user.AccountStatusWithRefTime
	if byAdmin {
		accountStatus, err = u.AccountStatus(now).ScheduleAnonymizationByAdmin(anonymizeAt)
	} else {
		err = errors.New("not implemented")
	}
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
	if err != nil {
		return err
	}

	err = c.terminateAllSessions(ctx, userID)
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
		err := c.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) UserUnscheduleAnonymizationByAdmin(ctx context.Context, userID string) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	accountStatus, err := u.AccountStatus(now).UnscheduleAnonymizationByAdmin()
	if err != nil {
		return err
	}

	err = c.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
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

	err = c.Events.DispatchEventOnCommit(ctx, e)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) UserCheckAnonymized(ctx context.Context, userID string) error {
	u, err := c.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return err
	}

	now := c.Clock.NowUTC()
	if u.AccountStatus(now).IsAnonymized() {
		return ErrUserIsAnonymized
	}

	return nil
}

func (c *Coordinator) UserUpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error {
	if endAt != nil {
		// Cannot set grace period in the past or more than 90 days in the future.
		if endAt.Before(c.Clock.NowUTC()) {
			return ErrMFAGracePeriodInvalid
		}

		maxEndAt := c.Clock.NowUTC().AddDate(0, 0, 90)
		if endAt.After(maxEndAt) {
			return ErrMFAGracePeriodInvalid
		}
	}

	err := c.UserCommands.UpdateMFAEnrollment(ctx, userID, endAt)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) MarkClaimVerifiedByAdmin(ctx context.Context, claim *verification.Claim) error {
	is, err := c.IdentityListByClaim(ctx, claim.Name, claim.Value)
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
		return api.ErrIdentityNotFound
	}

	if err := c.Verification.MarkClaimVerified(ctx, claim); err != nil {
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
		err = c.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) DeleteVerifiedClaimByAdmin(ctx context.Context, claim *verification.Claim) error {
	is, err := c.IdentityListByClaim(ctx, claim.Name, claim.Value)
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
		return api.ErrIdentityNotFound
	}

	if err := c.Verification.DeleteClaim(ctx, claim); err != nil {
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
		err = c.Events.DispatchEventOnCommit(ctx, e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Coordinator) AuthenticatorClearLockoutAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	return c.Authenticators.ClearLockoutAttempts(ctx, userID, usedMethods)
}

func (c *Coordinator) MFAGenerateDeviceToken(ctx context.Context) string {
	return c.MFA.GenerateDeviceToken(ctx)
}

func (c *Coordinator) MFACreateDeviceToken(ctx context.Context, userID string, token string) (*mfa.DeviceToken, error) {
	return c.MFA.CreateDeviceToken(ctx, userID, token)
}

func (c *Coordinator) MFAVerifyDeviceToken(ctx context.Context, userID string, token string) error {
	return c.MFA.VerifyDeviceToken(ctx, userID, token)
}

func (c *Coordinator) MFAInvalidateAllDeviceTokens(ctx context.Context, userID string) error {
	return c.MFA.InvalidateAllDeviceTokens(ctx, userID)
}

func (c *Coordinator) MFAVerifyRecoveryCode(ctx context.Context, userID string, code string) (*mfa.RecoveryCode, error) {
	rc, err := c.MFA.VerifyRecoveryCode(ctx, userID, code)
	authnType := authn.AuthenticationTypeRecoveryCode
	if errors.Is(err, mfa.ErrRecoveryCodeNotFound) || errors.Is(err, mfa.ErrRecoveryCodeConsumed) {
		err = c.dispatchAuthenticationFailedEvent(
			ctx,
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

func (c *Coordinator) MFAConsumeRecoveryCode(ctx context.Context, rc *mfa.RecoveryCode) error {
	return c.MFA.ConsumeRecoveryCode(ctx, rc)
}

func (c *Coordinator) MFAGenerateRecoveryCodes(ctx context.Context) []string {
	return c.MFA.GenerateRecoveryCodes(ctx)
}

func (c *Coordinator) MFAReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error) {
	return c.MFA.ReplaceRecoveryCodes(ctx, userID, codes)
}

func (c *Coordinator) MFAListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error) {
	return c.MFA.ListRecoveryCodes(ctx, userID)
}

func (c *Coordinator) GetUsersByStandardAttribute(ctx context.Context, attributeName string, attributeValue string) ([]string, error) {
	var loginIDKeyType model.LoginIDKeyType
	switch attributeName {
	case stdattrs.Email:
		loginIDKeyType = model.LoginIDKeyTypeEmail
	case stdattrs.PhoneNumber:
		loginIDKeyType = model.LoginIDKeyTypePhone
	case stdattrs.PreferredUsername:
		loginIDKeyType = model.LoginIDKeyTypeUsername
	default:
		return nil, api.ErrGetUsersInvalidArgument.New("attributeName must be email, phone_number or preferred_username")
	}

	normalized, _, err := c.Identities.Normalize(ctx, loginIDKeyType, attributeValue)
	if err != nil {
		return nil, api.ErrGetUsersInvalidArgument.New("invalid attributeValue")
	}

	identities, err := c.Identities.ListByClaim(ctx, attributeName, normalized)

	if err != nil {
		return nil, err
	}

	var userIDsKeyMap = make(map[string]struct{})
	var uniqueUserIDs = make([]string, 0)

	for _, iden := range identities {
		userIDsKeyMap[iden.UserID] = struct{}{}
	}

	for k := range userIDsKeyMap {
		uniqueUserIDs = append(uniqueUserIDs, k)
	}

	return uniqueUserIDs, nil
}

func (c *Coordinator) GetUserByLoginID(ctx context.Context, loginIDKey string, loginIDValue string) (string, error) {
	info, err := c.Identities.AdminAPIGetByLoginIDKeyAndLoginIDValue(ctx, loginIDKey, loginIDValue)

	if errors.Is(err, api.ErrIdentityNotFound) {
		return "", api.ErrUserNotFound
	} else if err != nil {
		return "", err
	}

	return info.UserID, nil
}

func (c *Coordinator) GetUserByOAuth(ctx context.Context, oauthProviderAlias string, oauthProviderUserID string) (string, error) {
	info, err := c.Identities.AdminAPIGetByOAuthAliasAndSubject(ctx, oauthProviderAlias, oauthProviderUserID)

	if errors.Is(err, api.ErrIdentityNotFound) {
		return "", api.ErrUserNotFound
	} else if err != nil {
		return "", err
	}

	return info.UserID, nil
}

func (c *Coordinator) GetUserIDsByLoginHint(ctx context.Context, hint *oauth.LoginHint) ([]string, error) {
	if hint.Type != oauth.LoginHintTypeLoginID {
		panic("Only login_hint with type login_id is supported")
	}

	userIDs := setutil.Set[string]{}

	findUserIDsByClaim := func(claimName string, value string) error {
		ids, err := c.GetUsersByStandardAttribute(ctx, claimName, value)
		if err != nil {
			return err
		}
		for _, id := range ids {
			userIDs.Add(id)
		}
		return nil
	}

	switch {
	case hint.LoginIDEmail != "":
		err := findUserIDsByClaim(stdattrs.Email, hint.LoginIDEmail)
		if err != nil {
			return nil, err
		}
	case hint.LoginIDPhone != "":
		err := findUserIDsByClaim(stdattrs.PhoneNumber, hint.LoginIDPhone)
		if err != nil {
			return nil, err
		}
	case hint.LoginIDUsername != "":
		err := findUserIDsByClaim(stdattrs.PreferredUsername, hint.LoginIDUsername)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unable to find user by login_hint as no login_id provided")
	}
	return userIDs.Keys(), nil
}
