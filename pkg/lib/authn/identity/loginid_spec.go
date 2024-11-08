package identity

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

type LoginIDSpec struct {
	Key   string                     `json:"key"`
	Type  model.LoginIDKeyType       `json:"type"`
	Value stringutil.UserInputString `json:"value"`
}
