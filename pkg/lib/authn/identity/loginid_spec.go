package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type LoginIDSpec struct {
	Key   string               `json:"key"`
	Type  model.LoginIDKeyType `json:"type"`
	Value string               `json:"value"`
}
