package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentUseRecoveryCode{})
}

type IntentUseRecoveryCode struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseRecoveryCode{}
var _ authflow.Milestone = &IntentUseRecoveryCode{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentUseRecoveryCode{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentUseRecoveryCode{}
var _ MilestoneFlowAuthenticate = &IntentUseRecoveryCode{}
var _ authflow.InputReactor = &IntentUseRecoveryCode{}

func (*IntentUseRecoveryCode) Kind() string {
	return "IntentUseRecoveryCode"
}

func (*IntentUseRecoveryCode) Milestone() {}
func (n *IntentUseRecoveryCode) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentUseRecoveryCode) MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*IntentUseRecoveryCode) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
	if authenticated {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}

	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeRecoveryCode{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentUseRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeRecoveryCode inputTakeRecoveryCode
	if authflow.AsInput(input, &inputTakeRecoveryCode) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}
		recoveryCode := inputTakeRecoveryCode.GetRecoveryCode()

		rc, err := deps.MFA.VerifyRecoveryCode(ctx, n.UserID, recoveryCode)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoConsumeRecoveryCode{
			RecoveryCode: rc,
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
