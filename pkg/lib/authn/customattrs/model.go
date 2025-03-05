package customattrs

import (
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
)

// T is the representation form of custom attributes.
// The keys are derived from pointers.
type T map[string]interface{}

func (t T) Clone() T {
	out := make(T)
	for k, v := range t {
		out[k] = v
	}
	return out
}

func (t T) Update(accessControl accesscontrol.T, role accesscontrol.Role, pointerStrs []string, incoming T) (T, error) {
	out := t.Clone()
	for _, ptrStr := range pointerStrs {
		ptr, err := jsonpointer.Parse(ptrStr)
		if err != nil {
			return nil, err
		}

		subject := accesscontrol.Subject(ptrStr)
		level := accessControl.GetLevel(subject, role, config.AccessControlLevelReadwrite)
		allowed := level >= config.AccessControlLevelReadwrite

		if val, err := ptr.Traverse(incoming); err == nil {
			if !allowed {
				return nil, AccessControlViolated.NewWithInfo(
					fmt.Sprintf("%v being updated by %v with level %v", subject, role, level),
					apierrors.Details{
						"subject": subject,
						"role":    role,
						"level":   level,
					},
				)
			}

			err = jsonpointerutil.AssignToJSONObject(ptr, out, val)
			if err != nil {
				return nil, err
			}
		} else {
			if !allowed {
				return nil, AccessControlViolated.NewWithInfo(
					fmt.Sprintf("%v being deleted by %v with level %v", subject, role, level),
					apierrors.Details{
						"subject": subject,
						"role":    role,
						"level":   level,
					},
				)
			}

			err = jsonpointerutil.RemoveFromJSONObject(ptr, out)
			if err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

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
