package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaRenameHandler(t *testing.T) {
	Convey("Test SchemaRenameHandler", t, func() {
		recordStore := record.NewMockStore()
		recordStore.SchemaMap = record.SchemaMap{
			"note": {
				"field1": record.FieldType{Type: record.TypeString},
				"field2": record.FieldType{Type: record.TypeDateTime},
			},
		}

		sh := &SchemaRenameHandler{}
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("rename normal field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "newName"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field2", "type": "datetime"},
								{"name": "newName", "type": "string"}
							]
						}
					}
				}
			}`)
		})

		Convey("rename reserved field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "_id",
				"new_name": "newName"
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

		Convey("rename nonexisting field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "notexist",
				"new_name": "newName"
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

		Convey("rename to existing field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_type": "note",
				"item_name": "field1",
				"new_name": "field2"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 110,
					"message": "column type conflict",
					"name": "ResourceNotFound"
				}
			}`)
		})
	})
}
