package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepSelectDestination{})
}

type intentAccountRecoveryFlowStepSelectDestinationData struct {
	Options []AccountRecoveryDestinationOption `json:"options"`
}

var _ authflow.Data = intentAccountRecoveryFlowStepSelectDestinationData{}

func (intentAccountRecoveryFlowStepSelectDestinationData) Data() {}

type IntentAccountRecoveryFlowStepSelectDestination struct {
	JSONPointer jsonpointer.T                              `json:"json_pointer,omitempty"`
	StepName    string                                     `json:"step_name,omitempty"`
	Options     []AccountRecoveryDestinationOptionInternal `json:"options"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepSelectDestination{}

func (i *IntentAccountRecoveryFlowStepSelectDestination) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepSelectDestination{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepSelectDestination{}

func NewIntentAccountRecoveryFlowStepSelectDestination(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	i *IntentAccountRecoveryFlowStepSelectDestination,
) (*IntentAccountRecoveryFlowStepSelectDestination, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryIdentity](flows.Root)
	if !ok {
		return nil, InvalidFlowConfig.New("IntentAccountRecoveryFlowStepSelectDestination depends on MilestoneDoUseAccountRecoveryIdentity")
	}
	iden := milestone.MilestoneDoUseAccountRecoveryIdentity()
	step := i.step(current)

	optionsByUniqueKey := map[string]AccountRecoveryDestinationOptionInternal{}

	// Always include the user inputted login id
	switch iden.Identification {
	case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
		optionsByUniqueKey[iden.IdentitySpec.LoginID.Value] = AccountRecoveryDestinationOptionInternal{
			AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
				MaskedDisplayName: mail.MaskAddress(iden.IdentitySpec.LoginID.Value),
				Channel:           AccountRecoveryChannelEmail,
			},
			TargetLoginID: iden.IdentitySpec.LoginID.Value,
		}
	case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
		optionsByUniqueKey[iden.IdentitySpec.LoginID.Value] = AccountRecoveryDestinationOptionInternal{
			AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
				MaskedDisplayName: phone.Mask(iden.IdentitySpec.LoginID.Value),
				Channel:           AccountRecoveryChannelSMS,
			},
			TargetLoginID: iden.IdentitySpec.LoginID.Value,
		}
	}

	if iden.MaybeIdentity != nil && step.EnumerateDestinations {
		userID := iden.MaybeIdentity.UserID
		userIdens, err := deps.Identities.ListByUser(userID)
		if err != nil {
			return nil, err
		}
		for _, iden := range userIdens {
			if iden.Type != model.IdentityTypeLoginID {
				continue
			}
			switch iden.LoginID.LoginIDType {
			case model.LoginIDKeyTypeEmail:
				optionsByUniqueKey[iden.LoginID.LoginID] = AccountRecoveryDestinationOptionInternal{
					AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
						MaskedDisplayName: mail.MaskAddress(iden.LoginID.LoginID),
						Channel:           AccountRecoveryChannelEmail,
					},
					TargetLoginID: iden.LoginID.LoginID,
				}
			case model.LoginIDKeyTypePhone:
				optionsByUniqueKey[iden.LoginID.LoginID] = AccountRecoveryDestinationOptionInternal{
					AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
						MaskedDisplayName: phone.Mask(iden.LoginID.LoginID),
						Channel:           AccountRecoveryChannelSMS,
					},
					TargetLoginID: iden.LoginID.LoginID,
				}
			}
		}
	}

	options := []AccountRecoveryDestinationOptionInternal{}
	idx := 0
	for _, op := range optionsByUniqueKey {

		options = append(options, AccountRecoveryDestinationOptionInternal{
			AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
				MaskedDisplayName: op.MaskedDisplayName,
				Channel:           op.Channel,
			},
			TargetLoginID: op.TargetLoginID,
		})
		idx = idx + 1
	}

	i.Options = options
	return i, nil
}

func (*IntentAccountRecoveryFlowStepSelectDestination) Kind() string {
	return "IntentAccountRecoveryFlowStepSelectDestination"
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return &InputSchemaStepAccountRecoverySelectDestination{
			JSONPointer: i.JSONPointer,
			Options:     i.getOptions(),
		}, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryDestinationOptionIndex inputTakeAccountRecoveryDestinationOptionIndex
		if authflow.AsInput(input, &inputTakeAccountRecoveryDestinationOptionIndex) {
			optionIdx := inputTakeAccountRecoveryDestinationOptionIndex.GetAccountRecoveryDestinationOptionIndex()
			option := i.Options[optionIdx]
			return authflow.NewNodeSimple(&NodeUseAccountRecoveryDestination{
				TargetLoginID: option.TargetLoginID,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return intentAccountRecoveryFlowStepSelectDestinationData{
		Options: i.getOptions(),
	}, nil
}

func (*IntentAccountRecoveryFlowStepSelectDestination) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) getOptions() []AccountRecoveryDestinationOption {

	ops := []AccountRecoveryDestinationOption{}
	for _, op := range i.Options {
		ops = append(ops, op.AccountRecoveryDestinationOption)
	}
	return ops
}
