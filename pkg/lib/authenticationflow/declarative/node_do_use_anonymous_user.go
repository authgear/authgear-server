package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAnonymousUser{})
}

type NodeDoUseAnonymousUser struct {
	Identity      *identity.Info `json:"identity,omitempty"`
	JWT           string         `json:"jwt,omitempty"`
	PromotionCode string         `json:"promotion_code,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAnonymousUser{}
var _ authflow.Milestone = &NodeDoUseAnonymousUser{}
var _ MilestoneDoUseUser = &NodeDoUseAnonymousUser{}
var _ MilestoneDoUseAnonymousUser = &NodeDoUseAnonymousUser{}
var _ authflow.EffectGetter = &NodeDoUseAnonymousUser{}

func NewNodeDoUseAnonymousUser(ctx context.Context, deps *authflow.Dependencies) (*NodeDoUseAnonymousUser, error) {
	loginHintString := authflow.GetLoginHint(ctx)

	loginHint, err := oauth.ParseLoginHint(loginHintString)
	if err != nil {
		return nil, err
	}

	if loginHint.Type != oauth.LoginHintTypeAnonymous {
		return nil, fmt.Errorf("unexpected login hint type: %v", loginHint.Type)
	}

	switch {
	case loginHint.PromotionCode != "":
		promotionCode := loginHint.PromotionCode
		codeObj, err := deps.AnonymousUserPromotionCodeStore.GetPromotionCode(ctx, anonymous.HashPromotionCode(promotionCode))
		if err != nil {
			return nil, err
		}

		promoteUserID := codeObj.UserID
		promoteIdentityID := codeObj.IdentityID

		anonymousIdentity, err := deps.AnonymousIdentities.Get(ctx, promoteUserID, promoteIdentityID)
		if err != nil {
			return nil, err
		}

		if anonymousIdentity.KeyID != "" {
			panic(fmt.Errorf("anonymous user with key must use jwt to trigger promotion flow"))
		}

		return &NodeDoUseAnonymousUser{
			Identity:      anonymousIdentity.ToInfo(),
			PromotionCode: promotionCode,
		}, nil
	case loginHint.JWT != "":
		request, err := deps.AnonymousIdentities.ParseRequestUnverified(loginHint.JWT)
		if err != nil {
			return nil, api.ErrInvalidCredentials
		}

		chal, err := deps.Challenges.Get(ctx, request.Challenge)
		if err != nil || chal.Purpose != challenge.PurposeAnonymousRequest {
			return nil, api.ErrInvalidCredentials
		}

		anonymousIdentity, err := deps.AnonymousIdentities.GetByKeyID(ctx, request.KeyID)
		if err != nil {
			return nil, err
		}

		request, err = deps.AnonymousIdentities.ParseRequest(loginHint.JWT, anonymousIdentity)
		if err != nil {
			dispatchEvent := func() error {
				userID := anonymousIdentity.UserID
				userRef := model.UserRef{
					Meta: model.Meta{
						ID: userID,
					},
				}
				err = deps.Events.DispatchEventOnCommit(ctx, &nonblocking.AuthenticationFailedIdentityEventPayload{
					UserRef:      userRef,
					IdentityType: string(model.IdentityTypeAnonymous),
				})
				if err != nil {
					return err
				}

				return nil
			}
			_ = dispatchEvent()
			return nil, api.ErrInvalidCredentials
		}

		return &NodeDoUseAnonymousUser{
			Identity: anonymousIdentity.ToInfo(),
			JWT:      loginHint.JWT,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected login hint content: %v", loginHintString)
	}
}

func (*NodeDoUseAnonymousUser) Kind() string {
	return "NodeDoUseAnonymousUser"
}

func (*NodeDoUseAnonymousUser) Milestone()                                    {}
func (n *NodeDoUseAnonymousUser) MilestoneDoUseUser() string                  { return n.Identity.UserID }
func (n *NodeDoUseAnonymousUser) MilestoneDoUseAnonymousUser() *identity.Info { return n.Identity }

func (n *NodeDoUseAnonymousUser) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.JWT != "" {
				request, err := deps.AnonymousIdentities.ParseRequestUnverified(n.JWT)
				if err != nil {
					return err
				}

				_, err = deps.Challenges.Consume(ctx, request.Challenge)
				if err != nil {
					return err
				}

				return nil
			}

			return nil
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.PromotionCode != "" {
				codeObj, err := deps.AnonymousUserPromotionCodeStore.GetPromotionCode(ctx, anonymous.HashPromotionCode(n.PromotionCode))
				if err != nil {
					return err
				}

				err = deps.AnonymousUserPromotionCodeStore.DeletePromotionCode(ctx, codeObj)
				if err != nil {
					return err
				}

				return nil
			}

			return nil
		}),
	}, nil
}
