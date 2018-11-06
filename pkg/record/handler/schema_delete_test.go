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

func TestSchemaDeleteHandler(t *testing.T) {
	Convey("Test SchemaDeleteHandler", t, func() {
		// fixture
		recordStore := record.NewMockStore()
		recordStore.SchemaMap = record.SchemaMap{
			"note": {
				"field1": record.FieldType{Type: record.TypeString},
				"field2": record.FieldType{Type: record.TypeDateTime},
			},
		}

		sh := &SchemaDeleteHandler{}
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("delete normal field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "field1"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field2", "type": "datetime"}
							]
						}
					}
				}
			}`)
		})

		Convey("delete reserved field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "_id"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "attempts to change reserved key",
					"info": {
						"arguments": [
							"item_name"
						]
					},
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("delete nonexisting field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "notexist"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "column notexist does not exist",
					"name": "ResourceNotFound"
				}
			}`)
		})
	})
}
