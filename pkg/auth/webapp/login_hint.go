package webapp

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AnonymousTokenInput struct {
	JWT           string
	PromotionCode string
}

func (i *AnonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

func (i *AnonymousTokenInput) SignUpAnonymousUserWithoutKey() bool { return false }

func (i *AnonymousTokenInput) GetPromotionCode() string {
	return i.PromotionCode
}

var _ nodes.InputUseIdentityAnonymous = &AnonymousTokenInput{}

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type AnonymousUserPromotionService struct {
	Anonymous AnonymousIdentityProvider
	Clock     clock.Clock
}

func (r *AnonymousUserPromotionService) ConvertLoginHintToInput(loginHintString string) (*AnonymousTokenInput, error) {
	loginHint, err := oauth.ParseLoginHint(loginHintString)
	if err != nil {
		return nil, err
	}

	if loginHint.Type != oauth.LoginHintTypeAnonymous {
		return nil, fmt.Errorf("unexpected login hint type: %v", loginHint.Type)
	}

	promotionCode := loginHint.PromotionCode
	if promotionCode != "" {
		return &AnonymousTokenInput{
			PromotionCode: promotionCode,
		}, nil
	}

	// jwt flow
	jwt := loginHint.JWT
	request, err := r.Anonymous.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, err
	}

	switch request.Action {
	case anonymous.RequestActionPromote:
		return &AnonymousTokenInput{JWT: jwt}, nil
	case anonymous.RequestActionAuth:
		// TODO(webapp): support anonymous auth
		panic("webapp: anonymous auth through web app is not supported")
	default:
		return nil, errors.New("unknown anonymous request action")
	}
}
