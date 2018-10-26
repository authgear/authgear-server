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

func TestRecordSaveHandler(t *testing.T) {
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

		recordStore.SetRecordAccess("report", record.NewACL([]record.ACLEntry{
			record.NewACLEntryRole("admin", record.CreateLevel),
		}))

		recordStore.Save(&record.Record{
			ID: record.NewRecordID("note", "readonly"),
			ACL: record.ACL{
				record.NewACLEntryDirect("faseng.cat.id", record.ReadLevel),
			},
		})
		return
	}

	Convey("RecordSaveHandler", t, func() {
		sh := &SaveHandler{}
		sh.RecordStore = getRecordStore()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		sh.Logger = logging.LoggerEntry("handler")
		// TODO:
		sh.AssetStore = nil
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("Saves multiple records", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_id": "type1/id1",
					"k1": "v1",
					"k2": "v2"
				}, {
					"_recordType": "type2",
					"_recordID": "id2",
					"k3": "v3",
					"k4": "v4"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type1/id1",
					"_recordType": "type1",
					"_recordID": "id1",
					"_type": "record",
					"_access": null,
					"k1": "v1",
					"k2": "v2",
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}, {
					"_id": "type2/id2",
					"_recordType": "type2",
					"_recordID": "id2",
					"_type": "record",
					"_access": null,
					"k3": "v3",
					"k4": "v4",
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}]
			}`)
		})

		Convey("Should not be able to create record when no permission", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "report",
					"_recordID": "id1",
					"k1": "v1",
					"k2": "v2"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{
						"_id": "report/id1",
						"_recordType": "report",
						"_recordID": "id1",
						"_type": "error",
						"code": 102,
						"message": "no permission to create",
						"name": "PermissionDenied"
					}
				]
			}`)
		})

		Convey("Removes reserved keys on save", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "type1",
					"_recordID": "id1",
					"floatkey": 1,
					"_reserved_key": "reserved_value"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type1/id1",
					"_recordType": "type1",
					"_recordID": "id1",
					"_type": "record",
					"_access":null,
					"floatkey": 1,
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}]
			}`)
		})

		Convey("Returns error if _id is missing or malformated", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
				}, {
					"_id": "invalidkey"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "missing _recordType, expecting string"
				},{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "invalid record id"
				}]
			}`)
		})

		Convey("Returns error if _recordType/recordID is missing or malformated", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
				}, {
					"_recordType": "note"
				}, {
					"_recordID": "1234"
				}, {
					"_recordType": "note",
					"_recordID": ""
				}, {
					"_recordType": "note",
					"_recordID": 1234
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
				{
					"_type": "error",
					"code": 108,
					"message": "missing _recordType, expecting string",
					"name": "InvalidArgument"
				},
				{
					"_type": "error",
					"code": 108,
					"message": "missing _recordID, expecting string",
					"name": "InvalidArgument"
				},
				{
					"_type": "error",
					"code": 108,
					"message": "missing _recordType, expecting string",
					"name": "InvalidArgument"
				},
				{
					"_type": "error",
					"code": 108,
					"message": "missing _recordID, expecting string",
					"name": "InvalidArgument"
				},
				{
					"_type": "error",
					"code": 108,
					"message": "missing _recordType, expecting string",
					"name": "InvalidArgument"
				}
				]
			}`)
		})

		Convey("Permission denied on saving a read only record", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_id": "note/readonly",
					"_recordType": "note",
					"_recordID": "readonly",
					"content": "hello"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/readonly",
					"_recordType": "note",
					"_recordID": "readonly",
					"_type": "error",
					"code": 102,
					"message": "no permission to perform operation",
					"name": "PermissionDenied"
				}]
			}`)
		})
		Convey("REGRESSION #140: Save record correctly when record._access is null", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "type",
					"_recordID": "id",
					"_access": null
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "record",
					"_id": "type/id",
					"_recordType": "type",
					"_recordID": "id",
					"_access": null,
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}]
			}`)
		})

		Convey("REGRESSION #333: Save record with empty key be ignored as start with _", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "type",
					"_recordID": "id"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "record",
					"_id": "type/id",
					"_recordType": "type",
					"_recordID": "id",
					"_access": null,
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}]
			}`)
		})
	})

	Convey("RecordSaveHandler with admin", t, func() {
		sh := &SaveHandler{}
		sh.RecordStore = getRecordStore()
		sh.AuthContext = auth.NewMockContextGetterWithAdminUser()
		sh.Logger = logging.LoggerEntry("handler")
		// TODO:
		sh.AssetStore = nil
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("Should be able to create record when have permission", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "report",
					"_recordID": "id1",
					"k1": "v1",
					"k2": "v2"
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{
						"_id": "report/id1",
						"_recordType": "report",
						"_recordID": "id1",
						"_type": "record",
						"_access": null,
						"_created_by":"chima.cat.id",
						"_updated_by":"chima.cat.id",
						"_ownerID": "chima.cat.id",
						"k1":"v1",
						"k2":"v2"
					}
				]
			}`)
		})
	})
}

func TestRecordSaveHandlerWithFieldACL(t *testing.T) {
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

		recordStore.Save(&record.Record{
			ID:        record.NewRecordID("note", "note0"),
			OwnerID:   "faseng.cat.id",
			CreatorID: "faseng.cat.id",
			CreatedAt: timeNow(),
			UpdaterID: "faseng.cat.id",
			UpdatedAt: timeNow(),
			Data: map[string]interface{}{
				"content":  "Hello World!",
				"favorite": true,
				"category": "interesting",
			},
		})

		publicRole := record.FieldUserRole{
			Type: record.PublicFieldUserRoleType,
			Data: "",
		}

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
				RecordField: "content",
				UserRole:    publicRole,
				Writable:    false,
				Readable:    true,
			},
			{
				RecordType:  "note",
				RecordField: "category",
				UserRole:    publicRole,
				Writable:    true,
				Readable:    false,
			},
			{
				RecordType:  "note",
				RecordField: "*",
				UserRole:    publicRole,
				Writable:    true,
				Readable:    true,
			},
		}))
		return
	}

	Convey("RecordSaveHandler with Field ACL", t, func() {
		sh := &SaveHandler{}
		recordStore := getRecordStore()
		sh.RecordStore = recordStore
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		sh.Logger = logging.LoggerEntry("handler")
		// TODO:
		sh.AssetStore = nil
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should not save to read only field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"records": [{
					"_recordType": "note",
					"_recordID": "note0",
					"content": "Bye World!",
					"favorite": false
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
					"_recordType": "note",
					"_recordID": "note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"favorite": false,
					"_created_by":"faseng.cat.id",
					"_updated_by":"faseng.cat.id",
					"_ownerID": "faseng.cat.id"
				}]
			}`)
			So(recordStore.Map["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(recordStore.Map["note/note0"].Data["favorite"], ShouldEqual, false)
		})

		Convey("should fail request if atomic save to read only field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"atomic": true,
				"records": [{
					"_recordType": "note",
					"_recordID": "note0",
					"content": "Bye World!",
					"favorite": false
				}]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code":115,
					"info": {
						"note/note0": {
							"code":123,
							"info": { "arguments":["content"] },
							"message":"Unable to save to some record fields because of Field ACL denied update.",
							"name":"DeniedArgument"
						}
					},
					"message":"Atomic Operation rolled back due to one or more errors",
					"name":"AtomicOperationFailure"
				}
			}`)
			So(recordStore.Map["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(recordStore.Map["note/note0"].Data["favorite"], ShouldEqual, true)
		})

		Convey("should not fail request if read only field did not change", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
					"atomic": true,
					"records": [{
						"_recordType": "note",
						"_recordID": "note0",
						"content": "Hello World!",
						"favorite": false
					}]
				}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
					"result": [{
						"_id": "note/note0",
						"_recordType": "note",
						"_recordID": "note0",
						"_type": "record",
						"_access": null,
						"content": "Hello World!",
						"favorite": false,
						"_created_by":"faseng.cat.id",
						"_updated_by":"faseng.cat.id",
						"_ownerID": "faseng.cat.id"
					}]
				}`)
			So(recordStore.Map["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(recordStore.Map["note/note0"].Data["favorite"], ShouldEqual, false)
		})

		Convey("should save a new record", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
					"atomic": true,
					"records": [{
						"_recordType": "note",
						"_recordID": "new-note",
						"category": "nice",
						"favorite": true
					}]
				}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
					"result": [{
						"_id": "note/new-note",
						"_recordType": "note",
						"_recordID": "new-note",
						"_type": "record",
						"_access": null,
						"favorite": true,
						"_created_by":"faseng.cat.id",
						"_updated_by":"faseng.cat.id",
						"_ownerID": "faseng.cat.id"
					}]
				}`)
			So(recordStore.Map["note/new-note"].Data["category"], ShouldEqual, "nice")
			So(recordStore.Map["note/new-note"].Data["favorite"], ShouldEqual, true)
		})
	})

	Convey("RecordSaveHandler with Field ACL using master key", t, func() {
		sh := &SaveHandler{}
		recordStore := getRecordStore()
		sh.RecordStore = recordStore
		sh.AuthContext = auth.NewMockContextGetterWithMasterkeyDefaultUser()
		sh.Logger = logging.LoggerEntry("handler")
		// TODO:
		sh.AssetStore = nil
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should save to read only field with master key", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
					"records": [{
						"_recordType": "note",
						"_recordID": "note0",
						"content": "Bye World!",
						"favorite": false
					}]
				}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
					"result": [{
						"_id": "note/note0",
						"_recordType": "note",
						"_recordID": "note0",
						"_type": "record",
						"_access": null,
						"content": "Bye World!",
						"favorite": false,
						"category": "interesting",
						"_created_by":"faseng.cat.id",
						"_updated_by":"faseng.cat.id",
						"_ownerID": "faseng.cat.id"
					}]
				}`)
			So(recordStore.Map["note/note0"].Data["content"], ShouldEqual, "Bye World!")
			So(recordStore.Map["note/note0"].Data["favorite"], ShouldEqual, false)
		})
	})
}
