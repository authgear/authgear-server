// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook/hooktest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

var ZeroTime time.Time

func TestRecordDeleteHandler(t *testing.T) {
	Convey("RecordDeleteHandler", t, func() {
		note0 := skydb.Record{
			ID:         skydb.NewRecordID("note", "0"),
			DatabaseID: "",
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.WriteLevel),
			},
		}
		note1 := skydb.Record{
			ID:         skydb.NewRecordID("note", "1"),
			DatabaseID: "",
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.WriteLevel),
			},
		}
		noteReadonly := skydb.Record{
			ID:         skydb.NewRecordID("note", "readonly"),
			DatabaseID: "",
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.ReadLevel),
			},
		}
		user := skydb.Record{
			ID:         skydb.NewRecordID("user", "0"),
			DatabaseID: "",
		}

		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		So(db.Save(&note0), ShouldBeNil)
		So(db.Save(&note1), ShouldBeNil)
		So(db.Save(&noteReadonly), ShouldBeNil)
		So(db.Save(&user), ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(&RecordDeleteHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
			p.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("deletes existing records", func() {
			resp := router.POST(`{
	"ids": ["note/0", "note/1"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"_id": "note/0", "_type": "record"},
		{"_id": "note/1", "_type": "record"}
	]
}`)
		})

		Convey("returns error when record doesn't exist", func() {
			resp := router.POST(`{
	"ids": ["note/0", "note/notexistid"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"_id": "note/0", "_type": "record"},
		{"_id": "note/notexistid", "_type": "error", "code": 110, "message": "record not found", "name": "ResourceNotFound"}
	]
}`)

		})

		Convey("cannot delete user record", func() {
			resp := router.POST(`{
	"ids": ["user/0"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"_id":"user/0","_type":"error","code":102,"message":"cannot delete user record","name":"PermissionDenied"}
	]
}`)

		})

		Convey("permission denied on delete a readonly record", func() {
			resp := router.POST(`{
				"ids": ["note/readonly"]
			}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/readonly",
					"_type": "error",
					"code":102,
					"message": "no permission to perform operation",
					"name": "PermissionDenied"
				}]
			}`)
		})

	})
}

// trueStore is a TokenStore that always noop on Put and assign itself on Get
type trueStore authtoken.Token

func (store *trueStore) Get(id string, token *authtoken.Token) error {
	*token = authtoken.Token(*store)
	return nil
}

func (store *trueStore) Put(token *authtoken.Token) error {
	return nil
}

// errStore is a TokenStore that always noop and returns itself as error
// on both Get and Put
type errStore authtoken.NotFoundError

func (store *errStore) Get(id string, token *authtoken.Token) error {
	return (*authtoken.NotFoundError)(store)
}

func (store *errStore) Put(token *authtoken.Token) error {
	return (*authtoken.NotFoundError)(store)
}

func TestRecordSaveHandler(t *testing.T) {
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordSaveHandler", t, func() {
		db := skydbtest.NewMapDB()
		conn := skydbtest.NewMapConn()
		conn.SetRecordAccess("report", skydb.NewRecordACL([]skydb.RecordACLEntry{
			skydb.NewRecordACLEntryRole("admin", skydb.CreateLevel),
		}))

		db.Save(&skydb.Record{
			ID: skydb.NewRecordID("note", "readonly"),
			ACL: skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.ReadLevel),
			},
		})

		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
			payload.DBConn = conn
			payload.Database = db
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("Saves multiple records", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "type1/id1",
					"k1": "v1",
					"k2": "v2"
				}, {
					"_id": "type2/id2",
					"k3": "v3",
					"k4": "v4"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type1/id1",
					"_type": "record",
					"_access": null,
					"k1": "v1",
					"k2": "v2",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}, {
					"_id": "type2/id2",
					"_type": "record",
					"_access": null,
					"k3": "v3",
					"k4": "v4",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})

		Convey("Should not be able to create record when no permission", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "report/id1",
					"k1": "v1",
					"k2": "v2"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [
					{
						"_id": "report/id1",
						"_type": "error",
						"code": 102,
						"message": "no permission to create",
						"name": "PermissionDenied"
					}
				]
			}`)
		})

		Convey("Removes reserved keys on save", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "type1/id1",
					"floatkey": 1,
					"_reserved_key": "reserved_value"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type1/id1",
					"_type": "record",
					"_access":null,
					"floatkey": 1,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})

		Convey("Returns error if _id is missing or malformated", func() {
			resp := r.POST(`{
				"records": [{
				}, {
					"_id": "invalidkey"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "missing required fields",
					"info": {"arguments":["id"]}
				},{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "record: \"_id\" should be of format '{type}/{id}', got \"invalidkey\"",
					"info": {"arguments":["id"]}
			}]}`)
		})

		Convey("Permission denied on saving a read only record", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "note/readonly",
					"content": "hello"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/readonly",
					"_type": "error",
					"code": 102,
					"message": "no permission to perform operation",
					"name": "PermissionDenied"
				}]
			}`)
		})
		Convey("REGRESSION #119: Returns record invalid error if _id is missing or malformated", func() {
			resp := r.POST(`{
				"records": [{
				}, {
					"_id": "invalidkey"
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "missing required fields",
					"info": {"arguments":["id"]}
				},{
					"_type": "error",
					"name": "InvalidArgument",
					"code": 108,
					"message": "record: \"_id\" should be of format '{type}/{id}', got \"invalidkey\"",
					"info": {"arguments":["id"]}
			}]}`)
		})

		Convey("REGRESSION #140: Save record correctly when record._access is null", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "type/id",
					"_access": null
				}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "record",
					"_id": "type/id",
					"_access": null,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})

		Convey("REGRESSION #333: Save record with empty key be ignored as start with _", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "type/id",
					"": ""
				}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_type": "record",
					"_id": "type/id",
					"_access": null,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})
	})

	Convey("RecordSaveHandler with Field ACL", t, func() {
		mapDB := skydbtest.NewMapDB()
		mapDB.RecordSchemaMap = skydbtest.RecordSchemaMap{
			"note": {
				"content":  skydb.FieldType{Type: skydb.TypeString},
				"favorite": skydb.FieldType{Type: skydb.TypeBoolean},
				"category": skydb.FieldType{Type: skydb.TypeString},
			},
		}
		db := skydbtest.NewMockTxDatabase(mapDB)
		conn := skydbtest.NewMapConn()
		publicRole := skydb.FieldUserRole{skydb.PublicFieldUserRoleType, ""}

		db.Save(&skydb.Record{
			ID:        skydb.NewRecordID("note", "note0"),
			OwnerID:   "user0",
			CreatorID: "user0",
			CreatedAt: timeNow(),
			UpdaterID: "user0",
			UpdatedAt: timeNow(),
			Data: map[string]interface{}{
				"content":  "Hello World!",
				"favorite": true,
				"category": "interesting",
			},
		})

		conn.SetRecordFieldAccess(skydb.NewFieldACL(skydb.FieldACLEntryList{
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

		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
			payload.DBConn = conn
			payload.Database = db
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
			if apiKey, ok := payload.Data["api_key"]; ok && apiKey == "master" {
				payload.AccessKey = router.MasterAccessKey
			}
		})

		Convey("should not save to read only field", func() {
			resp := r.POST(`{
				"records": [{
					"_id": "note/note0",
					"content": "Bye World!",
					"favorite": false
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"favorite": false,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
			So(mapDB.RecordMap["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(mapDB.RecordMap["note/note0"].Data["favorite"], ShouldEqual, false)
		})

		Convey("should save to read only field with master key", func() {
			resp := r.POST(`{
				"api_key": "master",
				"records": [{
					"_id": "note/note0",
					"content": "Bye World!",
					"favorite": false
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
					"_type": "record",
					"_access": null,
					"content": "Bye World!",
					"favorite": false,
					"category": "interesting",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
			So(mapDB.RecordMap["note/note0"].Data["content"], ShouldEqual, "Bye World!")
			So(mapDB.RecordMap["note/note0"].Data["favorite"], ShouldEqual, false)
		})

		Convey("should fail request if atomic save to read only field", func() {
			resp := r.POST(`{
				"atomic": true,
				"records": [{
					"_id": "note/note0",
					"content": "Bye World!",
					"favorite": false
				}]
			}`)
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
			So(mapDB.RecordMap["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(mapDB.RecordMap["note/note0"].Data["favorite"], ShouldEqual, true)
		})

		Convey("should not fail request if read only field did not change", func() {
			resp := r.POST(`{
				"atomic": true,
				"records": [{
					"_id": "note/note0",
					"content": "Hello World!",
					"favorite": false
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"favorite": false,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
			So(mapDB.RecordMap["note/note0"].Data["content"], ShouldEqual, "Hello World!")
			So(mapDB.RecordMap["note/note0"].Data["favorite"], ShouldEqual, false)
		})
	})
}

func TestRecordSaveDataType(t *testing.T) {
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordSaveHandler", t, func() {
		db := skydbtest.NewMapDB()
		conn := skydbtest.NewMapConn()
		conn.AssetMap = map[string]skydb.Asset{
			"asset-name": skydb.Asset{
				Name:        "asset-name",
				ContentType: "plain/text",
			},
		}

		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
			p.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("Parses date", func() {
			resp := r.POST(`{
	"records": [{
		"_id": "type1/id1",
		"date_value": {"$type": "date", "$date": "2015-04-10T17:35:20+08:00"}
	}]
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "type1/id1",
		"_type": "record",
		"_access": null,
		"date_value": {"$type": "date", "$date": "2015-04-10T09:35:20Z"},
		"_created_by":"user0",
		"_updated_by":"user0",
		"_ownerID": "user0"
	}]
}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"date_value": time.Date(2015, 4, 10, 9, 35, 20, 0, time.UTC),
				},
				OwnerID:   "user0",
				CreatorID: "user0",
				UpdaterID: "user0",
			})
		})

		Convey("Parses Asset", func() {
			resp := r.POST(`{
	"records": [{
		"_id": "type1/id1",
		"asset": {"$type": "asset", "$name": "asset-name", "$content_type":"plain/text"}
	}]
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "type1/id1",
		"_type": "record",
		"_access": null,
		"asset": {"$type": "asset", "$name": "asset-name", "$content_type":"plain/text"},
		"_created_by":"user0",
		"_updated_by":"user0",
		"_ownerID": "user0"
	}]
}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"asset": &skydb.Asset{
						Name:        "asset-name",
						ContentType: "plain/text",
					},
				},
				OwnerID:   "user0",
				CreatorID: "user0",
				UpdaterID: "user0",
			})
		})

		Convey("Parses Reference", func() {
			resp := r.POST(`{
	"records": [{
		"_id": "type1/id1",
		"ref": {"$type": "ref", "$id": "type2/id2"}
	}]
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "type1/id1",
		"_type": "record",
		"_access": null,
		"ref": {"$type": "ref", "$id": "type2/id2"},
		"_created_by":"user0",
		"_updated_by":"user0",
		"_ownerID": "user0"
	}]
}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"ref": skydb.NewReference("type2", "id2"),
				},
				OwnerID:   "user0",
				CreatorID: "user0",
				UpdaterID: "user0",
			})
		})

		Convey("Parses Location", func() {
			resp := r.POST(`{
	"records": [{
		"_id": "type1/id1",
		"geo": {"$type": "geo", "$lng": 1, "$lat": 2}
	}]
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "type1/id1",
		"_type": "record",
		"_access": null,
		"geo": {"$type": "geo", "$lng": 1, "$lat": 2},
		"_created_by":"user0",
		"_updated_by":"user0",
		"_ownerID": "user0"
	}]
}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"geo": skydb.NewLocation(1, 2),
				},
				OwnerID:   "user0",
				CreatorID: "user0",
				UpdaterID: "user0",
			})
		})
	})
}

type bogusFieldDatabaseConnection struct {
	skydb.Conn
}

func (db bogusFieldDatabaseConnection) GetRecordAccess(recordType string) (skydb.RecordACL, error) {
	return skydb.NewRecordACL([]skydb.RecordACLEntry{}), nil
}

func (db bogusFieldDatabaseConnection) GetRecordDefaultAccess(recordType string) (skydb.RecordACL, error) {
	return nil, nil
}

func (db bogusFieldDatabaseConnection) GetRecordFieldAccess() (skydb.FieldACL, error) {
	return skydb.FieldACL{}, nil
}

type bogusFieldDatabase struct {
	SaveFunc func(record *skydb.Record) error
	GetFunc  func(id skydb.RecordID, record *skydb.Record) error
	skydb.Database
}

func (db bogusFieldDatabase) IsReadOnly() bool { return false }

func (db bogusFieldDatabase) ID() string { return "" }

func (db bogusFieldDatabase) Extend(recordType string, schema skydb.RecordSchema) (bool, error) {
	return false, nil
}

func (db bogusFieldDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	return db.GetFunc(id, record)
}

func (db bogusFieldDatabase) Save(record *skydb.Record) error {
	return db.SaveFunc(record)
}

func (db bogusFieldDatabase) GetSchema(recordType string) (skydb.RecordSchema, error) {
	return skydb.RecordSchema{}, nil
}

func TestRecordSaveBogusField(t *testing.T) {
	timeNow = func() time.Time {
		return ZeroTime
	}
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordSaveHandler", t, func() {
		db := bogusFieldDatabase{}
		conn := bogusFieldDatabaseConnection{}
		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
			payload.DBConn = conn
			payload.Database = db
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("parse sequence field correctly", func(c C) {
			db.SaveFunc = func(record *skydb.Record) error {
				c.So(record, ShouldResemble, &skydb.Record{
					ID:        skydb.NewRecordID("record", "id"),
					Data:      skydb.Data{},
					OwnerID:   "user0",
					CreatorID: "user0",
					UpdaterID: "user0",
				})

				record.Data["seq"] = int64(1)

				return nil
			}
			db.GetFunc = func(id skydb.RecordID, record *skydb.Record) error {
				return skydb.ErrRecordNotFound
			}

			resp := r.POST(`{
				"records": [{
					"_id": "record/id",
					"seq": {"$type": "seq"}
				}]
			}`)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"seq": 1,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})

		Convey("can save without specifying seq", func(c C) {
			db.SaveFunc = func(record *skydb.Record) error {
				c.So(record, ShouldResemble, &skydb.Record{
					ID:        skydb.NewRecordID("record", "id"),
					Data:      skydb.Data{},
					OwnerID:   "user0",
					CreatorID: "user0",
					UpdaterID: "user0",
				})
				record.Data["seq"] = int64(2)
				return nil
			}
			db.GetFunc = func(id skydb.RecordID, record *skydb.Record) error {
				c.So(id, ShouldResemble, skydb.NewRecordID("record", "id"))
				record.ID = skydb.NewRecordID("record", "id")
				record.OwnerID = "user0"
				record.CreatorID = "user0"
				record.CreatedAt = timeNow()
				record.UpdaterID = "user0"
				record.UpdatedAt = timeNow()
				record.Data = skydb.Data{
					"seq": int64(1),
				}
				return nil
			}

			resp := r.POST(`{
				"records": [{
					"_id": "record/id"
				}]
			}`)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"seq": 2,
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})
	})
}

type noExtendDatabase struct {
	calledExtend bool
	skydb.Database
}

func (db *noExtendDatabase) IsReadOnly() bool { return false }

func (db *noExtendDatabase) Extend(recordType string, schema skydb.RecordSchema) (bool, error) {
	db.calledExtend = true
	return false, errors.New("You shalt not call Extend")
}

func TestRecordSaveNoExtendIfRecordMalformed(t *testing.T) {
	Convey("RecordSaveHandler", t, func() {
		noExtendDB := &noExtendDatabase{}
		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
			payload.Database = noExtendDB
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("REGRESSION #119: Database.Extend should be called when all record are invalid", func() {
			r.POST(`{
				"records": [{
				}, {
					"_id": "invalidkey"
				}]
			}`)
			So(noExtendDB.calledExtend, ShouldBeFalse)
		})
	})
}

type queryDatabase struct {
	lastquery  *skydb.Query
	databaseID string
	skydb.Database
}

func (db *queryDatabase) IsReadOnly() bool { return false }

func (db *queryDatabase) ID() string {
	if db.databaseID == "" {
		return skydb.PublicDatabaseIdentifier
	}
	return db.databaseID
}

func (db *queryDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	db.lastquery = query
	return 0, nil
}

func (db *queryDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	db.lastquery = query
	return skydb.EmptyRows, nil
}

type queryResultsDatabase struct {
	records    []skydb.Record
	databaseID string
	typemap    map[string]skydb.RecordSchema
	skydb.Database
}

func (db *queryResultsDatabase) IsReadOnly() bool { return false }

func (db *queryResultsDatabase) ID() string {
	if db.databaseID == "" {
		return skydb.PublicDatabaseIdentifier
	}
	return db.databaseID
}

func (db *queryResultsDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	return uint64(len(db.records)), nil
}

func (db *queryResultsDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	return skydb.NewRows(skydb.NewMemoryRows(db.records)), nil
}

func (db *queryResultsDatabase) GetSchema(recordType string) (skydb.RecordSchema, error) {
	return db.typemap[recordType], nil
}

func TestRecordQueryResults(t *testing.T) {
	Convey("Given a Database with records", t, func() {
		record0 := skydb.Record{
			ID: skydb.NewRecordID("note", "0"),
		}
		record1 := skydb.Record{
			ID: skydb.NewRecordID("note", "1"),
		}
		record2 := skydb.Record{
			ID: skydb.NewRecordID("note", "2"),
		}

		conn := skydbtest.NewMapConn()
		db := &queryResultsDatabase{}
		db.records = []skydb.Record{record1, record0, record2}

		r := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
		})

		Convey("REGRESSION #227: query returns correct results from db", func() {
			resp := r.POST(`{
				"record_type": "note"
			}`)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": [{
					"_type": "record",
					"_id": "note/1",
					"_access": null
				},
				{
					"_type": "record",
					"_id": "note/0",
					"_access": null
				},
				{
					"_type": "record",
					"_id": "note/2",
					"_access": null
				}]
			}`)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}

func TestRecordQuery(t *testing.T) {
	Convey("Given a Database", t, func() {
		db := &queryDatabase{}
		conn := skydbtest.NewMapConn()

		Convey("Queries records with type", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type: "note",
			})
		})

		Convey("Queries records with type and user", func() {
			authInfo := skydb.AuthInfo{
				ID: "user0",
			}
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
				},
				DBConn:   conn,
				Database: db,
				AuthInfo: &authInfo,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type:       "note",
				ViewAsUser: &authInfo,
			})
		})

		Convey("Queries records with type and master key", func() {
			authInfo := skydb.AuthInfo{
				ID: "user0",
			}
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
				},
				DBConn:    conn,
				Database:  db,
				AuthInfo:  &authInfo,
				AccessKey: router.MasterAccessKey,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type:                "note",
				ViewAsUser:          &authInfo,
				BypassAccessControl: true,
			})
		})

		Convey("Queries records with sorting", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"sort": []interface{}{
						[]interface{}{
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "noteOrder",
							},
							"desc",
						},
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						Expression: skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "noteOrder",
						},
						Order: skydb.Desc,
					},
				},
			})
		})

		Convey("Queries records with sorting by distance function", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"sort": []interface{}{
						[]interface{}{
							[]interface{}{
								"func",
								"distance",
								map[string]interface{}{
									"$type": "keypath",
									"$val":  "location",
								},
								map[string]interface{}{
									"$type": "geo",
									"$lng":  float64(1),
									"$lat":  float64(2),
								},
							},
							"desc",
						},
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						Expression: skydb.Expression{
							Type: skydb.Function,
							Value: skydb.DistanceFunc{
								Field:    "location",
								Location: skydb.NewLocation(1, 2),
							},
						},
						Order: skydb.Desc,
					},
				},
			})
		})

		Convey("Queries records with predicate", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"predicate": []interface{}{
						"eq",
						map[string]interface{}{
							"$type": "keypath",
							"$val":  "noteOrder",
						},
						float64(1),
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
				Operator: skydb.Equal,
				Children: []interface{}{
					skydb.Expression{skydb.KeyPath, "noteOrder"},
					skydb.Expression{skydb.Literal, float64(1)},
				},
			})
		})

		Convey("Queries records with complex predicate", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"predicate": []interface{}{
						"and",
						[]interface{}{
							"eq",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "content",
							},
							"text",
						},
						[]interface{}{
							"gt",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "noteOrder",
							},
							float64(1),
						},
						[]interface{}{
							"neq",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "content",
							},
							nil,
						},
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
				Operator: skydb.And,
				Children: []interface{}{
					skydb.Predicate{
						Operator: skydb.Equal,
						Children: []interface{}{
							skydb.Expression{skydb.KeyPath, "content"},
							skydb.Expression{skydb.Literal, "text"},
						},
					},
					skydb.Predicate{
						Operator: skydb.GreaterThan,
						Children: []interface{}{
							skydb.Expression{skydb.KeyPath, "noteOrder"},
							skydb.Expression{skydb.Literal, float64(1)},
						},
					},
					skydb.Predicate{
						Operator: skydb.NotEqual,
						Children: []interface{}{
							skydb.Expression{skydb.KeyPath, "content"},
							skydb.Expression{skydb.Literal, nil},
						},
					},
				},
			})
		})

		Convey("Queries records by distance func", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"predicate": []interface{}{
						"lte",
						[]interface{}{
							"func",
							"distance",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "location",
							},
							map[string]interface{}{
								"$type": "geo",
								"$lng":  float64(1),
								"$lat":  float64(2),
							},
						},
						float64(500),
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
				Operator: skydb.LessThanOrEqual,
				Children: []interface{}{
					skydb.Expression{
						skydb.Function,
						skydb.DistanceFunc{
							Field:    "location",
							Location: skydb.NewLocation(1, 2),
						},
					},
					skydb.Expression{skydb.Literal, float64(500)},
				},
			})
		})

		Convey("Return calculated distance", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"include": map[string]interface{}{
						"distance": []interface{}{
							"func",
							"distance",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "location",
							},
							map[string]interface{}{
								"$type": "geo",
								"$lng":  float64(1),
								"$lat":  float64(2),
							},
						},
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.ComputedKeys, ShouldResemble, map[string]skydb.Expression{
				"distance": skydb.Expression{
					skydb.Function,
					skydb.DistanceFunc{
						Field:    "location",
						Location: skydb.NewLocation(1, 2),
					},
				},
			})
		})

		Convey("Return records with desired keys only", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type":  "note",
					"desired_keys": []interface{}{"location"},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.DesiredKeys, ShouldResemble, []string{"location"})
		})

		Convey("Return records when desired keys is empty", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type":  "note",
					"desired_keys": []interface{}{},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.DesiredKeys, ShouldResemble, []string{})
		})

		Convey("Return records when desired keys is nil", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type":  "note",
					"desired_keys": nil,
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.DesiredKeys, ShouldBeNil)
		})

		Convey("Queries records with offset", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"limit":       float64(200),
					"offset":      float64(400),
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.Limit, ShouldNotBeNil)
			So(*db.lastquery.Limit, ShouldEqual, 200)
			So(db.lastquery.Offset, ShouldEqual, 400)
		})

		Convey("Queries records with count", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"count":       true,
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.GetCount, ShouldBeTrue)
		})

		Convey("Propagate invalid query error", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
					"predicate": []interface{}{
						"eq",
						map[string]interface{}{
							"$type": "keypath",
							"$val":  "content",
						},
						map[string]interface{}{},
					},
				},
				DBConn:   conn,
				Database: db,
			}
			response := router.Response{}

			handler := &RecordQueryHandler{}
			handler.Handle(&payload, &response)

			So(response.Err, ShouldNotBeNil)
		})

		Convey("with FieldACL", func() {
			publicRole := skydb.FieldUserRole{skydb.PublicFieldUserRoleType, ""}

			conn.SetRecordFieldAccess(skydb.NewFieldACL(skydb.FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "category",
					UserRole:     publicRole,
					Writable:     false,
					Readable:     false,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "index",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     true,
					Comparable:   false,
					Discoverable: false,
				},
				{
					RecordType:   "note",
					RecordField:  "title",
					UserRole:     publicRole,
					Writable:     true,
					Readable:     true,
					Comparable:   false,
					Discoverable: true,
				},
			}))

			Convey("should block non-comparable, non-discoverable field", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"predicate": []interface{}{
							"gt",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "index",
							},
							float64(1),
						},
					},
					DBConn:   conn,
					Database: db,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldNotBeNil)
				So(response.Err.Code(), ShouldEqual, skyerr.RecordQueryDenied)
			})

			Convey("should not block non-comparable, non-discoverable field with master key", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"predicate": []interface{}{
							"gt",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "index",
							},
							float64(1),
						},
					},
					DBConn:    conn,
					Database:  db,
					AccessKey: router.MasterAccessKey,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldBeNil)
			})

			Convey("should block non-comparable", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"predicate": []interface{}{
							"gt",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "title",
							},
							"Tale of Two Cities",
						},
					},
					DBConn:   conn,
					Database: db,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldNotBeNil)
				So(response.Err.Code(), ShouldEqual, skyerr.RecordQueryDenied)
			})

			Convey("should allow comparable field with equality", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"predicate": []interface{}{
							"eq",
							map[string]interface{}{
								"$type": "keypath",
								"$val":  "title",
							},
							"Tale of Two Cities",
						},
					},
					DBConn:   conn,
					Database: db,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldBeNil)
			})

			Convey("should block non-comparable field in sort", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"sort": []interface{}{
							[]interface{}{
								map[string]interface{}{
									"$type": "keypath",
									"$val":  "title",
								},
								"desc",
							},
						},
					},
					DBConn:   conn,
					Database: db,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldNotBeNil)
			})

			Convey("should block non-readable field in transient include", func() {
				payload := router.Payload{
					Data: map[string]interface{}{
						"record_type": "note",
						"include": map[string]interface{}{
							"category": map[string]interface{}{"$type": "keypath", "$val": "category"},
						},
					},
					DBConn:   conn,
					Database: db,
				}
				response := router.Response{}

				handler := &RecordQueryHandler{}
				handler.Handle(&payload, &response)
				So(response.Err, ShouldNotBeNil)
			})
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type singleRecordDatabase struct {
	record       skydb.Record
	databaseID   string
	recordSchema skydb.RecordSchema
	skydb.Database
}

func (db *singleRecordDatabase) IsReadOnly() bool { return false }

func (db *singleRecordDatabase) ID() string {
	if db.databaseID == "" {
		return skydb.PublicDatabaseIdentifier
	}
	return db.databaseID
}

func (db *singleRecordDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	*record = db.record
	return nil
}

func (db *singleRecordDatabase) Save(record *skydb.Record) error {
	*record = db.record
	return nil
}

func (db *singleRecordDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	return uint64(1), nil
}

func (db *singleRecordDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	return skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{db.record})), nil
}

func (db *singleRecordDatabase) Extend(recordType string, schema skydb.RecordSchema) (bool, error) {
	return false, nil
}

func (db *singleRecordDatabase) GetSchema(recordType string) (skydb.RecordSchema, error) {
	return db.recordSchema, nil
}

func TestRecordFetchHandler(t *testing.T) {
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordFetchHandler with Field ACL", t, func() {
		db := skydbtest.NewMapDB()
		db.RecordSchemaMap = skydbtest.RecordSchemaMap{
			"note": {
				"content":  skydb.FieldType{Type: skydb.TypeString},
				"favorite": skydb.FieldType{Type: skydb.TypeBoolean},
				"category": skydb.FieldType{Type: skydb.TypeString},
			},
		}
		conn := skydbtest.NewMapConn()
		publicRole := skydb.FieldUserRole{skydb.PublicFieldUserRoleType, ""}

		db.Save(&skydb.Record{
			ID:        skydb.NewRecordID("note", "note0"),
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

		conn.SetRecordFieldAccess(skydb.NewFieldACL(skydb.FieldACLEntryList{
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

		r := handlertest.NewSingleRouteRouter(&RecordFetchHandler{}, func(payload *router.Payload) {
			payload.DBConn = conn
			payload.Database = db
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
			if apiKey, ok := payload.Data["api_key"]; ok && apiKey == "master" {
				payload.AccessKey = router.MasterAccessKey
			}
		})

		Convey("should fetch without non-readable fields", func() {
			resp := r.POST(`{
				"ids": ["note/note0"]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
					"_type": "record",
					"_access": null,
					"content": "Hello World!",
					"_created_by":"user0",
					"_updated_by":"user0",
					"_ownerID": "user0"
				}]
			}`)
		})

		Convey("should fetch with all fields with master key", func() {
			resp := r.POST(`{
				"api_key": "master",
				"ids": ["note/note0"]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note0",
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

func TestRecordOwnerIDSerialization(t *testing.T) {
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("Given a record with owner id in DB", t, func() {
		record := skydb.Record{
			ID:      skydb.NewRecordID("type", "id"),
			OwnerID: "ownerID",
		}
		db := &singleRecordDatabase{
			record:       record,
			recordSchema: skydb.RecordSchema{},
		}
		conn := skydbtest.NewMapConn()

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
			payload.DBConn = conn
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "ownerID",
			}
		}

		Convey("fetched record serializes owner id correctly", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordFetchHandler{}, injectDBFunc).POST(`{
				"ids": ["do/notCare"]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID"
				}]
			}`)
		})

		Convey("saved record serializes owner id correctly", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, injectDBFunc).POST(`{
				"records": [{
					"_id": "type/id"
				}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID"
				}]
			}`)
		})

		Convey("queried record serializes owner id correctly", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, injectDBFunc).POST(`{
				"record_type": "doNotCare"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID"
				}]
			}`)
		})
	})
}

func TestRecordMetaData(t *testing.T) {
	Convey("Record Meta Data", t, func() {
		db := skydbtest.NewMapDB()
		conn := skydbtest.NewMapConn()
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = timeNowUTC
		}()

		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
			payload.DBConn = conn
			payload.Database = db
			payload.AuthInfoID = "requestUserID"
			payload.AuthInfo = &skydb.AuthInfo{
				ID: "requestUserID",
			}
		})
		Convey("on a newly created record", func() {

			req := r.POST(`{
				"records": [{
					"_id": "record/id"
				}]
			}`)
			So(req.Body.String(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"_ownerID": "requestUserID",
					"_created_at": "2006-01-02T15:04:05Z",
					"_created_by": "requestUserID",
					"_updated_at": "2006-01-02T15:04:05Z",
					"_updated_by": "requestUserID"
				}]
			}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("record", "id"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID:        skydb.NewRecordID("record", "id"),
				OwnerID:   "requestUserID",
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				CreatorID: "requestUserID",
				UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID: "requestUserID",
				Data:      skydb.Data{},
			})
		})

		Convey("on an existing record", func() {
			db.Save(&skydb.Record{
				ID:        skydb.NewRecordID("record", "id"),
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 4, 0, time.UTC),
				CreatorID: "creatorID",
				UpdatedAt: time.Date(2006, 1, 2, 15, 4, 4, 0, time.UTC),
				UpdaterID: "updaterID",
			})

			req := r.POST(`{
				"records": [{
					"_id": "record/id"
				}]
			}`)
			So(req.Body.String(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"_created_at": "2006-01-02T15:04:04Z",
					"_created_by": "creatorID",
					"_updated_at": "2006-01-02T15:04:05Z",
					"_updated_by": "requestUserID"
				}]
			}`)

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("record", "id"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID:        skydb.NewRecordID("record", "id"),
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 4, 0, time.UTC),
				CreatorID: "creatorID",
				UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID: "requestUserID",
				Data:      skydb.Data{},
			})

		})
	})
}

type urlOnlyAssetStore struct{}

func (s *urlOnlyAssetStore) GetFileReader(name string) (io.ReadCloser, error) {
	panic("not implemented")
}

func (s *urlOnlyAssetStore) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	panic("not implemented")
}

func (s *urlOnlyAssetStore) GeneratePostFileRequest(name string) (*asset.PostFileRequest, error) {
	panic("not implemented")
}

func (s *urlOnlyAssetStore) SignedURL(name string) (string, error) {
	return fmt.Sprintf("http://skygear.test/asset/%s?expiredAt=1997-07-01T00:00:00", name), nil
}

func (s *urlOnlyAssetStore) IsSignatureRequired() bool {
	return true
}

func TestRecordAssetSerialization(t *testing.T) {
	Convey("RecordAssetSerialization for fetch", t, func() {
		conn := skydbtest.NewMapConn()
		db := skydbtest.NewMapDB()
		db.Save(&skydb.Record{
			ID: skydb.NewRecordID("record", "id"),
			Data: map[string]interface{}{
				"asset": &skydb.Asset{
					Name:        "asset-name",
					ContentType: "plain/text",
				},
			},
		})

		assetStore := &urlOnlyAssetStore{}

		r := handlertest.NewSingleRouteRouter(&RecordFetchHandler{
			AssetStore: assetStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
		})

		Convey("serialize with $url", func() {
			resp := r.POST(`{
				"ids": ["record/id"]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"asset": {
						"$type": "asset",
						"$name": "asset-name",
						"$url": "http://skygear.test/asset/asset-name?expiredAt=1997-07-01T00:00:00",
						"$content_type":"plain/text"
					}
				}]
			}`)
		})
	})

	Convey("RecordAssetSerialization for query", t, func() {
		record0 := skydb.Record{
			ID: skydb.NewRecordID("record", "id"),
			Data: map[string]interface{}{
				"asset": &skydb.Asset{
					Name:        "asset-name",
					ContentType: "plain/text",
				},
			},
		}

		conn := skydbtest.NewMapConn()
		db := &queryResultsDatabase{}
		db.records = []skydb.Record{record0}

		assetStore := &urlOnlyAssetStore{}

		r := handlertest.NewSingleRouteRouter(&RecordQueryHandler{
			AssetStore: assetStore,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
		})

		Convey("serialize with $url", func() {
			resp := r.POST(`{
				"record_type": "record"
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"asset": {
						"$type": "asset",
						"$name": "asset-name",
						"$url": "http://skygear.test/asset/asset-name?expiredAt=1997-07-01T00:00:00",
						"$content_type":"plain/text"
					}
				}]
			}`)
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type referencedRecordDatabase struct {
	note       skydb.Record
	category   skydb.Record
	city       skydb.Record
	user       skydb.Record
	databaseID string
	skydb.Database
}

func (db *referencedRecordDatabase) IsReadOnly() bool { return false }

func (db *referencedRecordDatabase) ID() string {
	if db.databaseID == "" {
		return skydb.PublicDatabaseIdentifier
	}
	return db.databaseID
}

func (db *referencedRecordDatabase) UserRecordType() string { return "user" }

func (db *referencedRecordDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	switch id.String() {
	case "note/note1":
		*record = db.note
	case "category/important":
		*record = db.category
	case "city/beautiful":
		*record = db.city
	case "user/ownerID":
		*record = db.user
	}
	return nil
}

func (db *referencedRecordDatabase) GetByIDs(ids []skydb.RecordID) (*skydb.Rows, error) {
	records := []skydb.Record{}
	for _, id := range ids {
		switch id.String() {
		case "note/note1":
			records = append(records, db.note)
		case "category/important":
			records = append(records, db.category)
		case "city/beautiful":
			records = append(records, db.city)
		case "user/ownerID":
			records = append(records, db.user)
		}
	}
	return skydb.NewRows(skydb.NewMemoryRows(records)), nil
}

func (db *referencedRecordDatabase) Save(record *skydb.Record) error {
	return nil
}

func (db *referencedRecordDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	return uint64(1), nil
}

func (db *referencedRecordDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	return skydb.NewRows(skydb.NewMemoryRows([]skydb.Record{db.note})), nil
}

func (db *referencedRecordDatabase) Extend(recordType string, schema skydb.RecordSchema) (bool, error) {
	return false, nil
}

func (db *referencedRecordDatabase) GetSchema(recordType string) (skydb.RecordSchema, error) {
	typemap := map[string]skydb.RecordSchema{
		"note": skydb.RecordSchema{
			"category": skydb.FieldType{
				Type:          skydb.TypeReference,
				ReferenceType: "category",
			},
			"city": skydb.FieldType{
				Type:          skydb.TypeReference,
				ReferenceType: "city",
			},
		},
		"category": skydb.RecordSchema{
			"title": skydb.FieldType{
				Type: skydb.TypeString,
			},
		},
		"city": skydb.RecordSchema{
			"name": skydb.FieldType{
				Type: skydb.TypeString,
			},
		},
		"user": skydb.RecordSchema{
			"name": skydb.FieldType{
				Type: skydb.TypeString,
			},
		},
	}
	return typemap[recordType], nil
}

func TestRecordQueryWithEagerLoad(t *testing.T) {
	Convey("Given a referenced record in DB", t, func() {
		db := &referencedRecordDatabase{
			note: skydb.Record{
				ID:      skydb.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": skydb.NewReference("category", "important"),
					"city":     skydb.NewReference("city", "beautiful"),
				},
			},
			category: skydb.Record{
				ID:      skydb.NewRecordID("category", "important"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"title": "This is important.",
				},
			},
			city: skydb.Record{
				ID:      skydb.NewRecordID("city", "beautiful"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "This is beautiful.",
				},
			},
			user: skydb.Record{
				ID:      skydb.NewRecordID("user", "ownerID"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "Owner",
				},
			},
		}
		conn := skydbtest.NewMapConn()

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
			payload.DBConn = conn
		}

		Convey("query record with eager load", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, injectDBFunc).POST(`{
				"record_type": "note",
				"include": {"category": {"$type": "keypath", "$val": "category"}}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note1",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID",
					"category": {"$id":"category/important","$type":"ref"},
					"city": {"$id":"city/beautiful","$type":"ref"},
					"_transient": {
						"category": {"_access":null,"_id":"category/important","_type":"record","_ownerID":"ownerID", "title": "This is important."}
					}
				}]
			}`)
		})

		Convey("query record with multiple eager load", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, injectDBFunc).POST(`{
				"record_type": "note",
				"include": {
					"category": {"$type": "keypath", "$val": "category"},
					"city": {"$type": "keypath", "$val": "city"}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note1",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID",
					"category": {"$id":"category/important","$type":"ref"},
					"city": {"$id":"city/beautiful","$type":"ref"},
					"_transient": {
						"category": {"_access":null,"_id":"category/important","_type":"record","_ownerID":"ownerID", "title": "This is important."},
						"city": {"_access":null,"_id":"city/beautiful","_type":"record","_ownerID":"ownerID", "name": "This is beautiful."}
					}
				}]
			}`)
		})

		Convey("query record with eager load on user", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, injectDBFunc).POST(`{
				"record_type": "note",
				"include": {"user": {"$type": "keypath", "$val": "_owner"}}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note1",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID",
					"category": {"$id":"category/important","$type":"ref"},
					"city": {"$id":"city/beautiful","$type":"ref"},
					"_transient": {
						"user": {"_access":null,"_id":"user/ownerID","_type":"record","_ownerID":"ownerID", "name": "Owner"}
					}
				}]
			}`)
		})

	})

	Convey("Given a referenced record with null reference in DB", t, func() {
		db := &referencedRecordDatabase{
			note: skydb.Record{
				ID:      skydb.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": skydb.NewReference("category", "important"),
					"city":     nil,
				},
			},
			category: skydb.Record{
				ID:      skydb.NewRecordID("category", "important"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"title": "This is important.",
				},
			},
			city: skydb.Record{
				ID:      skydb.NewRecordID("city", "beautiful"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "This is beautiful.",
				},
			},
		}
		conn := skydbtest.NewMapConn()

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
			payload.DBConn = conn
		}

		Convey("query record with eager load", func() {
			resp := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, injectDBFunc).POST(`{
				"record_type": "note",
				"include": {"city": {"$type": "keypath", "$val": "city"}}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note1",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID",
					"category": {"$id":"category/important","$type":"ref"},
					"city": null,
					"_transient": {
						"city": null
					}
				}]
			}`)
		})
	})
}

func TestRecordQueryWithCount(t *testing.T) {
	Convey("Given a Database with records", t, func() {
		record0 := skydb.Record{
			ID: skydb.NewRecordID("note", "0"),
		}
		record1 := skydb.Record{
			ID: skydb.NewRecordID("note", "1"),
		}
		record2 := skydb.Record{
			ID: skydb.NewRecordID("note", "2"),
		}

		conn := skydbtest.NewMapConn()
		db := &queryResultsDatabase{}
		db.records = []skydb.Record{record1, record0, record2}
		db.typemap = map[string]skydb.RecordSchema{
			"note": skydb.RecordSchema{},
		}

		r := handlertest.NewSingleRouteRouter(&RecordQueryHandler{}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
		})

		Convey("get count of records", func() {
			resp := r.POST(`{
				"record_type": "note",
				"count": true
			}`)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"info": {
					"count": 3
				},
				"result": [{
					"_type": "record",
					"_id": "note/1",
					"_access": null
				},
				{
					"_type": "record",
					"_id": "note/0",
					"_access": null
				},
				{
					"_type": "record",
					"_id": "note/2",
					"_access": null
				}
				]
			}`)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}

type erroneousDB struct {
	skydb.Database
}

func (db erroneousDB) IsReadOnly() bool { return false }

func (db erroneousDB) Extend(string, skydb.RecordSchema) (bool, error) {
	return false, nil
}

func (db erroneousDB) Get(skydb.RecordID, *skydb.Record) error {
	return errors.New("erroneous save")
}

func (db erroneousDB) Save(*skydb.Record) error {
	return errors.New("erroneous save")
}

func TestHookExecution(t *testing.T) {
	Convey("Record(Save|Delete)Handler", t, func() {
		registry := hook.NewRegistry()
		handlerTests := []struct {
			kind             string
			handler          router.Handler
			beforeActionKind hook.Kind
			afterActionKind  hook.Kind
			reqBody          string
		}{
			{
				"Save",
				&RecordSaveHandler{
					HookRegistry: registry,
				},
				hook.BeforeSave,
				hook.AfterSave,
				`{"records": [{"_id": "record/id"}]}`,
			},
			{
				"Delete",
				&RecordDeleteHandler{
					HookRegistry: registry,
				},
				hook.BeforeDelete,
				hook.AfterDelete,
				`{"ids": ["record/id"]}`,
			},
		}

		record := &skydb.Record{
			ID: skydb.NewRecordID("record", "id"),
		}

		beforeHook := hooktest.StackingHook{}
		afterHook := hooktest.StackingHook{}

		for _, test := range handlerTests {
			testName := fmt.Sprintf("executes Before%[1]s and After%[1]s action hooks", test.kind)
			Convey(testName, func() {
				registry.Register(test.beforeActionKind, "record", beforeHook.Func)
				registry.Register(test.afterActionKind, "record", afterHook.Func)

				db := skydbtest.NewMapDB()
				So(db.Save(record), ShouldBeNil)

				r := handlertest.NewSingleRouteRouter(test.handler, func(p *router.Payload) {
					p.Database = db
					p.DBConn = skydbtest.NewMapConn()
					p.AuthInfo = &skydb.AuthInfo{
						ID: "user0",
					}
				})

				r.POST(test.reqBody)

				So(len(beforeHook.Records), ShouldEqual, 1)
				So(beforeHook.Records[0].ID, ShouldResemble, record.ID)
				So(len(afterHook.Records), ShouldEqual, 1)
				So(afterHook.Records[0].ID, ShouldResemble, record.ID)
			})

			testName = fmt.Sprintf("doesn't execute After%[1]s hooks if db.%[1]s returns an error", test.kind)
			Convey(testName, func() {
				registry.Register(test.afterActionKind, "record", afterHook.Func)
				r := handlertest.NewSingleRouteRouter(test.handler, func(p *router.Payload) {
					p.Database = erroneousDB{}
					p.AuthInfo = &skydb.AuthInfo{
						ID: "user0",
					}
				})

				r.POST(test.reqBody)
				So(afterHook.Records, ShouldBeEmpty)
			})
		}
	})

	Convey("HookRegistry", t, func() {
		registry := hook.NewRegistry()
		db := skydbtest.NewMapDB()
		conn := skydbtest.NewMapConn()
		r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{
			HookRegistry: registry,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.Database = db
			p.AuthInfo = &skydb.AuthInfo{
				ID: "user0",
			}
		})

		Convey("record is not saved if BeforeSave's hook returns an error", func() {
			registry.Register(hook.BeforeSave, "record", func(context.Context, *skydb.Record, *skydb.Record) skyerr.Error {
				return skyerr.NewError(skyerr.UnexpectedError, "no hooks for you")
			})
			r.POST(`{
				"records": [{
					"_id": "record/id"
				}]
			}`)

			var record skydb.Record
			So(db.Get(skydb.NewRecordID("record", "id"), &record), ShouldEqual, skydb.ErrRecordNotFound)
		})

		Convey("BeforeSave should be fed fully fetched record", func() {
			existingRecord := skydb.Record{
				ID: skydb.NewRecordID("record", "id"),
				Data: map[string]interface{}{
					"old": true,
				},
			}
			So(db.Save(&existingRecord), ShouldBeNil)

			called := false
			registry.Register(hook.BeforeSave, "record", func(ctx context.Context, record *skydb.Record, originalRecord *skydb.Record) skyerr.Error {
				called = true
				So(*record, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("record", "id"),
					Data: map[string]interface{}{
						"old": true,
						"new": true,
					},
				})
				So(*originalRecord, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("record", "id"),
					Data: map[string]interface{}{
						"old": true,
					},
				})
				return nil
			})

			r.POST(`{
				"records": [{
					"_id": "record/id",
					"new": true
				}]
			}`)

			So(called, ShouldBeTrue)
		})

		Convey("BeforeSave should set originalRecord as nil for new record", func() {
			called := false
			registry.Register(hook.BeforeSave, "record", func(ctx context.Context, record *skydb.Record, originalRecord *skydb.Record) skyerr.Error {
				called = true
				So(*record, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("record", "id"),
					Data: map[string]interface{}{
						"new": true,
					},
					OwnerID: "user0",
				})
				So(originalRecord, ShouldBeNil)
				return nil
			})

			r.POST(`{
				"records": [{
					"_id": "record/id",
					"new": true
				}]
			}`)

			So(called, ShouldBeTrue)
		})
	})
}

type filterFuncDef func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error

// selectiveDatabase filter Get, Save and Delete by executing filterFunc
// if filterFunc return nil, the operation is delegated to underlying Database
// otherwise, the error is returned directly
type selectiveDatabase struct {
	filterFunc filterFuncDef
	skydb.Database
}

func newSelectiveDatabase(backingDB skydb.Database) *selectiveDatabase {
	return &selectiveDatabase{
		Database: backingDB,
	}
}

func (db *selectiveDatabase) IsReadOnly() bool { return false }

func (db *selectiveDatabase) SetFilter(filterFunc filterFuncDef) {
	db.filterFunc = filterFunc
}

func (db *selectiveDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	if err := db.filterFunc("GET", id, nil); err != nil {
		return err
	}

	return db.Database.Get(id, record)
}

func (db *selectiveDatabase) Save(record *skydb.Record) error {
	if err := db.filterFunc("SAVE", record.ID, record); err != nil {
		return err
	}

	return db.Database.Save(record)
}

func (db *selectiveDatabase) Delete(id skydb.RecordID) error {
	if err := db.filterFunc("DELETE", id, nil); err != nil {
		return err
	}

	return db.Database.Delete(id)
}

func (db *selectiveDatabase) Begin() error {
	return db.Database.(skydb.TxDatabase).Begin()
}

func (db *selectiveDatabase) Commit() error {
	return db.Database.(skydb.TxDatabase).Commit()
}

func (db *selectiveDatabase) Rollback() error {
	return db.Database.(skydb.TxDatabase).Rollback()
}

func TestAtomicOperation(t *testing.T) {
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("Atomic Operation", t, func() {
		conn := skydbtest.NewMapConn()
		backingDB := skydbtest.NewMapDB()
		txDB := skydbtest.NewMockTxDatabase(backingDB)
		db := newSelectiveDatabase(txDB)

		Convey("for RecordSaveHandler", func() {
			r := handlertest.NewSingleRouteRouter(&RecordSaveHandler{}, func(payload *router.Payload) {
				payload.DBConn = conn
				payload.Database = db
				payload.AuthInfo = &skydb.AuthInfo{
					ID: "user0",
				}
			})

			Convey("rolls back saved records on error", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error {
					if op == "SAVE" && recordID.Key == "1" {
						return skyerr.NewError(skyerr.UnexpectedError, "Original Sin")
					}
					return nil
				})

				resp := r.POST(`{
					"records": [{
						"_id": "note/0",
						"_type": "record"
					},
					{
						"_id": "note/1",
						"_type": "record"
					},
					{
						"_id": "note/2",
						"_type": "record"
					}],
					"atomic": true
				}`)

				So(resp.Body.String(), ShouldEqualJSON, `{
					"error": {
						"code": 115,
						"name": "AtomicOperationFailure",
						"message": "Atomic Operation rolled back due to one or more errors",
						"info": {
							"note/1": {
								"code": 10000,
								"message": "Original Sin",
								"name": "UnexpectedError"
							}
						}
					}
				}`)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeFalse)
				So(txDB.DidRollback, ShouldBeTrue)
			})

			Convey("commit saved records when there are no errors", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error {
					return nil
				})

				resp := r.POST(`{
					"records": [{
						"_id": "note/0",
						"_type": "record"
					},
					{
						"_id": "note/1",
						"_type": "record"
					}],
					"atomic": true
				}`)

				So(resp.Body.String(), ShouldEqualJSON, `{
					"result": [{
							"_id": "note/0",
							"_type": "record",
							"_access": null,
							"_created_by":"user0",
							"_updated_by":"user0",
							"_ownerID": "user0"
						}, {
							"_id": "note/1",
							"_type": "record",
							"_access": null,
							"_created_by":"user0",
							"_updated_by":"user0",
							"_ownerID": "user0"
						}]
				}`)

				var record skydb.Record
				So(backingDB.Get(skydb.NewRecordID("note", "0"), &record), ShouldBeNil)
				So(record, ShouldResemble, skydb.Record{
					ID:        skydb.NewRecordID("note", "0"),
					Data:      map[string]interface{}{},
					OwnerID:   "user0",
					CreatorID: "user0",
					UpdaterID: "user0",
				})
				So(backingDB.Get(skydb.NewRecordID("note", "1"), &record), ShouldBeNil)
				So(record, ShouldResemble, skydb.Record{
					ID:        skydb.NewRecordID("note", "1"),
					Data:      map[string]interface{}{},
					OwnerID:   "user0",
					CreatorID: "user0",
					UpdaterID: "user0",
				})

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeTrue)
				So(txDB.DidRollback, ShouldBeFalse)
			})

			Convey("fails whole request on any records mal-format", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error {
					return nil
				})

				resp := r.POST(`{
					"records": [{
						"_id": "note0",
						"_type": "record"
					},
					{
						"_id": "note/1",
						"_access": "note/1"
					}],
					"atomic": true
				}`)

				So(resp.Body.String(), ShouldEqualJSON, `{
					"error": {
						"code": 108,
						"info": {
							"arguments": "records",
							"errors": [{
								"code": 108,
								"info": {"arguments":["id"]},
								"message": "record: \"_id\" should be of format '{type}/{id}', got \"note0\"",
								"name": "InvalidArgument"
							}, {
								"code": 108,
								"info": {"arguments":["_access"]},
								"message": "_access must be an array",
								"name": "InvalidArgument"
							}]
						},
						"message": "fails to de-serialize records",
						"name": "InvalidArgument"
					}
				}`)

				So(txDB.DidBegin, ShouldBeFalse)
				So(txDB.DidCommit, ShouldBeFalse)
				So(txDB.DidRollback, ShouldBeFalse)
			})
		})

		Convey("for RecordDeleteHandler", func() {
			So(backingDB.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "0"),
			}), ShouldBeNil)
			So(backingDB.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "1"),
			}), ShouldBeNil)
			So(backingDB.Save(&skydb.Record{
				ID: skydb.NewRecordID("note", "2"),
			}), ShouldBeNil)

			r := handlertest.NewSingleRouteRouter(&RecordDeleteHandler{}, func(payload *router.Payload) {
				payload.Database = db
				payload.AuthInfo = &skydb.AuthInfo{
					ID: "user0",
				}
			})

			Convey("rolls back deleted records on error", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error {
					if op == "DELETE" && recordID.Key == "1" {
						return skyerr.NewError(skyerr.UnexpectedError, "Original Sin")
					}
					return nil
				})

				resp := r.POST(`{
					"ids": [
						"note/0",
						"note/1",
						"note/2"
					],
					"atomic": true
				}`)

				So(resp.Body.String(), ShouldEqualJSON, `{
					"error": {
						"code": 115,
						"name": "AtomicOperationFailure",
						"message": "Atomic Operation rolled back due to one or more errors",
						"info": {
							"note/1": {
								"code": 10000,
								"message": "Original Sin",
								"name": "UnexpectedError"
							}
						}
					}
				}`)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeFalse)
				So(txDB.DidRollback, ShouldBeTrue)
			})

			Convey("commits deleted records", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) skyerr.Error {
					return nil
				})

				resp := r.POST(`{
					"ids": [
						"note/0",
						"note/1",
						"note/2"
					],
					"atomic": true
				}`)

				So(resp.Body.String(), ShouldEqualJSON, `{
					"result": [
						{"_type": "record", "_id": "note/0"},
						{"_type": "record", "_id": "note/1"},
						{"_type": "record", "_id": "note/2"}
					]
				}`)

				var record skydb.Record
				So(backingDB.Get(skydb.NewRecordID("record", "0"), &record), ShouldEqual, skydb.ErrRecordNotFound)
				So(backingDB.Get(skydb.NewRecordID("record", "1"), &record), ShouldEqual, skydb.ErrRecordNotFound)
				So(backingDB.Get(skydb.NewRecordID("record", "2"), &record), ShouldEqual, skydb.ErrRecordNotFound)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeTrue)
				So(txDB.DidRollback, ShouldBeFalse)
			})
		})
	})
}

func TestDeriveDeltaRecord(t *testing.T) {
	Convey("DeriveDeltaRecord", t, func() {
		Convey("set ACL when delta is non-nil", func() {
			acl := skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.WriteLevel),
			}

			dst := skydb.Record{}
			base := skydb.Record{
				ID:      skydb.NewRecordID("record", "id"),
				OwnerID: "user0",
			}
			delta := skydb.Record{
				ID:  skydb.NewRecordID("record", "id"),
				ACL: acl,
			}

			deriveDeltaRecord(&dst, &base, &delta)

			So(dst.ACL, ShouldResemble, acl)
		})

		Convey("preserve ACL when delta is nil", func() {
			acl := skydb.RecordACL{
				skydb.NewRecordACLEntryDirect("user0", skydb.WriteLevel),
			}

			dst := skydb.Record{}
			base := skydb.Record{
				ID:      skydb.NewRecordID("record", "id"),
				OwnerID: "user0",
				ACL:     acl,
			}
			delta := skydb.Record{
				ID: skydb.NewRecordID("record", "id"),
			}

			deriveDeltaRecord(&dst, &base, &delta)

			So(dst.ACL, ShouldResemble, acl)
		})
	})
}
