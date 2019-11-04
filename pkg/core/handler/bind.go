package handler

import (
	"io"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type BodyDefaulter interface {
	SetDefaultValue()
}

type BodyValidator interface {
	Validate() []validation.ErrorCause
}

func BindJSONBody(r *http.Request, w http.ResponseWriter, v *validation.Validator, schemaID string, payload interface{}) error {
	const message = "invalid request body"
	return ParseJSONBody(r, w, func(reader io.Reader, value interface{}) error {
		err := v.WithMessage(message).ParseReader(schemaID, reader, value)
		if err != nil {
			if !skyerr.IsKind(err, validation.ValidationFailed) {
				return skyerr.NewBadRequest("invalid request body")
			}
			return err
		}
		if value, ok := value.(BodyDefaulter); ok {
			value.SetDefaultValue()
		}
		if value, ok := value.(BodyValidator); ok {
			causes := value.Validate()
			if len(causes) > 0 {
				return validation.NewValidationFailed(message, causes)
			}
		}
		return nil
	}, payload)
}
