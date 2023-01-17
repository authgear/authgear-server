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
	JWT               string
	PromoteUserID     string
	PromoteIdentityID string
}

func (i *AnonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

func (i *AnonymousTokenInput) SignUpAnonymousUserWithoutKey() bool { return false }

func (i *AnonymousTokenInput) GetPromoteUserAndIdentityID() (string, string) {
	return i.PromoteUserID, i.PromoteIdentityID
}

var _ nodes.InputUseIdentityAnonymous = &AnonymousTokenInput{}

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type AnonymousPromotionCodeStore interface {
	GetPromotionCode(codeHash string) (*anonymous.PromotionCode, error)
	DeletePromotionCode(code *anonymous.PromotionCode) error
}

type AnonymousUserPromotionService struct {
	Anonymous               AnonymousIdentityProvider
	AnonymousPromotionCodes AnonymousPromotionCodeStore
	Clock                   clock.Clock
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
		// promotion code flow
		userID, identityID, err := r.resolvePromotionCode(promotionCode)
		if err != nil {
			return nil, err
		}
		return &AnonymousTokenInput{
			PromoteUserID:     userID,
			PromoteIdentityID: identityID,
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

func (r *AnonymousUserPromotionService) resolvePromotionCode(code string) (userID string, identityID string, err error) {
	codeObj, err := r.AnonymousPromotionCodes.GetPromotionCode(anonymous.HashPromotionCode(code))
	if err != nil {
		return
	}

	// FIXME: We need pass the promotion code to the interaction and let the interaction to consume the code.
	// err = r.AnonymousPromotionCodes.DeletePromotionCode(codeObj)
	// if err != nil {
	// 	return
	// }
	userID = codeObj.UserID
	identityID = codeObj.IdentityID
	return
}
