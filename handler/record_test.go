package handler

import (
	"bytes"
	"github.com/oursky/ourd/oddb/oddbtest"
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"

	"encoding/json"
	"errors"
	"reflect"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
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

func TestTransportRecordMarshalJSON(t *testing.T) {
	r := transportRecord{
		ID: oddb.NewRecordID("recordkey", "recordtype"),
		Data: map[string]interface{}{
			"stringkey": "stringvalue",
			"numkey":    1,
			"boolkey":   true,
		},
	}

	expectedMap := map[string]interface{}{
		"stringkey": "stringvalue",
		// NOTE(limouren): json unmarshal numbers to float64
		"numkey":  float64(1),
		"boolkey": true,
	}

	jsonBytes, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	// there is no guarantee key ordering in marshalled json,
	// so we compare the unmarshalled map
	marshalledMap := map[string]interface{}{}
	json.Unmarshal(jsonBytes, &marshalledMap)

	if !reflect.DeepEqual(marshalledMap, expectedMap) {
		t.Fatalf("got marshalledMap = %#v, expect %#v", marshalledMap, expectedMap)
	}
}

func TestTransportRecordUnmarshalJSON(t *testing.T) {
	jsonBytes := []byte(`{
		"_id": "recordtype/recordkey",
		"stringkey": "stringvalue",
		"numkey": 1,
		"boolkey": true}`)

	expectedRecord := transportRecord{
		ID: oddb.NewRecordID("recordtype", "recordkey"),
		Data: map[string]interface{}{
			"stringkey": "stringvalue",
			"numkey":    float64(1),
			"boolkey":   true,
		},
	}

	unmarshalledRecord := transportRecord{}
	if err := json.Unmarshal(jsonBytes, &unmarshalledRecord); err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(unmarshalledRecord, expectedRecord) {
		t.Fatalf("got unmarshalledRecord = %#v, expect %#v", unmarshalledRecord, expectedRecord)
	}
}

func TestResponseItemMarshal(t *testing.T) {
	record := transportRecord{
		ID:   oddb.NewRecordID("recordtype", "recordkey"),
		Data: map[string]interface{}{"key": "value"},
	}

	item := newResponseItem(&record)
	expectedJSON := []byte(`{"_id":"recordtype/recordkey","_type":"record","key":"value"}`)

	marshalledItem, err := json.Marshal(item)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(marshalledItem, expectedJSON) {
		t.Errorf("got marshalledItem = %s, want %s", marshalledItem, expectedJSON)
	}

	item = newResponseItemErr("recordtype/errorkey", oderr.NewUnknownErr(errors.New("a refreshing error")))
	expectedJSON = []byte(`{"_id":"recordtype/errorkey","_type":"error","type":"UnknownError","code":1,"message":"a refreshing error"}`)

	marshalledItem, err = json.Marshal(item)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(marshalledItem, expectedJSON) {
		t.Errorf("got marshalledItem = %s, want %s", marshalledItem, expectedJSON)
	}
}

func TestResponseItemMarshalEmpty(t *testing.T) {
	record := transportRecord{
		ID: oddb.NewRecordID("recordtype", "recordkey"),
	}
	item := newResponseItem(&record)
	expectedJSON := []byte(`{"_id":"recordtype/recordkey","_type":"record"}`)

	marshalled, err := json.Marshal(&item)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(marshalled, expectedJSON) {
		t.Errorf("got marshalled = %s, want %s", marshalled, expectedJSON)
	}

	record.Data = map[string]interface{}{}
	marshalled, err = json.Marshal(&item)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(marshalled, expectedJSON) {
		t.Errorf("got marshalled = %s, want %s", marshalled, expectedJSON)
	}

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

func TestRecordFetch(t *testing.T) {
	record1 := oddb.Record{ID: oddb.NewRecordID("type", "1")}
	record2 := oddb.Record{ID: oddb.NewRecordID("type", "2")}
	db := oddbtest.NewMapDB()
	db.Save(&record1)
	db.Save(&record2)

	Convey("Given a Database", t, func() {
		Convey("records can be fetched", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"ids": []interface{}{"type/1", "type/2"},
				},
				Database: db,
			}
			response := router.Response{}

			RecordFetchHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(response.Result, ShouldResemble, []responseItem{
				newResponseItem((*transportRecord)(&record1)),
				newResponseItem((*transportRecord)(&record2)),
			})
		})

		Convey("returns error in a list when non-exist records are fetched", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"ids": []interface{}{"type/1", "type/not-exist", "type/2"},
				},
				Database: db,
			}
			response := router.Response{}

			RecordFetchHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(response.Result, ShouldResemble, []responseItem{
				newResponseItem((*transportRecord)(&record1)),
				newResponseItemErr("type/not-exist", oderr.ErrRecordNotFound),
				newResponseItem((*transportRecord)(&record2)),
			})
		})

		Convey("returns error when non-string ids is supplied", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"ids": []interface{}{1, 2, 3},
				},
				Database: db,
			}
			response := router.Response{}

			RecordFetchHandler(&payload, &response)

			So(response.Result, ShouldBeNil)
			So(response.Err, ShouldResemble, oderr.NewRequestInvalidErr(errors.New("expected string id")))
		})
	})
}
