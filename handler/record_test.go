package handler

import (
	"errors"
	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbtest"
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestRecordFetch(t *testing.T) {
	Convey("FetchRecordHandler", t, func() {
		record1 := oddb.Record{ID: oddb.NewRecordID("type", "1")}
		record2 := oddb.Record{ID: oddb.NewRecordID("type", "2")}
		db := oddbtest.NewMapDB()
		So(db.Save(&record1), ShouldBeNil)
		So(db.Save(&record2), ShouldBeNil)

		r := newSingleRouteRouter(RecordFetchHandler, func(payload *router.Payload) {
			payload.Database = db
		})

		Convey("fetches multiple records", func() {
			resp := r.POST(`{
				"ids": ["type/1", "type/2"]
				}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/1",
					"_type": "record"
				}, {
					"_id": "type/2",
					"_type": "record"
				}]
			}`)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("returns error in a list when non-exist records are fetched", func() {
			resp := r.POST(`{
				"ids": ["type/1", "type/not-exist", "type/2"]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/1",
					"_type": "record"
				}, {
					"_id": "type/not-exist",
					"_type": "error",
					"type": "ResourceNotFound",
					"code": 103,
					"message": "record not found"
				}, {
					"_id": "type/2",
					"_type": "record"
				}]
			}`)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("returns error when non-string ids is supplied", func() {
			resp := r.POST(`{
				"ids": [1, 2, 3]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error":{
					"code": 101,
					"type": "RequestInvalid",
					"message": "expected string id"
				}
			}`)
		})
	})
}

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

		router := newSingleRouteRouter(RecordDeleteHandler, func(p *router.Payload) {
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

// TODO(limouren): refactor TokenStores commonly used in testing to
// a separate package

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
		r := newSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
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
					"k1": "v1",
					"k2": "v2"
				}, {
					"_id": "type2/id2",
					"_type": "record",
					"k3": "v3",
					"k4": "v4"
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
	})
}

func TestRecordSaveDataType(t *testing.T) {
	Convey("RecordSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		r := newSingleRouteRouter(RecordSaveHandler, func(p *router.Payload) {
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
		r := newSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
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
			resp := newSingleRouteRouter(RecordFetchHandler, injectDBFunc).POST(`{
				"ids": ["do/notCare"]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_ownerID": "ownerID"
				}]
			}`)
		})

		Convey("saved record serializes owner id correctly", func() {
			resp := newSingleRouteRouter(RecordSaveHandler, injectDBFunc).POST(`{
				"records": [{
					"_id": "do/notCare"
				}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_ownerID": "ownerID"
				}]
			}`)
		})

		Convey("queried record serializes owner id correctly", func() {
			resp := newSingleRouteRouter(RecordQueryHandler, injectDBFunc).POST(`{
				"record_type": "doNotCare"
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
					"_id": "type/id",
					"_type": "record",
					"_ownerID": "ownerID"
				}]
			}`)
		})
	})
}
