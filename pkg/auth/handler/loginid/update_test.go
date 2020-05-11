package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
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
	SkipConvey("Test UpdateLoginIDHandler", t, func() {
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

	})
}
