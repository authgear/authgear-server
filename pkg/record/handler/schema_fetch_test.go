package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaFetchHandler(t *testing.T) {
	Convey("Test SchemaFetchHandler", t, func() {
		// fixture
		recordStore := record.NewMockStore()
		recordStore.SchemaMap = record.SchemaMap{
			"note": {
				"field1": record.FieldType{Type: record.TypeString},
				"field2": record.FieldType{Type: record.TypeDateTime},
			},
			"user": {},
		}

		sh := &SchemaCreateHandler{}
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("fetch schemas", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field1", "type": "string"},
								{"name": "field2", "type": "datetime"}
							]
						},
						"user": {
							"fields": []
						}
					}
				}
			}`)
		})
	})
}
