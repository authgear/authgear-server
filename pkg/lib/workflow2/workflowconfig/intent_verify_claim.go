package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentVerifyClaim{})
}

type IntentVerifyClaim struct {
	UserID      string          `json:"user_id,omitempty"`
	Purpose     otp.Purpose     `json:"purpose,omitempty"`
	MessageType otp.MessageType `json:"message_type,omitempty"`
	ClaimName   model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue  string          `json:"claim_value,omitempty"`
}

var _ workflow.Intent = &IntentVerifyClaim{}
var _ workflow.Milestone = &IntentVerifyClaim{}
var _ MilestoneDoMarkClaimVerified = &IntentVerifyClaim{}

func (*IntentVerifyClaim) Kind() string {
	return "workflowconfig.IntentVerifyClaim"
}

func (*IntentVerifyClaim) Milestone()                    {}
func (*IntentVerifyClaim) MilestoneDoMarkClaimVerified() {}

func (i *IntentVerifyClaim) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		channels := i.getChannels(deps)
		return &InputSchemaTakeOOBOTPChannel{
			Channels: channels,
		}, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentVerifyClaim) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPChannel inputTakeOOBOTPChannel
	if workflow.AsInput(input, &inputTakeOOBOTPChannel) {
		channel := inputTakeOOBOTPChannel.GetChannel()
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
			// Ignore trigger cooldown rate limit error; continue the workflow
		} else if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(node), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentVerifyClaim) getChannels(deps *workflow.Dependencies) []model.AuthenticatorOOBChannel {
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
