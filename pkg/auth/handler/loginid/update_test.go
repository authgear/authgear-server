package loginid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockUpdateLoginIDInteractionFlow struct {
	ErrorMap map[string]error
}

func (m MockUpdateLoginIDInteractionFlow) UpdateLoginID(
	oldLoginID loginid.LoginID, newLoginID loginid.LoginID, session auth.AuthSession,
) (*interactionflows.AuthResult, error) {
	return nil, m.ErrorMap[newLoginID.Value]
}

func TestUpdateLoginIDHandler(t *testing.T) {
	Convey("Test UpdateLoginIDHandler", t, func() {
		h := &UpdateLoginIDHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			UpdateLoginIDRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		mockFlow := MockUpdateLoginIDInteractionFlow{}
		mockFlow.ErrorMap = map[string]error{
			"invalid": validation.NewValidationFailed("invalid login ID", []validation.ErrorCause{{
				Kind:    validation.ErrorStringFormat,
				Pointer: "/0/value",
				Message: "invalid login ID format",
				Details: map[string]interface{}{"format": "email"},
			}}),
		}
		h.Interactions = mockFlow

		Convey("should correct error pointer", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"old_login_id": {
					"key": "email", "value": "user@example.com"
				},
				"new_login_id": {
					"key": "email", "value": "invalid"
				}
			}`))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 400,
					"info": {
						"causes": [
							{
								"details": {
									"format": "email"
								},
								"kind": "StringFormat",
								"message": "invalid login ID format",
								"pointer": "/new_login_id/value"
							}
						]
					},
					"message": "invalid login ID",
					"name": "Invalid",
					"reason": "ValidationFailed"
				}
			}`)
		})
	})
}
