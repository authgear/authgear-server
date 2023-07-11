package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifyClaimSelectChannel{})
}

type NodeVerifyClaimSelectChannel struct {
	UserID      string          `json:"user_id,omitempty"`
	Purpose     otp.Purpose     `json:"purpose,omitempty"`
	MessageType otp.MessageType `json:"message_type,omitempty"`
	ClaimName   model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue  string          `json:"claim_value,omitempty"`
}

func (n *NodeVerifyClaimSelectChannel) Kind() string {
	return "workflowconfig.NodeVerifyClaimSelectChannel"
}

func (n *NodeVerifyClaimSelectChannel) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeVerifyClaimSelectChannel) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPChannel{},
	}, nil
}

func (n *NodeVerifyClaimSelectChannel) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPChannel inputTakeOOBOTPChannel

	switch {
	case workflow.AsInput(input, &inputTakeOOBOTPChannel):
		channel := inputTakeOOBOTPChannel.GetChannel()
		err := n.check(deps, channel)
		if err != nil {
			return nil, err
		}
		node := &NodeVerifyClaim{
			UserID:      n.UserID,
			Purpose:     n.Purpose,
			MessageType: n.MessageType,
			ClaimName:   n.ClaimName,
			ClaimValue:  n.ClaimValue,
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
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (n *NodeVerifyClaimSelectChannel) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (n *NodeVerifyClaimSelectChannel) check(deps *workflow.Dependencies, channel model.AuthenticatorOOBChannel) error {
	// 1. Check if the channel is compatible with the claim name.
	// 2. Check if the channel is compatible with the config.
	switch n.ClaimName {
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
		"claim_name":     n.ClaimName,
		"phone_otp_mode": deps.Config.Authenticator.OOB.SMS.PhoneOTPMode,
		"channel":        channel,
	})
}
