package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	authflow.RegisterIntent(&IntentRequestAccountRecoveryFlowStepSelectDestination{})
}

type IntentRequestAccountRecoveryFlowStepSelectDestinationData struct {
	Options []AccountRecoveryDestinationOption `json:"options"`
}

var _ authflow.Data = IntentRequestAccountRecoveryFlowStepSelectDestinationData{}

func (IntentRequestAccountRecoveryFlowStepSelectDestinationData) Data() {}

type IntentRequestAccountRecoveryFlowStepSelectDestination struct {
	JSONPointer jsonpointer.T                              `json:"json_pointer,omitempty"`
	StepName    string                                     `json:"step_name,omitempty"`
	Options     []AccountRecoveryDestinationOptionInternal `json:"options"`
}

var _ authflow.TargetStep = &IntentRequestAccountRecoveryFlowStepSelectDestination{}

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) GetName() string {
	return i.StepName
}

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentRequestAccountRecoveryFlowStepSelectDestination{}
var _ authflow.DataOutputer = &IntentRequestAccountRecoveryFlowStepSelectDestination{}

func NewIntentRequestAccountRecoveryFlowStepSelectDestination(
	ctx context.Context,
	deps *authflow.Dependencies,
	parentFlow *authflow.Flow,
	i *IntentRequestAccountRecoveryFlowStepSelectDestination,
) (*IntentRequestAccountRecoveryFlowStepSelectDestination, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryIdentity](parentFlow)
	if !ok {
		return i, fmt.Errorf("IntentRequestAccountRecoveryFlowStepSelectDestination depends on MilestoneDoUseAccountRecoveryIdentity")
	}
	iden := milestone.MilestoneDoUseAccountRecoveryIdentity()
	step := i.step(current)

	optionsByUniqueKey := map[string]AccountRecoveryDestinationOptionInternal{}

	// Always include the user inputted login id
	switch iden.Identification {
	case config.AuthenticationFlowRequestAccountRecoveryIdentificationEmail:
		optionsByUniqueKey[iden.IdentitySpec.LoginID.Value] = AccountRecoveryDestinationOptionInternal{
			AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
				MaskedDisplayName: mail.MaskAddress(iden.IdentitySpec.LoginID.Value),
				Channel:           AccountRecoveryChannelEmail,
			},
			TargetLoginID: iden.IdentitySpec.LoginID.Value,
		}
	case config.AuthenticationFlowRequestAccountRecoveryIdentificationPhone:
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
				ID:                fmt.Sprintf("%d", idx),
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

func (*IntentRequestAccountRecoveryFlowStepSelectDestination) Kind() string {
	return "IntentRequestAccountRecoveryFlowStepSelectDestination"
}

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryDestinationOptionID inputTakeAccountRecoveryDestinationOptionID
		if authflow.AsInput(input, &inputTakeAccountRecoveryDestinationOptionID) {
			optionID := inputTakeAccountRecoveryDestinationOptionID.GetAccountRecoveryDestinationOptionID()
			var option *AccountRecoveryDestinationOptionInternal = nil
			for _, op := range i.Options {
				if op.ID == optionID {
					o := op
					option = &o
					break
				}
			}
			if option == nil {
				return nil, authflow.ErrIncompatibleInput
			}
			err := deps.ForgotPassword.SendCode(option.TargetLoginID)
			if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
				return nil, err
			}
			return authflow.NewNodeSimple(&NodeAccountRecoveryCodeSent{
				TargetLoginID: option.TargetLoginID,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentRequestAccountRecoveryFlowStepSelectDestinationData{
		Options: i.getOptions(),
	}, nil
}

func (*IntentRequestAccountRecoveryFlowStepSelectDestination) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowRequestAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowRequestAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentRequestAccountRecoveryFlowStepSelectDestination) getOptions() []AccountRecoveryDestinationOption {

	ops := []AccountRecoveryDestinationOption{}
	for _, op := range i.Options {
		ops = append(ops, op.AccountRecoveryDestinationOption)
	}
	return ops
}
