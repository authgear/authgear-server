package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateEmailLoginLink{})
}

var IntentAuthenticateEmailLoginLinkSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (i *IntentAuthenticateEmailLoginLink) Kind() string {
	return "latte.IntentAuthenticateEmailLoginLink"
}

func (i *IntentAuthenticateEmailLoginLink) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateEmailLoginLinkSchema
}

func (i *IntentAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		authenticator := i.Authenticator
		_, err := (&SendOOBCode{
			Deps:                 deps,
			Stage:                authenticatorKindToStage(authenticator.Kind),
			IsAuthenticating:     true,
			AuthenticatorInfo:    authenticator,
			IgnoreRatelimitError: true,
			OTPMode:              otp.OTPModeMagicLink,
		}).Do()
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeAuthenticateEmailLoginLink{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	bucket := deps.AntiSpamOTPCodeBucket.MakeBucket(model.AuthenticatorOOBChannelEmail, i.Authenticator.OOBOTP.Email)
	pass, resetDuration, err := deps.RateLimiter.CheckToken(bucket)
	if err != nil {
		return nil, err
	}
	var sec int
	if pass {
		// allow sending immediately
		sec = 0
	} else {
		sec = int(resetDuration.Seconds())
	}

	return map[string]interface{}{
		"cooldown_sec": sec,
	}, nil
}

func (i *IntentAuthenticateEmailLoginLink) GetAMR() []string {
	return i.Authenticator.AMR()
}

var _ AMRGetter = &IntentAuthenticateEmailLoginLink{}
