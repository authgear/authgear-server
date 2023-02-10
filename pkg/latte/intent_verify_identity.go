package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentVerifyIdentity{})
}

var IntentVerifyIdentitySchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"user_id": { "type": "string" }
		},
		"required": ["user_id"]
	}
`)

type IntentVerifyIdentity struct {
	UserID string `json:"user_id"`
}

func (*IntentVerifyIdentity) Kind() string {
	return "latte.IntentVerifyIdentity"
}

func (*IntentVerifyIdentity) JSONSchema() *validation.SimpleSchema {
	return IntentVerifyIdentitySchema
}

func (*IntentVerifyIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Input{
			&InputTriggerVerification{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentVerifyIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var trigger inputTriggerVerification

	switch {
	case workflow.AsInput(input, &trigger):
		claimName, claimValue := trigger.ClaimToVerify()
		identities, err := deps.Identities.ListByClaim(claimName, claimValue)
		if err != nil {
			return nil, err
		}

		var iden *identity.Info
		for _, ii := range identities {
			if ii.UserID == i.UserID {
				iden = ii
				break
			}
		}
		if iden == nil {
			// FIXME: define new identity not found error?
			return nil, api.ErrUserNotFound
		}

		statuses, err := deps.Verification.GetIdentityVerificationStatus(iden)
		if err != nil {
			return nil, err
		}
		var status *verification.ClaimStatus
		for _, s := range statuses {
			s := s
			if s.Name == claimName {
				status = &s
				break
			}
		}
		if status == nil || !status.IsVerifiable() {
			return nil, api.ErrClaimNotVerifiable
		}

		if status.Verified {
			// Verified already; skip actual verification.
			return workflow.NewNodeSimple(&NodeVerifiedIdentity{
				IdentityID:       iden.ID,
				NewVerifiedClaim: nil,
			}), nil
		}

		// TODO: refactor OTP mode to identity config?
		phoneOTPMode := deps.Config.Authenticator.OOB.SMS.PhoneOTPMode

		var node interface {
			workflow.NodeSimple
			sendCode(deps *workflow.Dependencies, w *workflow.Workflow) error
		}
		switch model.ClaimName(claimName) {
		case model.ClaimEmail:
			node = &NodeVerifyEmail{
				UserID:     i.UserID,
				IdentityID: iden.ID,
				Email:      claimValue,
			}

		case model.ClaimPhoneNumber:
			if trigger.VerificationMethod() == VerificationMethodPhoneSMS && phoneOTPMode.IsSMSEnabled() {
				node = &NodeVerifyPhoneSMS{
					UserID:      i.UserID,
					IdentityID:  iden.ID,
					PhoneNumber: claimValue,
				}
			}
			// FIXME: whatsapp
		}

		if node == nil {
			return nil, api.ErrClaimNotVerifiable
		}

		err = node.sendCode(deps, w)
		if apierrors.IsKind(err, ratelimit.RateLimited) {
			// Ignore rate limit error; continue the workflow
		} else if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(node), nil

	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (*IntentVerifyIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentVerifyIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (i *IntentVerifyIdentity) VerifiedIdentity(w *workflow.Workflow) (*NodeVerifiedIdentity, bool) {
	return workflow.FindSingleNode[*NodeVerifiedIdentity](w)
}
