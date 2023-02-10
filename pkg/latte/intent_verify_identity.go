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

var IntentVerifyIdentitySchema = validation.NewSimpleSchema(`{}`)

type IntentVerifyIdentity struct {
	IdentityInfo *identity.Info
}

func (*IntentVerifyIdentity) Kind() string {
	return "latte.IntentVerifyIdentity"
}

func (*IntentVerifyIdentity) JSONSchema() *validation.SimpleSchema {
	return IntentVerifyIdentitySchema
}

func (*IntentVerifyIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentVerifyIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	statuses, err := deps.Verification.GetIdentityVerificationStatus(i.IdentityInfo)
	if err != nil {
		return nil, err
	}

	var status *verification.ClaimStatus
	if len(statuses) > 0 {
		status = &statuses[0]
	}

	if status == nil || !status.IsVerifiable() {
		return nil, api.ErrClaimNotVerifiable
	}

	if status.Verified {
		// Verified already; skip actual verification.
		return workflow.NewNodeSimple(&NodeVerifiedIdentity{
			IdentityID:       i.IdentityInfo.ID,
			NewVerifiedClaim: nil,
		}), nil
	}

	var node interface {
		workflow.NodeSimple
		sendCode(deps *workflow.Dependencies, w *workflow.Workflow) error
	}
	switch model.ClaimName(status.Name) {
	case model.ClaimEmail:
		node = &NodeVerifyEmail{
			UserID:     i.IdentityInfo.UserID,
			IdentityID: i.IdentityInfo.ID,
			Email:      status.Value,
		}

	case model.ClaimPhoneNumber:
		node = &NodeVerifyPhoneSMS{
			UserID:      i.IdentityInfo.UserID,
			IdentityID:  i.IdentityInfo.ID,
			PhoneNumber: status.Value,
		}
		// FIXME(workflow): verify phone via whatsapp
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
