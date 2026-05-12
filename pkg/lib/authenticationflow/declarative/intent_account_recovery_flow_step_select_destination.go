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
	TypedData
	Options []AccountRecoveryDestinationOption `json:"options"`
}

func NewIntentAccountRecoveryFlowStepSelectDestinationData(d IntentAccountRecoveryFlowStepSelectDestinationData) IntentAccountRecoveryFlowStepSelectDestinationData {
	d.Type = DataTypeAccountRecoverySelectDestinationData
	return d
}

var _ authflow.Data = IntentAccountRecoveryFlowStepSelectDestinationData{}

func (IntentAccountRecoveryFlowStepSelectDestinationData) Data() {}

type IntentAccountRecoveryFlowStepSelectDestination struct {
	FlowReference authflow.FlowReference                      `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                               `json:"json_pointer,omitempty"`
	StepName      string                                      `json:"step_name,omitempty"`
	Options       []*AccountRecoveryDestinationOptionInternal `json:"options"`
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
	originNode authflow.NodeOrIntent,
) (*IntentAccountRecoveryFlowStepSelectDestination, error) {
	current, err := i.currentFlowObject(deps, flows, originNode)
	if err != nil {
		return nil, err
	}

	ms := authflow.FindAllMilestones[MilestoneDoUseAccountRecoveryIdentity](flows.Root)
	if len(ms) == 0 {
		return nil, InvalidFlowConfig.New("IntentAccountRecoveryFlowStepSelectDestination depends on MilestoneDoUseAccountRecoveryIdentity")
	}
	milestone := ms[0]

	iden := milestone.MilestoneDoUseAccountRecoveryIdentity()
	step := i.step(current)

	options, err := deriveAccountRecoveryDestinationOptions(
		ctx,
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
		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
		if err != nil {
			return nil, err
		}
		return &InputSchemaStepAccountRecoverySelectDestination{
			JSONPointer:    i.JSONPointer,
			FlowRootObject: flowRootObject,
			Options:        i.getOptions(),
		}, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryDestinationOptionIndex inputTakeAccountRecoveryDestinationOptionIndex
		if authflow.AsInput(input, &inputTakeAccountRecoveryDestinationOptionIndex) {
			optionIdx := inputTakeAccountRecoveryDestinationOptionIndex.GetAccountRecoveryDestinationOptionIndex()
			option := i.Options[optionIdx]
			resolved, err := i.resolveUsernameTarget(ctx, deps, flows, option)
			if err != nil {
				return nil, err
			}
			return authflow.NewNodeSimple(&NodeUseAccountRecoveryDestination{
				Destination: resolved,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

// resolveUsernameTarget is called only for username + enumerate_destinations=false flows.
// It returns the option unchanged for all other flows.
// When the user is found and has an identity matching the picked channel, TargetLoginID
// is replaced with that identity's login ID value so SendCode delivers to the right address.
// In all other cases (user not found, or no matching identity for the channel) TargetLoginID
// is prefixed with accountRecoveryNoSendPrefix so SendCode always hits its generateDummyOTP
// path — no message is dispatched but rate limits are still charged per username, and a
// username that looks like an email cannot accidentally dispatch to a different user.
func (i *IntentAccountRecoveryFlowStepSelectDestination) resolveUsernameTarget(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	option *AccountRecoveryDestinationOptionInternal,
) (*AccountRecoveryDestinationOptionInternal, error) {
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}
	step := i.step(current)
	if step.EnumerateDestinations {
		return option, nil
	}

	ms := authflow.FindAllMilestones[MilestoneDoUseAccountRecoveryIdentity](flows.Root)
	if len(ms) == 0 {
		return option, nil
	}
	accIden := ms[0].MilestoneDoUseAccountRecoveryIdentity()
	if accIden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationUsername {
		return option, nil
	}

	noSend := func() *AccountRecoveryDestinationOptionInternal {
		copied := *option
		copied.TargetLoginID = accountRecoveryNoSendPrefix + option.TargetLoginID
		return &copied
	}

	if accIden.MaybeIdentity == nil {
		return noSend(), nil
	}

	userIdens, err := deps.Identities.ListByUser(ctx, accIden.MaybeIdentity.UserID)
	if err != nil {
		return nil, err
	}
	if target := firstMatchingLoginIDForChannel(userIdens, option.Channel); target != "" {
		copied := *option
		copied.TargetLoginID = target
		return &copied, nil
	}
	return noSend(), nil
}

// firstMatchingLoginIDForChannel returns the first login-id value among userIdens
// whose login-id type maps to the requested account-recovery channel.
// email → LoginIDKeyTypeEmail. sms/whatsapp → LoginIDKeyTypePhone.
// Returns "" when no matching identity is present.
func firstMatchingLoginIDForChannel(
	userIdens []*identity.Info,
	channel AccountRecoveryChannel,
) string {
	var wantType model.LoginIDKeyType
	switch channel {
	case AccountRecoveryChannelEmail:
		wantType = model.LoginIDKeyTypeEmail
	case AccountRecoveryChannelSMS, AccountRecoveryChannelWhatsapp:
		wantType = model.LoginIDKeyTypePhone
	default:
		return ""
	}
	for _, ui := range userIdens {
		if ui.Type != model.IdentityTypeLoginID {
			continue
		}
		if ui.LoginID.LoginIDType == wantType {
			return ui.LoginID.LoginID
		}
	}
	return ""
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIntentAccountRecoveryFlowStepSelectDestinationData(IntentAccountRecoveryFlowStepSelectDestinationData{
		Options: i.getOptions(),
	}), nil
}

func (*IntentAccountRecoveryFlowStepSelectDestination) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, originNode authflow.NodeOrIntent) (config.AuthenticationFlowObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}

func (i *IntentAccountRecoveryFlowStepSelectDestination) getOptions() []AccountRecoveryDestinationOption {

	ops := []AccountRecoveryDestinationOption{}
	for _, op := range i.Options {
		ops = append(ops, op.AccountRecoveryDestinationOption)
	}
	return ops
}

func deriveAccountRecoveryDestinationOptions(
	ctx context.Context,
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	iden AccountRecoveryIdentity,
	deps *authflow.Dependencies,
) ([]*AccountRecoveryDestinationOptionInternal, error) {
	allowedChannels := step.AllowedChannels
	if allowedChannels == nil || len(allowedChannels) == 0 {
		allowedChannels = config.GetAllAccountRecoveryChannel()
	}

	options := []*AccountRecoveryDestinationOptionInternal{}

	isUsername := iden.Identification == config.AuthenticationFlowAccountRecoveryIdentificationUsername

	switch {
	case iden.MaybeIdentity != nil && step.EnumerateDestinations:
		userID := iden.MaybeIdentity.UserID
		userIdens, err := deps.Identities.ListByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, channel := range allowedChannels {
			opts := enumerateAllowedAccountRecoveryDestinationOptions(channel, userIdens)
			options = append(options, opts...)
		}
	case isUsername && !step.EnumerateDestinations:
		// One option per allowed channel. TargetLoginID is the typed username and will be
		// resolved to the user's actual email/phone at ReactTo time when the user picks one.
		username := iden.IdentitySpec.LoginID.Value.TrimSpace()
		for _, channel := range allowedChannels {
			options = append(options, &AccountRecoveryDestinationOptionInternal{
				AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
					MaskedDisplayName: username,
					Channel:           AccountRecoveryChannel(channel.Channel),
					OTPForm:           AccountRecoveryOTPForm(channel.OTPForm),
				},
				TargetLoginID: username,
			})
		}
	default:
		// Existing email/phone non-enumerate path. Also covers username + enumerate=true + user not found.
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
					MaskedDisplayName: mail.MaskAddress(iden.IdentitySpec.LoginID.Value.TrimSpace()),
					Channel:           AccountRecoveryChannelEmail,
					OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
				},
				TargetLoginID: iden.IdentitySpec.LoginID.Value.TrimSpace(),
			},
		}
	case config.AccountRecoveryCodeChannelSMS:
		if iden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationPhone {
			return []*AccountRecoveryDestinationOptionInternal{}
		}
		return []*AccountRecoveryDestinationOptionInternal{
			{
				AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
					MaskedDisplayName: phone.Mask(iden.IdentitySpec.LoginID.Value.TrimSpace()),
					Channel:           AccountRecoveryChannelSMS,
					OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
				},
				TargetLoginID: iden.IdentitySpec.LoginID.Value.TrimSpace(),
			},
		}
	case config.AccountRecoveryCodeChannelWhatsapp:
		if iden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationPhone {
			return []*AccountRecoveryDestinationOptionInternal{}
		}
		return []*AccountRecoveryDestinationOptionInternal{
			{
				AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
					MaskedDisplayName: phone.Mask(iden.IdentitySpec.LoginID.Value.TrimSpace()),
					Channel:           AccountRecoveryChannelWhatsapp,
					OTPForm:           AccountRecoveryOTPForm(allowedChannel.OTPForm),
				},
				TargetLoginID: iden.IdentitySpec.LoginID.Value.TrimSpace(),
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
			if allowedChannel.Channel == config.AccountRecoveryCodeChannelWhatsapp {
				newOption := &AccountRecoveryDestinationOptionInternal{
					AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
						MaskedDisplayName: phone.Mask(iden.LoginID.LoginID),
						Channel:           AccountRecoveryChannelWhatsapp,
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
