package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepSelectDestination{})
}

type IntentAccountRecoveryFlowStepSelectDestinationData struct {
	Options []AccountRecoveryDestinationOption `json:"options"`
}

var _ authflow.Data = IntentAccountRecoveryFlowStepSelectDestinationData{}

func (IntentAccountRecoveryFlowStepSelectDestinationData) Data() {}

type IntentAccountRecoveryFlowStepSelectDestination struct {
	JSONPointer jsonpointer.T                               `json:"json_pointer,omitempty"`
	StepName    string                                      `json:"step_name,omitempty"`
	Options     []*AccountRecoveryDestinationOptionInternal `json:"options"`
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

	options, err := deriveAccountRecoveryDestinationOptions(
		step,
		iden,
		deps,
	)
	if err != nil {
		return nil, err
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
				Destination: option,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentAccountRecoveryFlowStepSelectDestinationData{
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

func deriveAccountRecoveryDestinationOptions(
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	iden AccountRecoveryIdentity,
	deps *authflow.Dependencies,
) ([]*AccountRecoveryDestinationOptionInternal, error) {
	allowedChannels := step.AllowedChannels
	if allowedChannels == nil || len(allowedChannels) == 0 {
		allowedChannels = config.GetAllAccountRecoveryChannel()
	}

	options := []*AccountRecoveryDestinationOptionInternal{}

	if iden.MaybeIdentity != nil && step.EnumerateDestinations {
		userID := iden.MaybeIdentity.UserID
		userIdens, err := deps.Identities.ListByUser(userID)
		if err != nil {
			return nil, err
		}
		for _, channel := range allowedChannels {
			opts := enumerateAllowedAccountRecoveryDestinationOptions(channel, userIdens)
			options = append(options, opts...)
		}
	} else {
		for _, channel := range allowedChannels {
			opts := deriveAllowedAccountRecoveryDestinationOptions(channel, iden)
			options = append(options, opts...)
		}
	}
	return options, nil
}

func deriveAllowedAccountRecoveryDestinationOptions(
	allowedChannel *config.AccountRecoveryChannel,
	iden AccountRecoveryIdentity,
) []*AccountRecoveryDestinationOptionInternal {
	switch allowedChannel.Channel {
	case config.AccountRecoveryCodeChannelEmail:
		if iden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationEmail {
			return []*AccountRecoveryDestinationOptionInternal{}
		}
		return []*AccountRecoveryDestinationOptionInternal{
			{
				AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
					MaskedDisplayName: mail.MaskAddress(iden.IdentitySpec.LoginID.Value),
					Channel:           AccountRecoveryChannelEmail,
					OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
				},
				TargetLoginID: iden.IdentitySpec.LoginID.Value,
			},
		}
	case config.AccountRecoveryCodeChannelSMS:
		if iden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationPhone {
			return []*AccountRecoveryDestinationOptionInternal{}
		}
		return []*AccountRecoveryDestinationOptionInternal{
			{
				AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
					MaskedDisplayName: phone.Mask(iden.IdentitySpec.LoginID.Value),
					Channel:           AccountRecoveryChannelSMS,
					OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
				},
				TargetLoginID: iden.IdentitySpec.LoginID.Value,
			},
		}
	}
	panic("account recovery: unknown allowed channel")
}

func enumerateAllowedAccountRecoveryDestinationOptions(
	allowedChannel *config.AccountRecoveryChannel,
	userIdens []*identity.Info,
) []*AccountRecoveryDestinationOptionInternal {
	options := []*AccountRecoveryDestinationOptionInternal{}
	for _, iden := range userIdens {
		if iden.Type != model.IdentityTypeLoginID {
			continue
		}
		switch iden.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			if allowedChannel.Channel == config.AccountRecoveryCodeChannelEmail {
				newOption := &AccountRecoveryDestinationOptionInternal{
					AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
						MaskedDisplayName: mail.MaskAddress(iden.LoginID.LoginID),
						Channel:           AccountRecoveryChannelEmail,
						OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
					},
					TargetLoginID: iden.LoginID.LoginID,
				}
				options = append(options, newOption)
			}
		case model.LoginIDKeyTypePhone:
			if allowedChannel.Channel == config.AccountRecoveryCodeChannelSMS {
				newOption := &AccountRecoveryDestinationOptionInternal{
					AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
						MaskedDisplayName: phone.Mask(iden.LoginID.LoginID),
						Channel:           AccountRecoveryChannelSMS,
						OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
					},
					TargetLoginID: iden.LoginID.LoginID,
				}
				options = append(options, newOption)
			}
		}
	}
	return options
}
