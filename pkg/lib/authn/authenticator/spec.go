package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Spec struct {
	UserID    string                   `json:"user_id"`
	Type      model.AuthenticatorType  `json:"type"`
	IsDefault bool                     `json:"is_default"`
	Kind      Kind                     `json:"kind"`
	Claims    map[ClaimKey]interface{} `json:"claims"`
}
