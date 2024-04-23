package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

func init() {
	authflow.RegisterIntent(&IntentAccountLinking{})
}

type IntentAccountLinking struct {
	JSONPointer           jsonpointer.T    `json:"json_pointer,omitempty"`
	SkipLogin             bool             `json:"skip_login,omitempty"`
	LoginFlowName         string           `json:"login_flow_name,omitempty"`
	OAuthIdentitySpec     *identity.Spec   `json:"oauth_identity_spec,omitempty"`
	ConflictingIdentities []*identity.Info `json:"conflicting_identities,omitempty"`
}

var _ authflow.Intent = &IntentAccountLinking{}
var _ authflow.DataOutputer = &IntentAccountLinking{}

func (*IntentAccountLinking) Kind() string {
	return "IntentAccountLinking"
}

func (i *IntentAccountLinking) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewAccountLinkingIdentifyData(i.getOptions()), nil
}

func (i *IntentAccountLinking) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // Ask for identity to link
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaAccountLinkingIdentification{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			Options:        i.getOptions(),
		}, nil
	case 1: // Enter the login flow
		return nil, nil
	case 2: // Add NodeDoCreateIdentity
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAccountLinking) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Root.Nodes) == 0 {
		var inputTakeAccountLinkingIdentification inputTakeAccountLinkingIdentification
		if authflow.AsInput(input, &inputTakeAccountLinkingIdentification) {
			idx := inputTakeAccountLinkingIdentification.GetAccountLinkingIdentificationIndex()
			redirectURI := inputTakeAccountLinkingIdentification.GetAccountLinkingOAuthRedirectURI()
			responseMode := inputTakeAccountLinkingIdentification.GetAccountLinkingOAuthResponseMode()
			selectedOption := i.getOptions()[idx]

			return authflow.NewNodeSimple(&NodeUseAccountLinkingIdentification{
				Option:       selectedOption.AccountLinkingIdentificationOption,
				Identity:     selectedOption.Identity,
				RedirectURI:  redirectURI,
				ResponseMode: responseMode,
			}), nil
		}
		return nil, authflow.ErrIncompatibleInput
	}

	milestone, ok := authflow.FindMilestoneInCurrentFlow[MilestoneUseAccountLinkingIdentification](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("expected milestone MilestoneUseAccountLinkingIdentification not found"))
	}

	flowUserID, err := getUserID(flows)
	conflictedIdentity := milestone.MilestoneUseAccountLinkingIdentification()
	conflictedUserID := conflictedIdentity.UserID
	if err != nil {
		return nil, err
	}
	if flowUserID != conflictedUserID {
		err = authflow.TraverseFlow(authflow.Traverser{
			NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
				milestone, ok := nodeSimple.(MilestoneSwitchToExistingUser)
				if ok {
					err = milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
					if err != nil {
						return err
					}
				}
				return nil
			},
			Intent: func(intent authflow.Intent, w *authflow.Flow) error {
				milestone, ok := intent.(MilestoneSwitchToExistingUser)
				if ok {
					err = milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
					if err != nil {
						return err
					}
				}
				return nil
			},
		}, flows.Root)
		if err != nil {
			return nil, err
		}
		// Use synthetic input to auto select the conflicted identity in the login flow
		return nil, &authflow.ErrorRewriteFlow{
			Intent:         flows.Root.Intent,
			Nodes:          flows.Root.Nodes,
			SyntheticInput: i.createSyntheticInputOAuthConflict(milestone, i.OAuthIdentitySpec, conflictedIdentity),
		}
	}

	switch len(flows.Nearest.Nodes) {
	case 1:
		if i.SkipLogin {
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
		if i.LoginFlowName == "" {
			panic(fmt.Errorf("login_flow_name must be specified"))
		}
		flowReference := authflow.FlowReference{
			Type: authflow.FlowTypeLogin,
			Name: i.LoginFlowName,
		}
		loginIntent := IntentLoginFlow{
			TargetUserID:  conflictedUserID,
			FlowReference: flowReference,
		}
		return authflow.NewSubFlow(&loginIntent), nil
	case 2:
		info, err := newIdentityInfo(deps, conflictedIdentity.UserID, i.OAuthIdentitySpec)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountLinking) createSyntheticInputOAuthConflict(
	milestone MilestoneUseAccountLinkingIdentification,
	oauthIden *identity.Spec,
	conflictedInfo *identity.Info) *SyntheticInputAccountLinkingIdentify {
	input := &SyntheticInputAccountLinkingIdentify{}

	switch conflictedInfo.Type {
	case model.IdentityTypeLoginID:
		input.LoginID = conflictedInfo.LoginID.LoginID
		switch conflictedInfo.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			input.Identification = config.AuthenticationFlowIdentificationEmail
		case model.LoginIDKeyTypePhone:
			input.Identification = config.AuthenticationFlowIdentificationPhone
		case model.LoginIDKeyTypeUsername:
			input.Identification = config.AuthenticationFlowIdentificationUsername
		}
	case model.IdentityTypeOAuth:
		input.Identification = config.AuthenticationFlowIdentificationOAuth
		input.Alias = conflictedInfo.OAuth.ProviderAlias
		input.RedirectURI = milestone.MilestoneUseAccountLinkingIdentificationRedirectURI()
		input.ResponseMode = milestone.MilestoneUseAccountLinkingIdentificationResponseMode()
	default:
		// This is a panic because the node should not provide option that we don't know how to handle to the user
		panic(fmt.Errorf("unable to create synthetic input from identity type %v", conflictedInfo.Type))
	}
	return input
}

func (i *IntentAccountLinking) getOptions() []AccountLinkingIdentificationOptionInternal {
	return slice.FlatMap(i.ConflictingIdentities, func(identity *identity.Info) []AccountLinkingIdentificationOptionInternal {
		var identifcation config.AuthenticationFlowIdentification
		var maskedDisplayName string
		var providerType config.OAuthSSOProviderType
		var providerAlias string

		switch identity.Type {
		case model.IdentityTypeLoginID:
			switch identity.LoginID.LoginIDType {
			case model.LoginIDKeyTypeEmail:
				identifcation = config.AuthenticationFlowIdentificationEmail
				maskedDisplayName = mail.MaskAddress(identity.LoginID.LoginID)
			case model.LoginIDKeyTypePhone:
				identifcation = config.AuthenticationFlowIdentificationPhone
				maskedDisplayName = phone.Mask(identity.LoginID.LoginID)
			case model.LoginIDKeyTypeUsername:
				identifcation = config.AuthenticationFlowIdentificationUsername
				maskedDisplayName = identity.LoginID.LoginID
			}
		case model.IdentityTypeOAuth:
			identifcation = config.AuthenticationFlowIdentificationOAuth
			providerType = config.OAuthSSOProviderType(identity.OAuth.ProviderID.Type)
			providerAlias = identity.OAuth.ProviderAlias
		default:
			// Other types are not supported in account linking, exclude them in options
			return []AccountLinkingIdentificationOptionInternal{}
		}

		return []AccountLinkingIdentificationOptionInternal{{
			AccountLinkingIdentificationOption: AccountLinkingIdentificationOption{
				Identifcation:     identifcation,
				MaskedDisplayName: maskedDisplayName,
				ProviderType:      providerType,
				Alias:             providerAlias,
			},
			Identity: identity,
		}}
	})
}
