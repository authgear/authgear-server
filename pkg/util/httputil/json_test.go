package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func TestBindJSONBody(t *testing.T) {
	Convey("BindJSONBody", t, func() {
		Convey("WithBodyMaxSize", func() {
			schema := validation.SchemaBuilder{}.ToSimpleSchema()
			body := strings.NewReader("{}")
			r, _ := http.NewRequest("POST", "", body)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			var m interface{}
			err := BindJSONBody(r, w, schema.Validator(), &m, WithBodyMaxSize(1))
			So(apierrors.AsAPIError(err), ShouldResemble, &apierrors.APIError{
				Kind: apierrors.Kind{
					Name:   apierrors.RequestEntityTooLarge,
					Reason: "JSONTooLarge",
				},
				Message: "request body too large",
				Code:    413,
				Info: map[string]interface{}{
					"limit": int64(1),
				},
			})
		})

		Convey("Validate against schema", func() {
			schema := validation.SchemaBuilder{}.
				Type(validation.TypeNumber).
				ToSimpleSchema()

			body := strings.NewReader(`{}`)
			r, _ := http.NewRequest("POST", "", body)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			var m interface{}
			err := BindJSONBody(r, w, schema.Validator(), &m)
			So(err, ShouldBeError, `invalid request body:
<root>: type
  map[actual:[object] expected:[number]]`)
		})
	})
}
