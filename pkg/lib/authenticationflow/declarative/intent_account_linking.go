package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func init() {
	authflow.RegisterIntent(&IntentAccountLinking{})
}

type IntentAccountLinking struct {
	JSONPointer          jsonpointer.T             `json:"json_pointer,omitempty"`
	IncomingIdentitySpec *identity.Spec            `json:"incoming_identity_spec,omitempty"`
	Conflicts            []*AccountLinkingConflict `json:"conflicts,omitempty"`
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
	err := i.errorIfSomeConflictIsError()
	if err != nil {
		return nil, err
	}

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
	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountLinkingIdentification inputTakeAccountLinkingIdentification
		if authflow.AsInput(input, &inputTakeAccountLinkingIdentification) {
			idx := inputTakeAccountLinkingIdentification.GetAccountLinkingIdentificationIndex()
			redirectURI := inputTakeAccountLinkingIdentification.GetAccountLinkingOAuthRedirectURI()
			responseMode := inputTakeAccountLinkingIdentification.GetAccountLinkingOAuthResponseMode()
			selectedOption := i.getOptions()[idx]

			return authflow.NewNodeSimple(&NodeUseAccountLinkingIdentification{
				Option:       selectedOption.AccountLinkingIdentificationOption,
				Conflict:     selectedOption.Conflict,
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
	selectedConflict := milestone.MilestoneUseAccountLinkingIdentification()
	conflictedIdentity := selectedConflict.Identity
	conflictedUserID := conflictedIdentity.UserID
	if err != nil {
		return nil, err
	}
	if flowUserID != conflictedUserID {
		return i.rewriteFlowIntoUserIDOfConflictedIdentity(deps, flows, milestone)
	}

	switch len(flows.Nearest.Nodes) {
	case 1:
		var skipLogin bool
		var loginFlow string = selectedConflict.LoginFlow
		switch selectedConflict.Action {
		case config.AccountLinkingActionError:
			panic(fmt.Errorf("unexpected: conflict should be be choosable if action is error"))
			// When we support actions which can skip login, set skipLogin to true
		case config.AccountLinkingActionLoginAndLink:
			if loginFlow == "" {
				// Use the current flow name if it is not specified
				loginFlow = authflow.FindCurrentFlowReference(flows.Root).Name
			}
		default:
			skipLogin = false
		}
		if skipLogin {
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
		if loginFlow == "" {
			panic(fmt.Errorf("login_flow must be specified"))
		}
		flowReference := authflow.FlowReference{
			Type: authflow.FlowTypeLogin,
			Name: loginFlow,
		}
		loginIntent := IntentLoginFlow{
			TargetUserID:  conflictedUserID,
			FlowReference: flowReference,
		}
		return authflow.NewSubFlow(&loginIntent), nil
	case 2:
		info, err := newIdentityInfo(deps, conflictedIdentity.UserID, i.IncomingIdentitySpec)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountLinking) rewriteFlowIntoUserIDOfConflictedIdentity(
	deps *authflow.Dependencies,
	flows authflow.Flows,
	milestone MilestoneUseAccountLinkingIdentification) (*authflow.Node, error) {

	conflictedIdentity := milestone.MilestoneUseAccountLinkingIdentification().Identity
	conflictedUserID := conflictedIdentity.UserID
	err := authflow.TraverseFlow(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			milestone, ok := nodeSimple.(MilestoneSwitchToExistingUser)
			if ok {
				err := milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			milestone, ok := intent.(MilestoneSwitchToExistingUser)
			if ok {
				err := milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
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
		SyntheticInput: i.createSyntheticInputOAuthConflict(milestone, i.IncomingIdentitySpec, conflictedIdentity),
	}
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
	return slice.FlatMap(i.Conflicts, func(c *AccountLinkingConflict) []AccountLinkingIdentificationOptionInternal {
		var identifcation config.AuthenticationFlowIdentification
		var maskedDisplayName string
		var providerType string
		var providerAlias string

		identity := c.Identity

		if c.Action == config.AccountLinkingActionError {
			// We don't show error as options
			return []AccountLinkingIdentificationOptionInternal{}
		}

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
			providerType = identity.OAuth.ProviderID.Type
			maskedDisplayName = identity.OAuth.GetDisplayName()
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
				Action:            c.Action,
			},
			Conflict: c,
		}}
	})
}

func (i *IntentAccountLinking) errorIfSomeConflictIsError() error {
	var errorConflicts []*AccountLinkingConflict = []*AccountLinkingConflict{}
	for _, conflict := range i.Conflicts {
		conflict := conflict
		if conflict.Action == config.AccountLinkingActionError {
			errorConflicts = append(errorConflicts, conflict)
		}
	}

	// If there is at least one conflict with action=error,
	// return error
	if len(errorConflicts) > 0 {
		spec := i.IncomingIdentitySpec
		conflictSpecs := slice.Map(errorConflicts, func(c *AccountLinkingConflict) *identity.Spec {
			s := c.Identity.ToSpec()
			return &s
		})
		return identityFillDetailsMany(api.ErrDuplicatedIdentity, spec, conflictSpecs)
	}

	return nil
}
