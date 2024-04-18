package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func init() {
	authflow.RegisterIntent(&IntentVerifyClaim{})
}

type IntentVerifyClaim struct {
	JSONPointer jsonpointer.T   `json:"json_pointer,omitempty"`
	UserID      string          `json:"user_id,omitempty"`
	Purpose     otp.Purpose     `json:"purpose,omitempty"`
	MessageType otp.MessageType `json:"message_type,omitempty"`
	Form        otp.Form        `json:"form,omitempty"`
	ClaimName   model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue  string          `json:"claim_value,omitempty"`
}

var _ authflow.Intent = &IntentVerifyClaim{}
var _ authflow.DataOutputer = &IntentVerifyClaim{}
var _ authflow.Milestone = &IntentVerifyClaim{}
var _ MilestoneVerifyClaim = &IntentVerifyClaim{}

func (*IntentVerifyClaim) Kind() string {
	return "IntentVerifyClaim"
}

func (*IntentVerifyClaim) Milestone()            {}
func (*IntentVerifyClaim) MilestoneVerifyClaim() {}
func (i *IntentVerifyClaim) MilestoneVerifyClaimUpdateUserID(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	i.UserID = newUserID

	milestone, ok := authflow.FindFirstMilestone[MilestoneDoMarkClaimVerified](flow)
	if ok {
		milestone.MilestoneDoMarkClaimVerifiedUpdateUserID(newUserID)
	}

	return nil
}

func (i *IntentVerifyClaim) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		// We have a special case here.
		// If there is only one channel, we do not take any input.
		// The rationale is that the only possible input is that channel.
		// So it is trivial that we can proceed without the input.
		channels := i.getChannels(deps)
		if len(channels) == 1 {
			return nil, nil
		}
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaTakeOOBOTPChannel{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			Channels:       channels,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentVerifyClaim) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	channels := i.getChannels(deps)
	var inputTakeOOBOTPChannel inputTakeOOBOTPChannel
	var channel model.AuthenticatorOOBChannel

	switch {
	case len(channels) == 1:
		channel = channels[0]
	case authflow.AsInput(input, &inputTakeOOBOTPChannel):
		channel = inputTakeOOBOTPChannel.GetChannel()
	default:
		return nil, authflow.ErrIncompatibleInput
	}

	node := NewNodeVerifyClaim(&NodeVerifyClaim{
		JSONPointer: i.JSONPointer,
		UserID:      i.UserID,
		Purpose:     i.Purpose,
		MessageType: i.MessageType,
		Form:        i.Form,
		ClaimName:   i.ClaimName,
		ClaimValue:  i.ClaimValue,
		Channel:     channel,
	})
	kind := node.otpKind(deps)
	err := node.SendCode(ctx, deps)
	if ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(node.ClaimValue).Name) {
		// Ignore trigger cooldown rate limit error; continue the flow
	} else if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(node), nil
}

func (i *IntentVerifyClaim) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	channels := i.getChannels(deps)
	return NewOOBData(SelectOOBOTPChannelsData{
		Channels:         channels,
		MaskedClaimValue: getMaskedOTPTarget(i.ClaimName, i.ClaimValue),
	}), nil
}

func (i *IntentVerifyClaim) getChannels(deps *authflow.Dependencies) []model.AuthenticatorOOBChannel {
	return getChannels(i.ClaimName, deps.Config.Authenticator.OOB)
}
