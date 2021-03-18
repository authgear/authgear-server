package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type IdentityService interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	GetBySpec(spec *identity.Spec) (*identity.Info, error)
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
	Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
	WithSecret(authenticatorInfo *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error)
	Create(authenticatorInfo *authenticator.Info) error
	Update(authenticatorInfo *authenticator.Info) error
	Delete(authenticatorInfo *authenticator.Info) error
	VerifySecret(info *authenticator.Info, state map[string]string, secret string) error
	RemoveOrphans(identities []*identity.Info) error
}

type VerificationService interface {
	GetClaimVerificationStatus(userID string, name string, value string) (verification.Status, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	RemoveOrphanedClaims(identities []*identity.Info, authenticators []*authenticator.Info) error
	ResetVerificationStatus(userID string) error
}

type MFAService interface {
	InvalidateAllRecoveryCode(userID string) error
}

type UserCommands interface {
	Delete(userID string) error
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

type IDPSessionManager SessionManager
type OAuthSessionManager SessionManager

// Coordinator represents interaction between identities, authenticators, and
// other high-level features (such as verification).
// FIXME(interaction): This is used to avoid circular dependency between
//                     feature implementations. We should investigate a proper
//                     resolution, as the interactions between features will
//                     get complicated fast.
// FIXME(mfa): remove all MFA recovery code when last secondary authenticator is
//             removed, so that recovery codes are re-generated when setup again.
type Coordinator struct {
	Identities      IdentityService
	Authenticators  AuthenticatorService
	Verification    VerificationService
	MFA             MFAService
	Users           UserCommands
	PasswordHistory PasswordHistoryStore
	OAuth           OAuthService
	IDPSessions     IDPSessionManager
	OAuthSessions   OAuthSessionManager
	IdentityConfig  *config.IdentityConfig
}

func (c *Coordinator) IdentityGet(userID string, typ authn.IdentityType, id string) (*identity.Info, error) {
	return c.Identities.Get(userID, typ, id)
}

func (c *Coordinator) IdentityGetBySpec(spec *identity.Spec) (*identity.Info, error) {
	return c.Identities.GetBySpec(spec)
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

	if is.Type == authn.IdentityTypeOAuth {
		providerID := config.NewProviderID(
			is.Claims[identity.IdentityClaimOAuthProviderKeys].(map[string]interface{}),
		)
		var cfg *config.OAuthSSOProviderConfig
		for _, c := range c.IdentityConfig.OAuth.Providers {
			if c.ProviderID().Equal(&providerID) {
				c := c
				cfg = &c
				break
			}
		}

		email, ok := is.Claims[identity.StandardClaimEmail].(string)
		if ok && cfg != nil && *cfg.Claims.Email.AssumeVerified {
			// Mark as verified if OAuth email is assumed to be verified
			err = c.markVerified(is.UserID, map[authn.ClaimName]string{
				authn.ClaimEmail: email,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Coordinator) IdentityUpdate(info *identity.Info) error {
	err := c.Identities.Update(info)
	if err != nil {
		return err
	}

	err = c.removeOrphans(info.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityDelete(is *identity.Info) error {
	err := c.Identities.Delete(is)
	if err != nil {
		return err
	}

	err = c.removeOrphans(is.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Coordinator) IdentityCheckDuplicated(info *identity.Info) (*identity.Info, error) {
	return c.Identities.CheckDuplicated(info)
}

func (c *Coordinator) AuthenticatorGet(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error) {
	return c.Authenticators.Get(userID, typ, id)
}

func (c *Coordinator) AuthenticatorList(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return c.Authenticators.List(userID, filters...)
}

func (c *Coordinator) AuthenticatorNew(spec *authenticator.Spec, secret string) (*authenticator.Info, error) {
	return c.Authenticators.New(spec, secret)
}

func (c *Coordinator) AuthenticatorWithSecret(authenticatorInfo *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error) {
	return c.Authenticators.WithSecret(authenticatorInfo, secret)
}

func (c *Coordinator) AuthenticatorCreate(authenticatorInfo *authenticator.Info) error {
	err := c.Authenticators.Create(authenticatorInfo)
	if err != nil {
		return err
	}

	// Mark as verified for authenticators.
	err = c.markVerified(authenticatorInfo.UserID, authenticatorInfo.StandardClaims())
	if err != nil {
		return err
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

func (c *Coordinator) AuthenticatorVerifySecret(info *authenticator.Info, state map[string]string, secret string) error {
	return c.Authenticators.VerifySecret(info, state, secret)
}

func (c *Coordinator) UserDelete(userID string) error {
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

	// Finally, delete the user.
	return c.Users.Delete(userID)
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

func (c *Coordinator) markVerified(userID string, claims map[authn.ClaimName]string) error {
	for name, value := range claims {
		name := string(name)
		status, err := c.Verification.GetClaimVerificationStatus(userID, name, value)
		if err != nil {
			return err
		}
		if status != verification.StatusPending && status != verification.StatusRequired {
			continue
		}

		claim := c.Verification.NewVerifiedClaim(userID, name, value)
		err = c.Verification.MarkClaimVerified(claim)
		if err != nil {
			return err
		}
	}
	return nil
}
