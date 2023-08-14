package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentVerifyClaim{})
}

var IntentVerifyClaimSchema = validation.NewSimpleSchema(`{}`)

type IntentVerifyClaim struct {
	UserID      string          `json:"user_id,omitempty"`
	Purpose     otp.Purpose     `json:"purpose,omitempty"`
	MessageType otp.MessageType `json:"message_type,omitempty"`
	ClaimName   model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue  string          `json:"claim_value,omitempty"`
}

var _ MilestoneDoMarkClaimVerified = &IntentVerifyClaim{}

func (*IntentVerifyClaim) Milestone()                    {}
func (*IntentVerifyClaim) MilestoneDoMarkClaimVerified() {}

var _ workflow.Intent = &IntentVerifyClaim{}

func (*IntentVerifyClaim) Kind() string {
	return "workflowconfig.IntentVerifyClaim"
}

func (*IntentVerifyClaim) JSONSchema() *validation.SimpleSchema {
	return IntentVerifyClaimSchema
}

func (*IntentVerifyClaim) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeOOBOTPChannel{},
		}, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentVerifyClaim) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPChannel inputTakeOOBOTPChannel
	if workflow.AsInput(input, &inputTakeOOBOTPChannel) {
		channel := inputTakeOOBOTPChannel.GetChannel()
		err := i.check(deps, channel)
		if err != nil {
			return nil, err
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
		err = node.SendCode(ctx, deps)
		if ratelimit.IsRateLimitErrorWithBucketName(err, kind.RateLimitTriggerCooldown(node.otpTarget()).Name) {
			// Ignore trigger cooldown rate limit error; continue the workflow
		} else if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(node), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentVerifyClaim) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*IntentVerifyClaim) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentVerifyClaim) check(deps *workflow.Dependencies, channel model.AuthenticatorOOBChannel) error {
	// 1. Check if the channel is compatible with the claim name.
	// 2. Check if the channel is compatible with the config.
	switch i.ClaimName {
	case model.ClaimEmail:
		if channel == model.AuthenticatorOOBChannelEmail {
			return nil
		}
	case model.ClaimPhoneNumber:
		switch channel {
		case model.AuthenticatorOOBChannelSMS:
			switch deps.Config.Authenticator.OOB.SMS.PhoneOTPMode {
			case config.AuthenticatorPhoneOTPModeWhatsappSMS:
				return nil
			case config.AuthenticatorPhoneOTPModeSMSOnly:
				return nil
			}
		case model.AuthenticatorOOBChannelWhatsapp:
			switch deps.Config.Authenticator.OOB.SMS.PhoneOTPMode {
			case config.AuthenticatorPhoneOTPModeWhatsappSMS:
				return nil
			case config.AuthenticatorPhoneOTPModeWhatsappOnly:
				return nil
			}
		}
	}
	return InvalidOOBOTPChannel.NewWithInfo("invalid channel", apierrors.Details{
		"claim_name":     i.ClaimName,
		"phone_otp_mode": deps.Config.Authenticator.OOB.SMS.PhoneOTPMode,
		"channel":        channel,
	})
}
