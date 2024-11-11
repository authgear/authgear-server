package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifiedAuthenticator{})
}

type NodeVerifiedAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeVerifiedAuthenticator) Kind() string {
	// FIXME: It should be NodeVerifiedAuthenticator
	return "latte.NodeAuthenticateOOBOTPPhoneEnd"
}

func (n *NodeVerifiedAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// Mark authenticator as verified after login
			var verifiedClaim *verification.Claim
			switch n.Authenticator.Type {
			case model.AuthenticatorTypeOOBEmail:
				verifiedClaim = deps.Verification.NewVerifiedClaim(ctx, n.Authenticator.UserID, string(model.ClaimEmail), n.Authenticator.OOBOTP.Email)
			case model.AuthenticatorTypeOOBSMS:
				verifiedClaim = deps.Verification.NewVerifiedClaim(ctx, n.Authenticator.UserID, string(model.ClaimPhoneNumber), n.Authenticator.OOBOTP.Phone)
			}

			if verifiedClaim == nil {
				return nil
			}

			if err := deps.Verification.MarkClaimVerified(ctx, verifiedClaim); err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (*NodeVerifiedAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeVerifiedAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeVerifiedAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{}, nil
}

func (n *NodeVerifiedAuthenticator) GetAMR() []string {
	return n.Authenticator.AMR()
}

var _ VerifiedAuthenticationLockoutMethodGetter = &NodeVerifiedAuthenticator{}

func (n *NodeVerifiedAuthenticator) GetVerifiedAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	if n.Authenticator != nil {
		return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
	}
	return "", false
}
