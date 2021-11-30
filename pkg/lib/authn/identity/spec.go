package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Spec struct {
	Type   model.IdentityType     `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}
