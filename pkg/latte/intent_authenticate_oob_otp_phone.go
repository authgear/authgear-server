package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentAuthenticateOOBOTPPhone{})
}

var IntentAuthenticateOOBOTPPhoneSchema = validation.NewSimpleSchema(`{}`)

type IntentAuthenticateOOBOTPPhone struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (i *IntentAuthenticateOOBOTPPhone) Kind() string {
	return "latte.IntentAuthenticateOOBOTPPhone"
}

func (i *IntentAuthenticateOOBOTPPhone) JSONSchema() *validation.SimpleSchema {
	return IntentAuthenticateOOBOTPPhoneSchema
}

func (i *IntentAuthenticateOOBOTPPhone) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentAuthenticateOOBOTPPhone) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	switch len(w.Nodes) {
	case 0:
		authenticator := i.Authenticator
		_, err := (&SendOOBCode{
			Deps:                 deps,
			Stage:                authenticatorKindToStage(authenticator.Kind),
			IsAuthenticating:     true,
			AuthenticatorInfo:    authenticator,
			IgnoreRatelimitError: true,
		}).Do()
		if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeAuthenticateOOBOTPPhone{
			Authenticator: authenticator,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentAuthenticateOOBOTPPhone) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentAuthenticateOOBOTPPhone) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	bucket := deps.AntiSpamOTPCodeBucket.MakeBucket(model.AuthenticatorOOBChannelSMS, i.Authenticator.OOBOTP.Phone)
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

func (i *IntentAuthenticateOOBOTPPhone) GetAMR() []string {
	return i.Authenticator.AMR()
}

var _ AMRGetter = &IntentAuthenticateEmailLoginLink{}
