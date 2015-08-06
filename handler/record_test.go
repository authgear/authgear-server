package handler

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/handler/handlertest"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbtest"
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRecordDeleteHandler(t *testing.T) {
	Convey("RecordDeleteHandler", t, func() {
		note0 := oddb.Record{
			ID: oddb.NewRecordID("note", "0"),
		}
		note1 := oddb.Record{
			ID: oddb.NewRecordID("note", "1"),
		}

		db := oddbtest.NewMapDB()
		So(db.Save(&note0), ShouldBeNil)
		So(db.Save(&note1), ShouldBeNil)

		router := handlertest.NewSingleRouteRouter(RecordDeleteHandler, func(p *router.Payload) {
			p.Database = db
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
		{"_id": "note/notexistid", "_type": "error", "code": 103, "message": "record not found", "type": "ResourceNotFound"}
	]
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
	Convey("RecordSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
			payload.Database = db
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
					"_access":null,
					"k1": "v1",
					"k2": "v2"
				}, {
					"_id": "type2/id2",
					"_type": "record",
					"_access":null,
					"k3": "v3",
					"k4": "v4"
				}]
			}`)
		})

		Convey("Update existing record", func() {
			record := oddb.Record{
				ID: oddb.NewRecordID("record", "id"),
				Data: map[string]interface{}{
					"existing": "YES",
					"old":      true,
				},
			}
			So(db.Save(&record), ShouldBeNil)

			resp := r.POST(`{
				"records": [{
					"_id": "record/id",
					"old": false,
					"new": 1
				}]
			}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "record/id",
					"_type": "record",
					"_access": null,
					"existing": "YES",
					"old": false,
					"new": 1
				}]
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
					"floatkey": 1
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
					"type": "RequestInvalid",
					"code": 101,
					"message": "record: required field \"_id\" not found"
				},{
					"_type": "error",
					"type": "RequestInvalid",
					"code": 101,
					"message": "record: \"_id\" should be of format '{type}/{id}', got \"invalidkey\""
			}]}`)
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
					"type": "RequestInvalid",
					"code": 101,
					"message": "record: required field \"_id\" not found"
				},{
					"_type": "error",
					"type": "RequestInvalid",
					"code": 101,
					"message": "record: \"_id\" should be of format '{type}/{id}', got \"invalidkey\""
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
					"_access": null
				}]
			}`)
		})
	})
}

func TestRecordSaveDataType(t *testing.T) {
	Convey("RecordSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(p *router.Payload) {
			p.Database = db
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
		"date_value": {"$type": "date", "$date": "2015-04-10T09:35:20Z"}
	}]
}`)

			record := oddb.Record{}
			So(db.Get(oddb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"date_value": time.Date(2015, 4, 10, 9, 35, 20, 0, time.UTC),
				},
			})
		})

		Convey("Parses Asset", func() {
			resp := r.POST(`{
	"records": [{
		"_id": "type1/id1",
		"asset": {"$type": "asset", "$name": "asset-name"}
	}]
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"_id": "type1/id1",
		"_type": "record",
		"_access": null,
		"asset": {"$type": "asset", "$name": "asset-name"}
	}]
}`)

			record := oddb.Record{}
			So(db.Get(oddb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"asset": oddb.Asset{Name: "asset-name"},
				},
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
		"ref": {"$type": "ref", "$id": "type2/id2"}
	}]
}`)

			record := oddb.Record{}
			So(db.Get(oddb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"ref": oddb.NewReference("type2", "id2"),
				},
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
		"geo": {"$type": "geo", "$lng": 1, "$lat": 2}
	}]
}`)

			record := oddb.Record{}
			So(db.Get(oddb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"geo": oddb.NewLocation(1, 2),
				},
			})
		})
	})
}

type noExtendDatabase struct {
	calledExtend bool
	oddb.Database
}

func (db *noExtendDatabase) Extend(recordType string, schema oddb.RecordSchema) error {
	db.calledExtend = true
	return errors.New("You shalt not call Extend")
}

func TestRecordSaveNoExtendIfRecordMalformed(t *testing.T) {
	Convey("RecordSaveHandler", t, func() {
		noExtendDB := &noExtendDatabase{}
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
			payload.Database = noExtendDB
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
	lastquery *oddb.Query
	oddb.Database
}

func (db *queryDatabase) Query(query *oddb.Query) (*oddb.Rows, error) {
	db.lastquery = query
	return oddb.EmptyRows, nil
}

func TestRecordQuery(t *testing.T) {
	Convey("Given a Database", t, func() {
		db := &queryDatabase{}
		Convey("Queries records with type", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type": "note",
				},
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &oddb.Query{
				Type: "note",
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &oddb.Query{
				Type: "note",
				Sorts: []oddb.Sort{
					oddb.Sort{
						KeyPath: "noteOrder",
						Order:   oddb.Desc,
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(*db.lastquery.Predicate, ShouldResemble, oddb.Predicate{
				Operator: oddb.Equal,
				Children: []interface{}{
					oddb.Expression{oddb.KeyPath, "noteOrder"},
					oddb.Expression{oddb.Literal, float64(1)},
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
					},
				},
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(*db.lastquery.Predicate, ShouldResemble, oddb.Predicate{
				Operator: oddb.And,
				Children: []interface{}{
					oddb.Predicate{
						Operator: oddb.Equal,
						Children: []interface{}{
							oddb.Expression{oddb.KeyPath, "content"},
							oddb.Expression{oddb.Literal, "text"},
						},
					},
					oddb.Predicate{
						Operator: oddb.GreaterThan,
						Children: []interface{}{
							oddb.Expression{oddb.KeyPath, "noteOrder"},
							oddb.Expression{oddb.Literal, float64(1)},
						},
					},
				},
			})
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type singleRecordDatabase struct {
	record oddb.Record
	oddb.Database
}

func (db *singleRecordDatabase) Get(id oddb.RecordID, record *oddb.Record) error {
	*record = db.record
	return nil
}

func (db *singleRecordDatabase) Save(record *oddb.Record) error {
	*record = db.record
	return nil
}

func (db *singleRecordDatabase) Query(query *oddb.Query) (*oddb.Rows, error) {
	return oddb.NewRows(oddb.NewMemoryRows([]oddb.Record{db.record})), nil
}

func (db *singleRecordDatabase) Extend(recordType string, schema oddb.RecordSchema) error {
	return nil
}

func TestRecordOwnerIDSerialization(t *testing.T) {
	Convey("Given a record with owner id in DB", t, func() {
		record := oddb.Record{
			ID:      oddb.NewRecordID("type", "id"),
			OwnerID: "ownerID",
		}
		db := &singleRecordDatabase{
			record: record,
		}

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
		}

		Convey("fetched record serializes owner id correctly", func() {
			resp := handlertest.NewSingleRouteRouter(RecordFetchHandler, injectDBFunc).POST(`{
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
			resp := handlertest.NewSingleRouteRouter(RecordSaveHandler, injectDBFunc).POST(`{
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
			resp := handlertest.NewSingleRouteRouter(RecordQueryHandler, injectDBFunc).POST(`{
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

type urlOnlyAssetStore struct{}

func (s *urlOnlyAssetStore) GetFileReader(name string) (io.ReadCloser, error) {
	panic("not implemented")
}

func (s *urlOnlyAssetStore) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	panic("not implemented")
}

func (s *urlOnlyAssetStore) SignedURL(name string, expiredAt time.Time) string {
	return fmt.Sprintf("http://ourd.test/asset/%s?expiredAt=1997-07-01T00:00:00", name)
}

func TestRecordAssetSerialization(t *testing.T) {
	Convey("RecordAssetSerialization", t, func() {
		db := oddbtest.NewMapDB()
		db.Save(&oddb.Record{
			ID: oddb.NewRecordID("record", "id"),
			Data: map[string]interface{}{
				"asset": oddb.Asset{Name: "asset-name"},
			},
		})

		assetStore := &urlOnlyAssetStore{}

		r := handlertest.NewSingleRouteRouter(RecordFetchHandler, func(p *router.Payload) {
			p.Database = db
			p.AssetStore = assetStore
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
						"$url": "http://ourd.test/asset/asset-name?expiredAt=1997-07-01T00:00:00"
					}
				}]
			}`)
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type referencedRecordDatabase struct {
	note     oddb.Record
	category oddb.Record
	oddb.Database
}

func (db *referencedRecordDatabase) Get(id oddb.RecordID, record *oddb.Record) error {
	switch id.String() {
	case "note/note1":
		*record = db.note
	case "category/important":
		*record = db.category
	}
	return nil
}

func (db *referencedRecordDatabase) Save(record *oddb.Record) error {
	return nil
}

func (db *referencedRecordDatabase) Query(query *oddb.Query) (*oddb.Rows, error) {
	return oddb.NewRows(oddb.NewMemoryRows([]oddb.Record{db.note})), nil
}

func (db *referencedRecordDatabase) Extend(recordType string, schema oddb.RecordSchema) error {
	return nil
}

func TestRecordQueryWithEagerLoad(t *testing.T) {
	Convey("Given a referenced record in DB", t, func() {
		db := &referencedRecordDatabase{
			note: oddb.Record{
				ID:      oddb.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": oddb.NewReference("category", "important"),
				},
			},
			category: oddb.Record{
				ID:      oddb.NewRecordID("category", "important"),
				OwnerID: "ownerID",
			},
		}

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
		}

		Convey("query record with eager load", func() {
			resp := handlertest.NewSingleRouteRouter(RecordQueryHandler, injectDBFunc).POST(`{
				"record_type": "note",
				"eager": [{"$type": "keypath", "$val": "category"}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "note/note1",
					"_type": "record",
					"_access": null,
					"_ownerID": "ownerID",
					"category": {"$id":"category/important","$type":"ref"}
				}],
				"other_result": {"eager_load":[
				{"_access":null,"_id":"category/important","_type":"record","_ownerID":"ownerID"}
				]}
			}`)
		})
	})
}

type stackingHook struct {
	records         []*oddb.Record
	originalRecords []*oddb.Record
}

func (p *stackingHook) Func(record *oddb.Record, originalRecord *oddb.Record) error {
	p.records = append(p.records, record)
	p.originalRecords = append(p.originalRecords, originalRecord)
	return nil
}

type erroneousDB struct {
	oddb.Database
}

func (db erroneousDB) Extend(string, oddb.RecordSchema) error {
	return nil
}

func (db erroneousDB) Get(oddb.RecordID, *oddb.Record) error {
	return errors.New("erroneous save")
}

func (db erroneousDB) Save(*oddb.Record) error {
	return errors.New("erroneous save")
}

func TestHookExecution(t *testing.T) {
	Convey("Record(Save|Delete)Handler", t, func() {
		handlerTests := []struct {
			kind             string
			handler          func(*router.Payload, *router.Response)
			beforeActionKind hook.Kind
			afterActionKind  hook.Kind
			reqBody          string
		}{
			{
				"Save",
				RecordSaveHandler,
				hook.BeforeSave,
				hook.AfterSave,
				`{"records": [{"_id": "record/id"}]}`,
			},
			{
				"Delete",
				RecordDeleteHandler,
				hook.BeforeDelete,
				hook.AfterDelete,
				`{"ids": ["record/id"]}`,
			},
		}

		record := &oddb.Record{
			ID: oddb.NewRecordID("record", "id"),
		}

		registry := hook.NewRegistry()
		beforeHook := stackingHook{}
		afterHook := stackingHook{}

		for _, test := range handlerTests {
			testName := fmt.Sprintf("executes Before%[1]s and After%[1]s action hooks", test.kind)
			Convey(testName, func() {
				registry.Register(test.beforeActionKind, "record", beforeHook.Func)
				registry.Register(test.afterActionKind, "record", afterHook.Func)

				db := oddbtest.NewMapDB()
				So(db.Save(record), ShouldBeNil)

				r := handlertest.NewSingleRouteRouter(test.handler, func(p *router.Payload) {
					p.Database = db
					p.HookRegistry = registry
				})

				r.POST(test.reqBody)

				So(len(beforeHook.records), ShouldEqual, 1)
				So(beforeHook.records[0].ID, ShouldResemble, record.ID)
				So(len(afterHook.records), ShouldEqual, 1)
				So(afterHook.records[0].ID, ShouldResemble, record.ID)
			})

			testName = fmt.Sprintf("doesn't execute After%[1]s hooks if db.%[1]s returns an error", test.kind)
			Convey(testName, func() {
				registry.Register(test.afterActionKind, "record", afterHook.Func)
				r := handlertest.NewSingleRouteRouter(test.handler, func(p *router.Payload) {
					p.Database = erroneousDB{}
					p.HookRegistry = registry
				})

				r.POST(test.reqBody)
				So(afterHook.records, ShouldBeEmpty)
			})
		}
	})

	Convey("HookRegistry", t, func() {
		registry := hook.NewRegistry()
		db := oddbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(p *router.Payload) {
			p.Database = db
			p.HookRegistry = registry
		})

		Convey("record is not saved if BeforeSave's hook returns an error", func() {
			registry.Register(hook.BeforeSave, "record", func(*oddb.Record, *oddb.Record) error {
				return errors.New("no hooks for you!")
			})
			r.POST(`{
				"records": [{
					"_id": "record/id"
				}]
			}`)

			var record oddb.Record
			So(db.Get(oddb.NewRecordID("record", "id"), &record), ShouldEqual, oddb.ErrRecordNotFound)
		})

		Convey("BeforeSave should be fed fully fetched record", func() {
			existingRecord := oddb.Record{
				ID: oddb.NewRecordID("record", "id"),
				Data: map[string]interface{}{
					"old": true,
				},
			}
			So(db.Save(&existingRecord), ShouldBeNil)

			called := false
			registry.Register(hook.BeforeSave, "record", func(record *oddb.Record, originalRecord *oddb.Record) error {
				called = true
				So(*record, ShouldResemble, oddb.Record{
					ID: oddb.NewRecordID("record", "id"),
					Data: map[string]interface{}{
						"old": true,
						"new": true,
					},
				})
				So(*originalRecord, ShouldResemble, oddb.Record{
					ID: oddb.NewRecordID("record", "id"),
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
	})
}

// mockTxDB implements and records TxDatabase's methods and delegates other
// calls to underlying Database
type mockTxDatabase struct {
	DidBegin, DidCommit, DidRollback bool
	oddb.Database
}

func newMockTxDatabase(backingDB oddb.Database) *mockTxDatabase {
	return &mockTxDatabase{Database: backingDB}
}

func (db *mockTxDatabase) Begin() error {
	db.DidBegin = true
	return nil
}

func (db *mockTxDatabase) Commit() error {
	db.DidCommit = true
	return nil
}

func (db *mockTxDatabase) Rollback() error {
	db.DidRollback = true
	return nil
}

var _ oddb.TxDatabase = &mockTxDatabase{}

type filterFuncDef func(op string, recordID oddb.RecordID, record *oddb.Record) error

// selectiveDatabase filter Get, Save and Delete by executing filterFunc
// if filterFunc return nil, the operation is delegated to underlying Database
// otherwise, the error is returned directly
type selectiveDatabase struct {
	filterFunc filterFuncDef
	oddb.Database
}

func newSelectiveDatabase(backingDB oddb.Database) *selectiveDatabase {
	return &selectiveDatabase{
		Database: backingDB,
	}
}

func (db *selectiveDatabase) SetFilter(filterFunc filterFuncDef) {
	db.filterFunc = filterFunc
}

func (db *selectiveDatabase) Get(id oddb.RecordID, record *oddb.Record) error {
	if err := db.filterFunc("GET", id, nil); err != nil {
		return err
	}

	return db.Database.Get(id, record)
}

func (db *selectiveDatabase) Save(record *oddb.Record) error {
	if err := db.filterFunc("SAVE", record.ID, record); err != nil {
		return err
	}

	return db.Database.Save(record)
}

func (db *selectiveDatabase) Delete(id oddb.RecordID) error {
	if err := db.filterFunc("DELETE", id, nil); err != nil {
		return err
	}

	return db.Database.Delete(id)
}

func (db *selectiveDatabase) Begin() error {
	return db.Database.(oddb.TxDatabase).Begin()
}

func (db *selectiveDatabase) Commit() error {
	return db.Database.(oddb.TxDatabase).Commit()
}

func (db *selectiveDatabase) Rollback() error {
	return db.Database.(oddb.TxDatabase).Rollback()
}

func TestAtomicOperation(t *testing.T) {
	Convey("Atomic Operation", t, func() {
		backingDB := oddbtest.NewMapDB()
		txDB := newMockTxDatabase(backingDB)
		db := newSelectiveDatabase(txDB)

		Convey("for RecordSaveHandler", func() {
			r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
				payload.Database = db
			})

			Convey("rolls back saved records on error", func() {
				db.SetFilter(func(op string, recordID oddb.RecordID, record *oddb.Record) error {
					if op == "SAVE" && recordID.Key == "1" {
						return errors.New("Original Sin")
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
						"type": "DatabaseError",
						"code": 666,
						"message": "Atomic Operation rolled back due to one or more errors",
						"info": {
							"note/1": "Original Sin"
						}
					}
				}`)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeFalse)
				So(txDB.DidRollback, ShouldBeTrue)
			})

			Convey("commit saved records when there are no errors", func() {
				db.SetFilter(func(op string, recordID oddb.RecordID, record *oddb.Record) error {
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
							"_access": null
						}, {
							"_id": "note/1",
							"_type": "record",
							"_access": null
						}]
				}`)

				var record oddb.Record
				So(backingDB.Get(oddb.NewRecordID("note", "0"), &record), ShouldBeNil)
				So(record, ShouldResemble, oddb.Record{
					ID:   oddb.NewRecordID("note", "0"),
					Data: map[string]interface{}{},
				})
				So(backingDB.Get(oddb.NewRecordID("note", "1"), &record), ShouldBeNil)
				So(record, ShouldResemble, oddb.Record{
					ID:   oddb.NewRecordID("note", "1"),
					Data: map[string]interface{}{},
				})

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeTrue)
				So(txDB.DidRollback, ShouldBeFalse)
			})
		})

		Convey("for RecordDeleteHandler", func() {
			So(backingDB.Save(&oddb.Record{
				ID: oddb.NewRecordID("note", "0"),
			}), ShouldBeNil)
			So(backingDB.Save(&oddb.Record{
				ID: oddb.NewRecordID("note", "1"),
			}), ShouldBeNil)
			So(backingDB.Save(&oddb.Record{
				ID: oddb.NewRecordID("note", "2"),
			}), ShouldBeNil)

			r := handlertest.NewSingleRouteRouter(RecordDeleteHandler, func(payload *router.Payload) {
				payload.Database = db
			})

			Convey("rolls back deleted records on error", func() {
				db.SetFilter(func(op string, recordID oddb.RecordID, record *oddb.Record) error {
					if op == "DELETE" && recordID.Key == "1" {
						return errors.New("Original Sin")
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
						"type": "DatabaseError",
						"code": 666,
						"message": "Atomic Operation rolled back due to one or more errors",
						"info": {
							"note/1": "Original Sin"
						}
					}
				}`)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeFalse)
				So(txDB.DidRollback, ShouldBeTrue)
			})

			Convey("commits deleted records", func() {
				db.SetFilter(func(op string, recordID oddb.RecordID, record *oddb.Record) error {
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

				var record oddb.Record
				So(backingDB.Get(oddb.NewRecordID("record", "0"), &record), ShouldEqual, oddb.ErrRecordNotFound)
				So(backingDB.Get(oddb.NewRecordID("record", "1"), &record), ShouldEqual, oddb.ErrRecordNotFound)
				So(backingDB.Get(oddb.NewRecordID("record", "2"), &record), ShouldEqual, oddb.ErrRecordNotFound)

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeTrue)
				So(txDB.DidRollback, ShouldBeFalse)
			})
		})
	})
}
