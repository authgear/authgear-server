package gear

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/template"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTemplatesHandler(t *testing.T) {
	Convey("TemplatesHandler", t, func() {
		Convey("should return information on gear templates", func() {
			engine := template.NewEngine(template.NewEngineOptions{})
			engine.Register(template.Spec{
				Type:    "email.txt",
				Default: "This is a test email.",
			})
			engine.Register(template.Spec{
				Type:    "email.html",
				Default: "",
				IsHTML:  true,
			})

			var handler = TemplatesHandler{
				TemplateEngine: engine,
			}

			r, _ := http.NewRequest("GET", "", nil)
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, r)

			So(rw.Body.Bytes(), ShouldEqualJSON, `[
				{
					"type": "email.html",
					"is_keyed": false,
					"is_html": true
				},
				{
					"type": "email.txt",
					"default": "This is a test email.",
					"is_keyed": false,
					"is_html": false
				}
			]`)
		})
	})
}
