package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func getUserID(flows authflow.Flows) (userID string, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDoUseUser); ok {
				id := n.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneDoUseUser); ok {
				id := i.MilestoneDoUseUser()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
	}, flows.Root)

	if userID == "" {
		err = ErrNoUserID
	}

	if err != nil {
		return
	}

	return
}

func collectIdentitySpecs(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (specs []*identity.Spec, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneGetIdentitySpecs); ok {
				specs = append(specs, n.MilestoneGetIdentitySpecs()...)
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneGetIdentitySpecs); ok {
				specs = append(specs, i.MilestoneGetIdentitySpecs()...)
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	return
}

func collectAuthenticationLockoutMethod(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (methods []config.AuthenticationLockoutMethod, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidUseAuthenticationLockoutMethod); ok {
				if m, ok := n.MilestoneDidUseAuthenticationLockoutMethod(); ok {
					methods = append(methods, m)
				}

			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneDidUseAuthenticationLockoutMethod); ok {
				if m, ok := i.MilestoneDidUseAuthenticationLockoutMethod(); ok {
					methods = append(methods, m)
				}
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	return
}

type MilestoneNestedSteps interface {
	authflow.Milestone
	MilestoneNestedSteps()
}

type MilestoneIdentificationMethod interface {
	authflow.Milestone
	MilestoneIdentificationMethod() model.AuthenticationFlowIdentification
}

type MilestoneFlowSelectAuthenticationMethod interface {
	authflow.Milestone
	MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (selected MilestoneDidSelectAuthenticationMethod, newFlows authflow.Flows, ok bool)
}

type MilestoneDidSelectAuthenticationMethod interface {
	authflow.Milestone
	MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication
}

type MilestoneFlowAuthenticate interface {
	authflow.Milestone
	MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool)
}

type MilestoneDidAuthenticate interface {
	authflow.Milestone
	MilestoneDidAuthenticate() (amr []string)
	MilestoneDidAuthenticateAuthenticator() (*authenticator.Info, bool)
	MilestoneDidAuthenticateAuthentication() (*model.Authentication, bool)
}

type MilestoneDoCreateSession interface {
	authflow.Milestone
	MilestoneDoCreateSession() (*idpsession.IDPSession, bool)
}

type MilestoneDoCreateUser interface {
	authflow.Milestone
	MilestoneDoCreateUser() (userID string, createUser bool)
	MilestoneDoCreateUserUseExisting(userID string)
}

type MilestoneFlowCreateIdentity interface {
	authflow.Milestone
	MilestoneFlowCreateIdentity(flows authflow.Flows) (created MilestoneDoCreateIdentity, newFlows authflow.Flows, ok bool)
}

type MilestoneFlowAccountLinking interface {
	authflow.Milestone
	MilestoneFlowCreateIdentity
	MilestoneFlowAccountLinking()
}

type MilestoneDoCreateIdentity interface {
	authflow.Milestone
	MilestoneDoCreateIdentity() *identity.Info
	MilestoneDoCreateIdentityIdentification() model.Identification
	MilestoneDoCreateIdentitySkipCreate()
	MilestoneDoCreateIdentityUpdate(newInfo *identity.Info)
}

type MilestoneFlowCreateAuthenticator interface {
	authflow.Milestone
	MilestoneFlowCreateAuthenticator(flows authflow.Flows) (created MilestoneDoCreateAuthenticator, newFlow authflow.Flows, ok bool)
}

type MilestoneDoCreateAuthenticator interface {
	authflow.Milestone
	MilestoneDoCreateAuthenticator() (*authenticator.Info, bool)
	MilestoneDoCreateAuthenticatorAuthentication() (*model.Authentication, bool)
	MilestoneDoCreateAuthenticatorSkipCreate()
	MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info)
}

type MilestoneDoCreatePasskey interface {
	authflow.Milestone
	MilestoneDoCreatePasskeyUpdateUserID(userID string)
}

type MilestoneDoUseUser interface {
	authflow.Milestone
	MilestoneDoUseUser() string
}

type MilestoneFlowUseIdentity interface {
	authflow.Milestone
	MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool)
}

type MilestoneDoUseIdentity interface {
	authflow.Milestone
	MilestoneDoUseIdentity() *identity.Info
	MilestoneDoUseIdentityIdentification() model.Identification
}

type MilestoneDoUseAccountRecoveryIdentity interface {
	authflow.Milestone
	MilestoneDoUseAccountRecoveryIdentity() AccountRecoveryIdentity
}

type MilestoneDoUseAccountRecoveryIdentificationMethod interface {
	authflow.Milestone
	MilestoneDoUseAccountRecoveryIdentificationMethod() config.AuthenticationFlowAccountRecoveryIdentification
}

type MilestoneDoUseAccountRecoveryDestination interface {
	authflow.Milestone
	MilestoneDoUseAccountRecoveryDestination() *AccountRecoveryDestinationOptionInternal
}

type MilestoneAccountRecoveryCode interface {
	authflow.Milestone
	MilestoneAccountRecoveryCode() struct {
		MaskedTarget string
		Code         string
	}
}

type MilestoneDidSelectAuthenticator interface {
	authflow.Milestone
	MilestoneDidSelectAuthenticator() *authenticator.Info
}

type MilestoneDoUseAuthenticatorPassword interface {
	authflow.Milestone
	MilestoneDoUseAuthenticatorPassword() *NodeDoUseAuthenticatorPassword
	GetJSONPointer() jsonpointer.T
}

type MilestoneDoPopulateStandardAttributes interface {
	authflow.Milestone
	MilestoneDoPopulateStandardAttributes()
	MilestoneDoPopulateStandardAttributesSkip()
}

type MilestoneVerifyClaim interface {
	authflow.Milestone
	MilestoneVerifyClaim()
	MilestoneVerifyClaimUpdateUserID(deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error
}

type MilestoneDoMarkClaimVerified interface {
	authflow.Milestone
	MilestoneDoMarkClaimVerified()
	MilestoneDoMarkClaimVerifiedUpdateUserID(newUserID string)
}

type MilestoneDeviceTokenInspected interface {
	authflow.Milestone
	MilestoneDeviceTokenInspected()
}

type MilestoneDoCreateDeviceTokenIfRequested interface {
	authflow.Milestone
	MilestoneDoCreateDeviceTokenIfRequested()
}

type MilestoneDidUseAuthenticationLockoutMethod interface {
	authflow.Milestone
	MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool)
}

type MilestoneDoUseAnonymousUser interface {
	authflow.Milestone
	MilestoneDoUseAnonymousUser() *identity.Info
}

type MilestoneDidReauthenticate interface {
	authflow.Milestone
	MilestoneDidReauthenticate()
}

type MilestoneSwitchToExistingUser interface {
	authflow.Milestone
	MilestoneSwitchToExistingUser(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error
}

type MilestoneDoReplaceRecoveryCode interface {
	authflow.Milestone
	MilestoneDoReplaceRecoveryCodeUpdateUserID(newUserID string)
}

type MilestoneDoUpdateUserProfile interface {
	authflow.Milestone
	MilestoneDoUpdateUserProfileSkip()
}

type MilestoneUseAccountLinkingIdentification interface {
	authflow.Milestone
	MilestoneUseAccountLinkingIdentification() *AccountLinkingConflict
	MilestoneUseAccountLinkingIdentificationSelectedOption() AccountLinkingIdentificationOption
	MilestoneUseAccountLinkingIdentificationRedirectURI() string
	MilestoneUseAccountLinkingIdentificationResponseMode() string
}

type MilestonePromptCreatePasskey interface {
	authflow.Milestone
	MilestonePromptCreatePasskey()
}

type MilestoneCheckLoginHint interface {
	authflow.Milestone
	MilestoneCheckLoginHint()
}

type MilestoneGetIdentitySpecs interface {
	authflow.Milestone
	MilestoneGetIdentitySpecs() []*identity.Spec
}

type MilestoneConstraintsProvider interface {
	authflow.Milestone
	MilestoneConstraintsProvider() *event.Constraints
}

type MilestoneBotProjectionRequirementsProvider interface {
	authflow.Milestone
	MilestoneBotProjectionRequirementsProvider() *event.BotProtectionRequirements
}

type MilestoneDidConsumeRecoveryCode interface {
	authflow.Milestone
	MilestoneDidConsumeRecoveryCode() *mfa.RecoveryCode
}

type MilestoneAuthenticationFlowObjectProvider interface {
	authflow.Milestone
	MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject
}

type MilestonePreAuthenticated interface {
	authflow.Milestone
	MilestonePreAuthenticated()
}

type MilestoneOOBOTPVerified interface {
	authflow.Milestone
	MilestoneOOBOTPVerifiedChannel() model.AuthenticatorOOBChannel
}

type MilestoneOOBOTPLastUsedChannelUpdated interface {
	authflow.Milestone
	MilestoneOOBOTPLastUsedChannelUpdated()
}
