package loginid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type MockAddLoginIDInteractionFlow struct {
	ErrorMap map[string]error
}

func (m MockAddLoginIDInteractionFlow) AddLoginID(
	loginIDKey string, loginID string, session auth.AuthSession,
) error {
	return m.ErrorMap[loginID]
}

func TestAddLoginIDHandler(t *testing.T) {
	Convey("Test AddLoginIDHandler", t, func() {
		h := &AddLoginIDHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			AddLoginIDRequestSchema,
		)
		h.Validator = validator
		h.TxContext = db.NewMockTxContext()
		mockFlow := MockAddLoginIDInteractionFlow{}
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
				"login_ids": [
					{ "key": "username", "value": "invalid" }
				]
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
								"pointer": "/login_ids/0/value"
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
