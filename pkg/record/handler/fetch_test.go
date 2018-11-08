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
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRecordFetchHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	getRecordStore := func() (recordStore *record.MockStore) {
		recordStore = record.NewMockStore()
		recordStore.SchemaMap = record.SchemaMap{
			"note": {
				"content":  record.FieldType{Type: record.TypeString},
				"favorite": record.FieldType{Type: record.TypeBoolean},
				"category": record.FieldType{Type: record.TypeString},
			},
		}

		publicRole := record.FieldUserRole{
			Type: record.PublicFieldUserRoleType,
			Data: "",
		}

		recordStore.Save(&record.Record{
			ID:        record.NewRecordID("note", "note0"),
			OwnerID:   "user0",
			CreatorID: "user0",
			CreatedAt: timeNow(),
			UpdaterID: "user0",
			UpdatedAt: timeNow(),
			Data: map[string]interface{}{
				"content":  "Hello World!",
				"category": "interesting",
			},
		})

		recordStore.SetRecordFieldAccess(record.NewFieldACL(record.FieldACLEntryList{
			{
				RecordType:  "*",
				RecordField: "*",
				UserRole:    publicRole,
				Writable:    true,
				Readable:    true,
			},
			{
				RecordType:  "note",
				RecordField: "category",
				UserRole:    publicRole,
				Writable:    true,
				Readable:    false,
			},
		}))

		return
	}

	Convey("Test FetchHandler", t, func() {
		fh := &FetchHandler{}
		fh.RecordStore = getRecordStore()
		fh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		fh.Logger = logging.LoggerEntry("handler")
		// TODO:
		fh.AssetStore = nil
		fh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(fh, fh.TxContext)

		Convey("decode valid request", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"ids": ["note/1", "message/1"]
				}
			`))
			payload, err := fh.DecodeRequest(req)
			fetchPayload, ok := payload.(FetchRequestPayload)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(fetchPayload, ShouldResemble, FetchRequestPayload{
				RawIDs: []string{"note/1", "message/1"},
				RecordIDs: []record.ID{
					record.ID{Type: "note", Key: "1"},
					record.ID{Type: "message", Key: "1"},
				},
			})
		})

		Convey("decode invalid request", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"ids": ["note1"]
				}
			`))
			payload, err := fh.DecodeRequest(req)
			So(err, ShouldNotBeNil)
			So(payload, ShouldBeNil)

			serr := err.(skyerr.Error)
			So(serr.Code(), ShouldEqual, skyerr.InvalidArgument)
		})

		Convey("should fetch without non-readable fields", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"ids": ["note/note0"]
				}
			`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_recordType": "note",
					"_recordID": "note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})
	})

	Convey("Test FetchHandler with master key", t, func() {
		fh := &FetchHandler{}
		fh.RecordStore = getRecordStore()
		fh.AuthContext = auth.NewMockContextGetterWithMasterkeyDefaultUser()
		fh.Logger = logging.LoggerEntry("handler")
		// TODO:
		fh.AssetStore = nil
		fh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(fh, fh.TxContext)

		Convey("should fetch with all fields with master key", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"ids": ["note/note0"]
				}
			`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_recordType": "note",
					"_recordID": "note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"category": "interesting",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})
	})
}
