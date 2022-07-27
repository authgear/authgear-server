package identity

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type LoginIDSpec struct {
	Key   string                `json:"key"`
	Type  config.LoginIDKeyType `json:"type"`
	Value string                `json:"value"`
}
