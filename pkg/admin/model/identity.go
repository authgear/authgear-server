package model

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

type IdentityDefLoginID struct {
	Key   string
	Value string
}

var _ IdentityDef = &IdentityDefLoginID{}
var _ nodes.InputUseIdentityLoginID = &IdentityDefLoginID{}

func (i *IdentityDefLoginID) Type() authn.IdentityType { return authn.IdentityTypeLoginID }
func (i *IdentityDefLoginID) GetLoginIDKey() string    { return i.Key }
func (i *IdentityDefLoginID) GetLoginID() string       { return i.Value }

type IdentityDef interface {
	Type() authn.IdentityType
}

func ParseIdentityDef(data map[string]interface{}) (IdentityDef, error) {
	var key string
	var value map[string]interface{}
	for k, v := range data {
		if v == nil {
			continue
		}

		if value != nil {
			value = nil
			break
		}
		key = k
		value = v.(map[string]interface{})
	}
	if value == nil {
		return nil, apierrors.NewInvalid("exactly 1 field must be present in identity definition")
	}

	switch key {
	case "loginID":
		key := value["key"].(string)
		value := value["value"].(string)
		if key == "" || value == "" {
			return nil, apierrors.NewInvalid("login ID key & value is required")
		}

		return &IdentityDefLoginID{
			Key:   key,
			Value: value,
		}, nil

	default:
		return nil, apierrors.NewInvalid("invalid identity type")
	}
}
