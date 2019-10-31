package validation

import (
	"encoding/json"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"

	. "github.com/smartystreets/goconvey/convey"
)

func TestError(t *testing.T) {
	Convey("validation API error", t, func() {
		err := NewValidationFailed("validation failed", []ErrorCause{
			ErrorCause{
				Kind:    ErrorRequired,
				Pointer: "/username",
				Message: "username is required",
			},
			ErrorCause{
				Kind:    ErrorStringLength,
				Pointer: "/password",
				Message: "password is too short",
				Details: map[string]interface{}{"gte": 8},
			},
		})
		apiErr := skyerr.AsAPIError(err)
		j, _ := json.Marshal(apiErr)
		So(string(j), ShouldEqualJSON, `{
			"name": "Invalid",
			"reason": "ValidationFailed",
			"message": "validation failed",
			"code": 400,
			"info": {
				"causes": [
					{
						"kind": "Required",
						"pointer": "/username",
						"message": "username is required"
					},
					{
						"kind": "StringLength",
						"pointer": "/password",
						"message": "password is too short",
						"details": { "gte": 8 }
					}
				]
			}
		}`)
	})
}
