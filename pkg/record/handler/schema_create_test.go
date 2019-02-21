package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaCreateHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test SchemaCreateHandler", t, func() {
		// fixture
		recordStore := record.NewMockStore()
		recordStore.SchemaMap = record.SchemaMap{
			"note": {
				"field1": record.FieldType{Type: record.TypeString},
				"field2": record.FieldType{Type: record.TypeDateTime},
			},
		}

		sh := &SchemaCreateHandler{}
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("create normal fields", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field3", "type": "string"},
							{"name": "field4", "type": "number"}
						]
					}
				}
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"record_types": {
						"note": {
							"fields": [
								{"name": "field1", "type": "string"},
								{"name": "field2", "type": "datetime"},
								{"name": "field3", "type": "string"},
								{"name": "field4", "type": "number"}
							]
						}
					}
				}
			}`)
		})

		Convey("create reserved field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "_field3", "type": "string"}
						]
					}
				}
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"message": "attempts to create reserved field",
					"info": {
						"arguments": [
							"_field3"
						]
					},
					"name": "InvalidArgument"
				}
			}`)
		})

		Convey("create existing field with conflict", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "integer"}
						]
					}
				}
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 114,
					"message": "Wrong type",
					"name": "IncompatibleSchema"
				}
			}`)
		})

		Convey("create existing field without conflict", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"record_types": {
					"note": {
						"fields": [
							{"name": "field1", "type": "string"}
						]
					}
				}
			}`))
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
						}
					}
				}
			}`)
		})
	})
}
