package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRecordDeleteHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	getRecordStore := func() (recordStore *record.MockStore) {
		rs := record.NewMockStore()
		rs.Map = map[string]record.Record{
			"note/0": record.Record{
				ID: record.NewRecordID("note", "0"),
				ACL: record.ACL{
					record.NewACLEntryDirect("faseng.cat.id", record.WriteLevel),
				},
			},
			"note/1": record.Record{
				ID: record.NewRecordID("note", "1"),
				ACL: record.ACL{
					record.NewACLEntryDirect("faseng.cat.id", record.WriteLevel),
				},
			},
			"note/readonly": record.Record{
				ID: record.NewRecordID("note", "readonly"),
				ACL: record.ACL{
					record.NewACLEntryDirect("faseng.cat.id", record.ReadLevel),
				},
			},
			"user/0": record.Record{
				ID: record.NewRecordID("user", "0"),
			},
		}
		return rs
	}

	Convey("Test DeleteHandler", t, func() {
		dh := &DeleteHandler{}
		dh.RecordStore = getRecordStore()
		dh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		dh.Logger = logging.LoggerEntry("handler")
		dh.TxContext = db.NewMockTxContext()

		Convey("deletes existing records", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"records": [
						{ "_recordType": "note", "_recordID": "0" },
						{ "_recordType": "note", "_recordID": "1" }
					]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(dh, dh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{"_id": "note/0","_recordType": "note","_recordID": "0", "_type": "record"},
					{"_id": "note/1","_recordType": "note","_recordID": "1", "_type": "record"}
				]
			}`)
		})

		Convey("DEPRECATED: deletes existing records with deprecated IDs", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"ids": ["note/0", "note/1"]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(dh, dh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{"_id": "note/0","_recordType": "note","_recordID": "0", "_type": "record"},
					{"_id": "note/1","_recordType": "note","_recordID": "1", "_type": "record"}
				]
			}`)
		})

		Convey("returns error when record doesn't exist", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"records": [
						{ "_recordType": "note", "_recordID": "0" },
						{ "_recordType": "note", "_recordID": "notexistid" }
					]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(dh, dh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{"_id": "note/0","_recordType": "note","_recordID": "0", "_type": "record"},
					{"_id": "note/notexistid","_recordType": "note","_recordID": "notexistid", "_type": "error", "code": 110, "message": "record not found", "name": "ResourceNotFound"}
				]
			}`)
		})

		Convey("cannot delete user record", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"records": [
						{ "_recordType": "user", "_recordID": "0" }
					]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(dh, dh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{"_id":"user/0","_recordType":"user","_recordID":"0","_type":"error","code":102,"message":"cannot delete user record","name":"PermissionDenied"}
				]
			}`)
		})

		Convey("permission denied on delete a readonly record", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"records": [
						{ "_recordType": "note", "_recordID": "readonly" }
					]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(dh, dh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/readonly",
					"_recordType": "note",
					"_recordID": "readonly",
					"_type": "error",
					"code":102,
					"message": "no permission to perform operation",
					"name": "PermissionDenied"
				}]
			}`)
		})
	})
}
