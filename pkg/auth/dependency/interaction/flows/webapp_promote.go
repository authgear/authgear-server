package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

const (
	// WebAppExtraStatePromotion is a extra state indicating the interaction
	// is for anonymous user promotion.
	WebAppExtraStateAnonymousUserPromotion string = "https://auth.skygear.io/claims/web_app/anonymous_user_promotion"
)

func (f *WebAppFlow) afterAnonymousUserPromotion(attrs *authn.Attrs) (*WebAppResult, error) {
	// Remove anonymous identity
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeAnonymous,
			Claims: map[string]interface{}{},
		},
	}, "", attrs.UserID)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}

	result, err := f.afterPrimaryAuthentication(i)
	if err != nil {
		return nil, err
	}

	// NOTE: existing anonymous sessions are not deleted, in case of commit
	// failure may cause lost users.

	return result, nil
}
