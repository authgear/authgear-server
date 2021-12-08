package customattrs

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

// T is the representation form of custom attributes.
// The keys are derived from pointers.
type T map[string]interface{}

func (t T) ReadWithAccessControl(accessControl accesscontrol.T, role accesscontrol.Role) T {
	out := T{}
	for key, val := range t {
		subject := accesscontrol.Subject(jsonpointer.T{key}.String())
		level := accessControl.GetLevel(subject, role, config.AccessControlLevelReadwrite)
		if level >= config.AccessControlLevelReadonly {
			out[key] = val
		}
	}
	return out
}

func (t T) ToMap() map[string]interface{} {
	return map[string]interface{}(t)
}
