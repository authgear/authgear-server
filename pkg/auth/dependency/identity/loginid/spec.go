package loginid

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
)

type Spec struct {
	Key   string
	Type  config.LoginIDKeyType
	Value string
}
