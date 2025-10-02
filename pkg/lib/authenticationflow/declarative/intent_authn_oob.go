package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

func init() {
	authflow.RegisterIntent(&IntentAuthenticationOOB{})
}

type IntentAuthenticationOOB struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	Purpose        otp.Purpose                            `json:"purpose,omitempty"`
	Form           otp.Form                               `json:"form,omitempty"`
	Info           *authenticator.Info                    `json:"info,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentAuthenticationOOB{}
var _ authflow.DataOutputer = &IntentAuthenticationOOB{}
var _ authflow.Milestone = &IntentAuthenticationOOB{}
var _ MilestoneDoMarkClaimVerified = &IntentAuthenticationOOB{}

func (*IntentAuthenticationOOB) Kind() string {
	return "IntentAuthenticationOOB"
}

func (*IntentAuthenticationOOB) Milestone()                    {}
func (*IntentAuthenticationOOB) MilestoneDoMarkClaimVerified() {}
func (i *IntentAuthenticationOOB) MilestoneDoMarkClaimVerifiedUpdateUserID(newUserID string) {
	i.UserID = newUserID
}

func (i *IntentAuthenticationOOB) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, verified := authflow.FindMilestoneInCurrentFlow[MilestoneOOBOTPVerified](flows)
	if !verified {
		// We have a special case here.
		// If there is only one channel, we do not take any input.
		// The rationale is that the only possible input is that channel.
		// So it is trivial that we can proceed without the input.
		channels := i.getChannels(deps)
		if len(channels) == 1 {
			return nil, nil
		}
		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
		if err != nil {
			return nil, err
		}
		return &InputSchemaTakeOOBOTPChannel{
			JSONPointer:    i.JSONPointer,
			FlowRootObject: flowRootObject,
			Channels:       channels,
		}, nil
	}

	_, _, lastUsedChannelUpdated := authflow.FindMilestoneInCurrentFlow[MilestoneOOBOTPLastUsedChannelUpdated](flows)
	if !lastUsedChannelUpdated {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAuthenticationOOB) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	milestone, _, verified := authflow.FindMilestoneInCurrentFlow[MilestoneOOBOTPVerified](flows)
	if !verified {
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

		node, err := NewNodeAuthenticationOOB(ctx, deps, &NodeAuthenticationOOB{
			JSONPointer:    i.JSONPointer,
			UserID:         i.UserID,
			Purpose:        i.Purpose,
			Form:           i.Form,
			Info:           i.Info,
			Channel:        channel,
			Authentication: i.Authentication,
		})
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	_, _, lastUsedChannelUpdated := authflow.FindMilestoneInCurrentFlow[MilestoneOOBOTPLastUsedChannelUpdated](flows)
	if !lastUsedChannelUpdated {
		return authflow.NewNodeSimple(&NodeDoUpdateLastUsedChannel{
			Channel: milestone.MilestoneOOBOTPVerifiedChannel(),
			Info:    i.Info,
		}), nil
	}

	panic(fmt.Errorf("unexpected: unreachable code reached"))
}

func (i *IntentAuthenticationOOB) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	channels := i.getChannels(deps)
	claimName, claimValue := i.Info.OOBOTP.ToClaimPair()
	return NewOOBData(SelectOOBOTPChannelsData{
		Channels:         channels,
		MaskedClaimValue: getMaskedOTPTarget(claimName, claimValue),
	}), nil
}

func (i *IntentAuthenticationOOB) getChannels(deps *authflow.Dependencies) []model.AuthenticatorOOBChannel {
	claimName, _ := i.Info.OOBOTP.ToClaimPair()
	return getChannels(claimName, deps.Config.Authenticator.OOB)
}
