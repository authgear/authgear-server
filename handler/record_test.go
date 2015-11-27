package handler

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/oursky/skygear/authtoken"
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/hook"
	. "github.com/oursky/skygear/ourtest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbtest"
	. "github.com/smartystreets/goconvey/convey"
)

var ZeroTime time.Time

func TestRecordDeleteHandler(t *testing.T) {
	Convey("RecordDeleteHandler", t, func() {
		note0 := skydb.Record{
			ID: skydb.NewRecordID("note", "0"),
		}
		note1 := skydb.Record{
			ID: skydb.NewRecordID("note", "1"),
		}

		db := skydbtest.NewMapDB()
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
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordSaveHandler", t, func() {
		db := skydbtest.NewMapDB()
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
					"_access": null,
					"k1": "v1",
					"k2": "v2"
				}, {
					"_id": "type2/id2",
					"_type": "record",
					"_access": null,
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
	timeNow = func() time.Time { return ZeroTime }
	defer func() {
		timeNow = timeNowUTC
	}()

	Convey("RecordSaveHandler", t, func() {
		db := skydbtest.NewMapDB()
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

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
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

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"asset": &skydb.Asset{Name: "asset-name"},
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

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"ref": skydb.NewReference("type2", "id2"),
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

			record := skydb.Record{}
			So(db.Get(skydb.NewRecordID("type1", "id1"), &record), ShouldBeNil)
			So(record, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("type1", "id1"),
				Data: map[string]interface{}{
					"geo": skydb.NewLocation(1, 2),
				},
			})
		})
	})
}

type bogusFieldDatabase struct {
	SaveFunc func(record *skydb.Record) error
	skydb.Database
}

func (db bogusFieldDatabase) Extend(recordType string, schema skydb.RecordSchema) error {
	return nil
}

func (db bogusFieldDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	return nil
}

func (db bogusFieldDatabase) Save(record *skydb.Record) error {
	return db.SaveFunc(record)
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
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
			payload.Database = db
		})

		Convey("parse sequence field correctly", func() {
			db.SaveFunc = func(record *skydb.Record) error {
				So(record, ShouldResemble, &skydb.Record{
					ID:   skydb.NewRecordID("record", "id"),
					Data: skydb.Data{},
				})

				record.Data["seq"] = int64(1)

				return nil
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
					"seq": 1
				}]
			}`)
		})
	})
}

type noExtendDatabase struct {
	calledExtend bool
	skydb.Database
}

func (db *noExtendDatabase) Extend(recordType string, schema skydb.RecordSchema) error {
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
	lastquery *skydb.Query
	skydb.Database
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
	records []skydb.Record
	skydb.Database
}

func (db *queryResultsDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	return uint64(len(db.records)), nil
}

func (db *queryResultsDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	return skydb.NewRows(skydb.NewMemoryRows(db.records)), nil
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

		db := &queryResultsDatabase{}
		db.records = []skydb.Record{record1, record0, record2}

		r := handlertest.NewSingleRouteRouter(RecordQueryHandler, func(p *router.Payload) {
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
			So(db.lastquery, ShouldResemble, &skydb.Query{
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
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						KeyPath: "noteOrder",
						Order:   skydb.Desc,
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery, ShouldResemble, &skydb.Query{
				Type: "note",
				Sorts: []skydb.Sort{
					skydb.Sort{
						Func: &skydb.DistanceFunc{
							Field:    "location",
							Location: skydb.NewLocation(1, 2),
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(*db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
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
					},
				},
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(*db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(*db.lastquery.Predicate, ShouldResemble, skydb.Predicate{
				Operator: skydb.LessThanOrEqual,
				Children: []interface{}{
					skydb.Expression{
						skydb.Function,
						&skydb.DistanceFunc{
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.ComputedKeys, ShouldResemble, map[string]skydb.Expression{
				"distance": skydb.Expression{
					skydb.Function,
					&skydb.DistanceFunc{
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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.DesiredKeys, ShouldResemble, []string{"location"})
		})

		Convey("Return records when desired keys is empty", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type":  "note",
					"desired_keys": []interface{}{},
				},
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.DesiredKeys, ShouldResemble, []string{})
		})

		Convey("Return records when desired keys is nil", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"record_type":  "note",
					"desired_keys": nil,
				},
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

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
				Database: db,
			}
			response := router.Response{}

			RecordQueryHandler(&payload, &response)

			So(response.Err, ShouldBeNil)
			So(db.lastquery.GetCount, ShouldBeTrue)
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type singleRecordDatabase struct {
	record skydb.Record
	skydb.Database
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

func (db *singleRecordDatabase) Extend(recordType string, schema skydb.RecordSchema) error {
	return nil
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

func TestRecordMetaData(t *testing.T) {
	Convey("Record Meta Data", t, func() {
		db := skydbtest.NewMapDB()
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = timeNowUTC
		}()

		Convey("on a newly created record", func() {
			r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
				payload.Database = db
				payload.UserInfoID = "requestUserID"
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

			r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
				payload.Database = db
				payload.UserInfoID = "requestUserID"
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

func (s *urlOnlyAssetStore) SignedURL(name string, expiredAt time.Time) string {
	return fmt.Sprintf("http://skygear.test/asset/%s?expiredAt=1997-07-01T00:00:00", name)
}

func TestRecordAssetSerialization(t *testing.T) {
	Convey("RecordAssetSerialization", t, func() {
		db := skydbtest.NewMapDB()
		db.Save(&skydb.Record{
			ID: skydb.NewRecordID("record", "id"),
			Data: map[string]interface{}{
				"asset": &skydb.Asset{Name: "asset-name"},
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
						"$url": "http://skygear.test/asset/asset-name?expiredAt=1997-07-01T00:00:00"
					}
				}]
			}`)
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type referencedRecordDatabase struct {
	note     skydb.Record
	category skydb.Record
	skydb.Database
}

func (db *referencedRecordDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	switch id.String() {
	case "note/note1":
		*record = db.note
	case "category/important":
		*record = db.category
	}
	return nil
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

func (db *referencedRecordDatabase) Extend(recordType string, schema skydb.RecordSchema) error {
	return nil
}

func TestRecordQueryWithEagerLoad(t *testing.T) {
	Convey("Given a referenced record in DB", t, func() {
		db := &referencedRecordDatabase{
			note: skydb.Record{
				ID:      skydb.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": skydb.NewReference("category", "important"),
				},
			},
			category: skydb.Record{
				ID:      skydb.NewRecordID("category", "important"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"title": "This is important.",
				},
			},
		}

		injectDBFunc := func(payload *router.Payload) {
			payload.Database = db
		}

		Convey("query record with eager load", func() {
			resp := handlertest.NewSingleRouteRouter(RecordQueryHandler, injectDBFunc).POST(`{
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
					"_transient": {
						"category": {"_access":null,"_id":"category/important","_type":"record","_ownerID":"ownerID", "title": "This is important."}
					}
				}]
			}`)
		})

		Convey("return error if multiple key paths are specified", func() {
			resp := handlertest.NewSingleRouteRouter(RecordQueryHandler, injectDBFunc).POST(`{
				"record_type": "note",
				"include": {
					"category": {"$type": "keypath", "$val": "category"},
					"city": {"$type": "keypath", "$val": "city"}
				}
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code":101,
					"message":"eager loading for multiple keys is not supported",
					"type":"RequestInvalid"
				}
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

		db := &queryResultsDatabase{}
		db.records = []skydb.Record{record1, record0, record2}

		r := handlertest.NewSingleRouteRouter(RecordQueryHandler, func(p *router.Payload) {
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

type stackingHook struct {
	records         []*skydb.Record
	originalRecords []*skydb.Record
}

func (p *stackingHook) Func(record *skydb.Record, originalRecord *skydb.Record) error {
	p.records = append(p.records, record)
	p.originalRecords = append(p.originalRecords, originalRecord)
	return nil
}

type erroneousDB struct {
	skydb.Database
}

func (db erroneousDB) Extend(string, skydb.RecordSchema) error {
	return nil
}

func (db erroneousDB) Get(skydb.RecordID, *skydb.Record) error {
	return errors.New("erroneous save")
}

func (db erroneousDB) Save(*skydb.Record) error {
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

		record := &skydb.Record{
			ID: skydb.NewRecordID("record", "id"),
		}

		registry := hook.NewRegistry()
		beforeHook := stackingHook{}
		afterHook := stackingHook{}

		for _, test := range handlerTests {
			testName := fmt.Sprintf("executes Before%[1]s and After%[1]s action hooks", test.kind)
			Convey(testName, func() {
				registry.Register(test.beforeActionKind, "record", beforeHook.Func)
				registry.Register(test.afterActionKind, "record", afterHook.Func)

				db := skydbtest.NewMapDB()
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
		db := skydbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(p *router.Payload) {
			p.Database = db
			p.HookRegistry = registry
		})

		Convey("record is not saved if BeforeSave's hook returns an error", func() {
			registry.Register(hook.BeforeSave, "record", func(*skydb.Record, *skydb.Record) error {
				return errors.New("no hooks for you")
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
			registry.Register(hook.BeforeSave, "record", func(record *skydb.Record, originalRecord *skydb.Record) error {
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
			registry.Register(hook.BeforeSave, "record", func(record *skydb.Record, originalRecord *skydb.Record) error {
				called = true
				So(*record, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("record", "id"),
					Data: map[string]interface{}{
						"new": true,
					},
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

// mockTxDB implements and records TxDatabase's methods and delegates other
// calls to underlying Database
type mockTxDatabase struct {
	DidBegin, DidCommit, DidRollback bool
	skydb.Database
}

func newMockTxDatabase(backingDB skydb.Database) *mockTxDatabase {
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

var _ skydb.TxDatabase = &mockTxDatabase{}

type filterFuncDef func(op string, recordID skydb.RecordID, record *skydb.Record) error

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
		backingDB := skydbtest.NewMapDB()
		txDB := newMockTxDatabase(backingDB)
		db := newSelectiveDatabase(txDB)

		Convey("for RecordSaveHandler", func() {
			r := handlertest.NewSingleRouteRouter(RecordSaveHandler, func(payload *router.Payload) {
				payload.Database = db
			})

			Convey("rolls back saved records on error", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) error {
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
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) error {
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

				var record skydb.Record
				So(backingDB.Get(skydb.NewRecordID("note", "0"), &record), ShouldBeNil)
				So(record, ShouldResemble, skydb.Record{
					ID:   skydb.NewRecordID("note", "0"),
					Data: map[string]interface{}{},
				})
				So(backingDB.Get(skydb.NewRecordID("note", "1"), &record), ShouldBeNil)
				So(record, ShouldResemble, skydb.Record{
					ID:   skydb.NewRecordID("note", "1"),
					Data: map[string]interface{}{},
				})

				So(txDB.DidBegin, ShouldBeTrue)
				So(txDB.DidCommit, ShouldBeTrue)
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

			r := handlertest.NewSingleRouteRouter(RecordDeleteHandler, func(payload *router.Payload) {
				payload.Database = db
			})

			Convey("rolls back deleted records on error", func() {
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) error {
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
				db.SetFilter(func(op string, recordID skydb.RecordID, record *skydb.Record) error {
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
