package authenticator

import "github.com/authgear/authgear-server/pkg/api/model"

type Kind = model.AuthenticatorKind

const (
	KindPrimary   Kind = model.AuthenticatorKindPrimary
	KindSecondary Kind = model.AuthenticatorKindSecondary
)
