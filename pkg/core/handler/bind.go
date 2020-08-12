package handler

import (
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
	"github.com/authgear/authgear-server/pkg/validation"
)

type BodyDefaulter interface {
	SetDefaults()
}

func BindJSONBody(r *http.Request, w http.ResponseWriter, v *validation.SchemaValidator, payload interface{}) error {
	const errorMessage = "invalid request body"
	return ParseJSONBody(r, w, func(reader io.Reader, value interface{}) error {
		err := v.ParseWithMessage(reader, errorMessage, value)
		if err != nil {
			if !apierrors.IsKind(err, validation.ValidationFailed) {
				return apierrors.NewBadRequest(errorMessage)
			}
			return err
		}

		if value, ok := value.(BodyDefaulter); ok {
			value.SetDefaults()
		}
		return validation.ValidateValueWithMessage(value, errorMessage)
	}, payload)
}
