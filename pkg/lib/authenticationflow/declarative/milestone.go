package declarative

import (
	"context"
	"sort"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/slice"
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

func collectAMR(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (amr []string, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				amr = append(amr, n.MilestoneDidAuthenticate()...)
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				amr = append(amr, i.MilestoneDidAuthenticate()...)
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	amr = slice.Deduplicate(amr)
	sort.Strings(amr)

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
	MilestoneIdentificationMethod() config.AuthenticationFlowIdentification
}

type MilestoneAuthenticationMethod interface {
	authflow.Milestone
	MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication
}

type MilestoneFlowAuthenticate interface {
	authflow.Milestone
	MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool)
}

type MilestoneDidAuthenticate interface {
	authflow.Milestone
	MilestoneDidAuthenticate() (amr []string)
}

type MilestoneDoCreateSession interface {
	authflow.Milestone
	MilestoneDoCreateSession() (*idpsession.IDPSession, bool)
}

type MilestoneDoCreateUser interface {
	authflow.Milestone
	MilestoneDoCreateUser() (string, bool)
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
	MilestoneDoCreateIdentitySkipCreate()
	MilestoneDoCreateIdentityUpdate(newInfo *identity.Info)
}

type MilestoneFlowCreateAuthenticator interface {
	authflow.Milestone
	MilestoneFlowCreateAuthenticator(flows authflow.Flows) (created MilestoneDoCreateAuthenticator, newFlow authflow.Flows, ok bool)
}

type MilestoneDoCreateAuthenticator interface {
	authflow.Milestone
	MilestoneDoCreateAuthenticator() *authenticator.Info
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
	MilestoneAccountRecoveryCode() string
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
	MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error
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

// For IntentBotProtection to check if itself has finished
type MilestoneDidPerformBotProtectionVerification interface {
	authflow.Milestone
	MilestoneDidPerformBotProtectionVerification()
}

// To prevent spawning multiple IntentBotProtection
type MilestoneFlowBotProtection interface {
	authflow.Milestone
	MilestoneFlowBotProtection()
}
