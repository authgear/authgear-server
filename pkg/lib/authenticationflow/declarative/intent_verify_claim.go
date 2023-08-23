package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

func init() {
	authflow.RegisterIntent(&IntentVerifyClaim{})
}

type IntentVerifyClaim struct {
	UserID      string          `json:"user_id,omitempty"`
	Purpose     otp.Purpose     `json:"purpose,omitempty"`
	MessageType otp.MessageType `json:"message_type,omitempty"`
	ClaimName   model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue  string          `json:"claim_value,omitempty"`
}

var _ authflow.Intent = &IntentVerifyClaim{}
var _ authflow.Milestone = &IntentVerifyClaim{}
var _ MilestoneDoMarkClaimVerified = &IntentVerifyClaim{}

func (*IntentVerifyClaim) Kind() string {
	return "IntentVerifyClaim"
}

func (*IntentVerifyClaim) Milestone()                    {}
func (*IntentVerifyClaim) MilestoneDoMarkClaimVerified() {}

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
		return &InputSchemaTakeOOBOTPChannel{
			Channels: channels,
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

	node := &NodeVerifyClaim{
		UserID:      i.UserID,
		Purpose:     i.Purpose,
		MessageType: i.MessageType,
		ClaimName:   i.ClaimName,
		ClaimValue:  i.ClaimValue,
		Channel:     channel,
	}
	kind := node.otpKind(deps)
	err := node.SendCode(ctx, deps)
	if ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(node.otpTarget()).Name) {
		// Ignore trigger cooldown rate limit error; continue the flow
	} else if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(node), nil
}

func (i *IntentVerifyClaim) getChannels(deps *authflow.Dependencies) []model.AuthenticatorOOBChannel {
	email := false
	sms := false
	whatsapp := false

	switch i.ClaimName {
	case model.ClaimEmail:
		email = true
	case model.ClaimPhoneNumber:
		switch deps.Config.Authenticator.OOB.SMS.PhoneOTPMode {
		case config.AuthenticatorPhoneOTPModeSMSOnly:
			sms = true
		case config.AuthenticatorPhoneOTPModeWhatsappOnly:
			whatsapp = true
		case config.AuthenticatorPhoneOTPModeWhatsappSMS:
			sms = true
			whatsapp = true
		}
	}

	channels := []model.AuthenticatorOOBChannel{}
	if email {
		channels = append(channels, model.AuthenticatorOOBChannelEmail)
	}
	if sms {
		channels = append(channels, model.AuthenticatorOOBChannelSMS)
	}
	if whatsapp {
		channels = append(channels, model.AuthenticatorOOBChannelWhatsapp)
	}

	return channels
}
