package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRecordQuery(t *testing.T) {
	getRecordStore := func() (recordStore *queryRecordStore) {
		return &queryRecordStore{
			Store: record.NewMockStore(),
		}
	}

	Convey("Test QueryHandler", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStore()
		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		qh.Logger = logging.LoggerEntry("handler")
		// TODO:
		qh.AssetStore = nil
		qh.TxContext = db.NewMockTxContext()

		Convey("Queries records with type", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note"
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery, ShouldResemble, &record.Query{
				Type: "note",
			})
			So(recordStore.lastAccessControlOptions, ShouldResemble, &record.AccessControlOptions{
				ViewAsUser: &authinfo.AuthInfo{
					ID:         "faseng.cat.id",
					Roles:      []string{"user"},
					Verified:   true,
					VerifyInfo: map[string]bool{},
				},
			})
		})

		Convey("Queries records with sorting", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"sort": [
						[{"$val": "noteOrder", "$type": "keypath"}, "desc"]
					]
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery, ShouldResemble, &record.Query{
				Type: "note",
				Sorts: []record.Sort{
					record.Sort{
						Expression: record.Expression{
							Type:  record.KeyPath,
							Value: "noteOrder",
						},
						Order: record.Desc,
					},
				},
			})
		})

		Convey("Queries records with sorting by distance function", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"sort": [
						[
							["func", "distance", {"$type": "keypath", "$val": "location"}, {"$type":"geo", "$lng": 1.0, "$lat": 2.0}],
							"desc"
						]
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery, ShouldResemble, &record.Query{
				Type: "note",
				Sorts: []record.Sort{
					record.Sort{
						Expression: record.Expression{
							Type: record.Function,
							Value: record.DistanceFunc{
								Field:    "location",
								Location: record.NewLocation(1, 2),
							},
						},
						Order: record.Desc,
					},
				},
			})
		})

		Convey("Queries records with predicate", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"eq", {"$type": "keypath", "$val": "noteOrder"}, 1.0
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.Predicate, ShouldResemble, record.Predicate{
				Operator: record.Equal,
				Children: []interface{}{
					record.Expression{record.KeyPath, "noteOrder"},
					record.Expression{record.Literal, float64(1)},
				},
			})
		})

		Convey("Queries records with complex predicate", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"and",
						["eq", {"$type": "keypath", "$val": "content"}, "text"],
						["gt", {"$type": "keypath", "$val": "noteOrder"}, 1.0],
						["neq", {"$type": "keypath", "$val": "content"}, null]
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.Predicate, ShouldResemble, record.Predicate{
				Operator: record.And,
				Children: []interface{}{
					record.Predicate{
						Operator: record.Equal,
						Children: []interface{}{
							record.Expression{record.KeyPath, "content"},
							record.Expression{record.Literal, "text"},
						},
					},
					record.Predicate{
						Operator: record.GreaterThan,
						Children: []interface{}{
							record.Expression{record.KeyPath, "noteOrder"},
							record.Expression{record.Literal, float64(1)},
						},
					},
					record.Predicate{
						Operator: record.NotEqual,
						Children: []interface{}{
							record.Expression{record.KeyPath, "content"},
							record.Expression{record.Literal, nil},
						},
					},
				},
			})
		})

		Convey("Queries records by distance func", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"lte",
						["func", "distance", {"$type": "keypath", "$val": "location"}, {"$type": "geo", "$lng": 1.0, "$lat": 2.0}],
						500.0
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.Predicate, ShouldResemble, record.Predicate{
				Operator: record.LessThanOrEqual,
				Children: []interface{}{
					record.Expression{
						record.Function,
						record.DistanceFunc{
							Field:    "location",
							Location: record.NewLocation(1, 2),
						},
					},
					record.Expression{record.Literal, float64(500)},
				},
			})
		})

		Convey("Return calculated distance", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {
						"distance": [
							"func",
							"distance",
							{"$type": "keypath", "$val": "location"},
							{"$type": "geo", "$lng": 1.0, "$lat": 2.0}
						]
					}
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.ComputedKeys, ShouldResemble, map[string]record.Expression{
				"distance": record.Expression{
					record.Function,
					record.DistanceFunc{
						Field:    "location",
						Location: record.NewLocation(1, 2),
					},
				},
			})
		})

		Convey("Return records with desired keys only", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"desired_keys": ["location"]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.DesiredKeys, ShouldResemble, []string{"location"})
		})

		Convey("Return records when desired keys is empty", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"desired_keys": []
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.DesiredKeys, ShouldResemble, []string{})
		})

		Convey("Return records when desired keys is nil", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"desired_keys": null
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.DesiredKeys, ShouldBeNil)
		})

		Convey("Queries records with offset", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"limit": 200.0,
					"offset": 400.0
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery.Limit, ShouldNotBeNil)
			So(*recordStore.lastquery.Limit, ShouldEqual, 200)
			So(recordStore.lastquery.Offset, ShouldEqual, 400)
		})

		Convey("Queries records with count", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"count": true
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[], "info": {"count": 0}}}`)
			So(recordStore.lastquery.GetCount, ShouldBeTrue)
		})

		Convey("Propagate invalid query error", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"eq", {"$type": "keypath", "$val": "content"}, {}
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			var result interface{}
			json.Unmarshal(resp.Body.Bytes(), &result)
			actualJSON, _ := result.(map[string]interface{})
			So(actualJSON["error"], ShouldNotBeNil)
		})

		Convey("Queries records with type and master key", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note"
				}
			`))
			resp := httptest.NewRecorder()

			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			qh.AuthContext = auth.NewMockContextGetterWithMasterkeyDefaultUser()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
			So(recordStore.lastquery, ShouldResemble, &record.Query{
				Type: "note",
			})
			So(recordStore.lastAccessControlOptions, ShouldResemble, &record.AccessControlOptions{
				ViewAsUser: &authinfo.AuthInfo{
					ID:         "faseng.cat.id",
					Roles:      []string{"user"},
					Verified:   true,
					VerifyInfo: map[string]bool{},
				},
				BypassAccessControl: true,
			})
		})
	})

	Convey("Test QueryHandler with Field ACL", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStore()
		publicRole := record.FieldUserRole{record.PublicFieldUserRoleType, ""}
		recordStore.SetRecordFieldAccess(record.NewFieldACL(record.FieldACLEntryList{
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

		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		qh.Logger = logging.LoggerEntry("handler")
		// TODO:
		qh.AssetStore = nil
		qh.TxContext = db.NewMockTxContext()

		Convey("should block non-comparable, non-discoverable field", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"gt", {"$type": "keypath", "$val": "index"}, 1.0
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			var result interface{}
			json.Unmarshal(resp.Body.Bytes(), &result)
			actualJSON, _ := result.(map[string]interface{})
			So(actualJSON["error"], ShouldNotBeNil)
			actualJSON, _ = actualJSON["error"].(map[string]interface{})
			So(actualJSON["code"], ShouldEqual, skyerr.RecordQueryDenied)
		})

		Convey("should not block non-comparable, non-discoverable field with master key", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"gt", {"$type": "keypath", "$val": "index"}, 1.0
					]
				}
			`))

			resp := httptest.NewRecorder()
			qh.AuthContext = auth.NewMockContextGetterWithMasterkeyDefaultUser()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
		})

		Convey("should block non-comparable", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"gt", {"$type": "keypath", "$val": "title"}, "Tale of Two Cities"
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			var result interface{}
			json.Unmarshal(resp.Body.Bytes(), &result)
			actualJSON, _ := result.(map[string]interface{})
			So(actualJSON["error"], ShouldNotBeNil)
			actualJSON, _ = actualJSON["error"].(map[string]interface{})
			So(actualJSON["code"], ShouldEqual, skyerr.RecordQueryDenied)
		})

		Convey("should allow comparable field with equality", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"predicate": [
						"eq", {"$type": "keypath", "$val": "title"}, "Tale of Two Cities"
					]
				}
			`))

			resp := httptest.NewRecorder()
			qh.AuthContext = auth.NewMockContextGetterWithMasterkeyDefaultUser()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"records":[]}}`)
		})

		Convey("should block non-comparable field in sort", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"sort": [
						[{"$type": "keypath", "$val": "title"}, "desc"]
					]
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			var result interface{}
			json.Unmarshal(resp.Body.Bytes(), &result)
			actualJSON, _ := result.(map[string]interface{})
			So(actualJSON["error"], ShouldNotBeNil)
			actualJSON, _ = actualJSON["error"].(map[string]interface{})
			So(actualJSON["code"], ShouldEqual, skyerr.RecordQueryDenied)
		})

		Convey("should block non-readable field in transient include", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {
						"category": {"$type": "keypath", "$val": "category"}
					}
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			var result interface{}
			json.Unmarshal(resp.Body.Bytes(), &result)
			actualJSON, _ := result.(map[string]interface{})
			So(actualJSON["error"], ShouldNotBeNil)
			actualJSON, _ = actualJSON["error"].(map[string]interface{})
			So(actualJSON["code"], ShouldEqual, skyerr.RecordQueryDenied)
		})
	})
}

type queryRecordStore struct {
	lastquery                *record.Query
	lastAccessControlOptions *record.AccessControlOptions
	record.Store
}

func (s *queryRecordStore) QueryCount(query *record.Query, accessControlOptions *record.AccessControlOptions) (uint64, error) {
	s.lastquery = query
	s.lastAccessControlOptions = accessControlOptions
	return 0, nil
}

func (s *queryRecordStore) Query(query *record.Query, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	s.lastquery = query
	s.lastAccessControlOptions = accessControlOptions
	return record.EmptyRows, nil
}

func TestRecordQueryResults(t *testing.T) {
	getRecordStore := func() (recordStore *queryResultsRecordStore) {
		return &queryResultsRecordStore{
			records: []record.Record{
				record.Record{
					ID: record.NewRecordID("note", "1"),
				},
				record.Record{
					ID: record.NewRecordID("note", "0"),
				},
				record.Record{
					ID: record.NewRecordID("note", "2"),
				},
			},
			MockStore: record.NewMockStore(),
		}
	}

	Convey("Test QueryHandler query result", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStore()
		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		qh.Logger = logging.LoggerEntry("handler")
		qh.TxContext = db.NewMockTxContext()

		Convey("REGRESSION #227: query returns correct results from db", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note"
				}
			`))

			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "1",
						"_access": null
					},
					{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "0",
						"_access": null
					},
					{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "2",
						"_access": null
					}]
				}
			}`)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}

type queryResultsRecordStore struct {
	records []record.Record
	typemap map[string]record.Schema
	*record.MockStore
}

func (s *queryResultsRecordStore) QueryCount(query *record.Query, accessControlOptions *record.AccessControlOptions) (uint64, error) {
	return uint64(len(s.records)), nil
}

func (s *queryResultsRecordStore) Query(query *record.Query, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	return record.NewRows(record.NewMemoryRows(s.records)), nil
}

func (s *queryResultsRecordStore) GetSchema(recordType string) (record.Schema, error) {
	return s.typemap[recordType], nil
}

func TestRecordQueryWithEagerLoad(t *testing.T) {
	getRecordStore := func() (recordStore *referencedRecordStore) {
		return &referencedRecordStore{
			note: record.Record{
				ID:      record.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": record.NewReference("category", "important"),
					"city":     record.NewReference("city", "beautiful"),
					"secret":   record.NewReference("secret", "secretID"),
				},
			},
			category: record.Record{
				ID:      record.NewRecordID("category", "important"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"title": "This is important.",
				},
			},
			city: record.Record{
				ID:      record.NewRecordID("city", "beautiful"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "This is beautiful.",
				},
			},
			user: record.Record{
				ID:      record.NewRecordID("user", "ownerID"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "Owner",
				},
			},
			secret: record.Record{
				ID:      record.NewRecordID("secret", "secretID"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"content": "Secret of the note",
				},
				ACL: record.ACL{
					record.NewACLEntryDirect("ownerID", record.WriteLevel),
				},
			},
			MockStore: record.NewMockStore(),
		}
	}

	Convey("Test QueryHandler with eager load", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStore()
		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithAPIKey()
		qh.Logger = logging.LoggerEntry("handler")
		qh.TxContext = db.NewMockTxContext()

		Convey("query record with eager load", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {"category": {"$type": "keypath", "$val": "category"}}
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_recordType": "note",
						"_recordID": "note1",
						"_type": "record",
						"_access": null,
						"_ownerID": "ownerID",
						"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
						"city": {"$recordType":"city","$recordID":"beautiful","$type":"ref"},
						"secret":{"$recordType":"secret","$recordID":"secretID","$type":"ref"},
						"_transient": {
							"category": {"_access":null,"_recordType":"category","_recordID":"important","_type":"record","_ownerID":"ownerID", "title": "This is important."}
						}
					}]
				}
			}`)
		})

		Convey("query record with multiple eager load", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {
						"category": {"$type": "keypath", "$val": "category"},
						"city": {"$type": "keypath", "$val": "city"}
					}
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_recordType": "note",
						"_recordID": "note1",
						"_type": "record",
						"_access": null,
						"_ownerID": "ownerID",
						"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
						"city": {"$recordType":"city","$recordID":"beautiful","$type":"ref"},
						"secret":{"$recordType":"secret","$recordID":"secretID","$type":"ref"},
						"_transient": {
							"category": {"_access":null,"_recordType":"category","_recordID":"important","_type":"record","_ownerID":"ownerID", "title": "This is important."},
							"city": {"_access":null,"_recordType":"city","_recordID":"beautiful","_type":"record","_ownerID":"ownerID", "name": "This is beautiful."}
						}
					}]
				}
			}`)
		})

		Convey("query record with eager load on user", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {"user": {"$type": "keypath", "$val": "_owner"}}
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_recordType": "note",
						"_recordID": "note1",
						"_type": "record",
						"_access": null,
						"_ownerID": "ownerID",
						"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
						"city": {"$recordType":"city","$recordID":"beautiful","$type":"ref"},
						"secret":{"$recordType":"secret","$recordID":"secretID","$type":"ref"},
						"_transient": {
							"user": {"_access":null,"_recordType":"user","_recordID":"ownerID","_type":"record","_ownerID":"ownerID", "name": "Owner"}
						}
					}]
				}
			}`)
		})

		Convey("query record with eager load on non public record", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {
						"secret": {"$type": "keypath", "$val": "secret"}
					}
				}
			`))
			resp := httptest.NewRecorder()
			qh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_recordType": "note",
						"_recordID": "note1",
						"_type": "record",
						"_access": null,
						"_ownerID": "ownerID",
						"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
						"city": {"$recordType":"city","$recordID":"beautiful","$type":"ref"},
						"secret":{"$recordType":"secret","$recordID":"secretID","$type":"ref"},
						"_transient": {"secret":null}
					}]
				}
			}`)
		})

		Convey("query record with eager load on non public record with permission", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {
						"secret": {"$type": "keypath", "$val": "secret"}
					}
				}
			`))
			resp := httptest.NewRecorder()
			qh.AuthContext = auth.NewMockContextGetterWithUser("ownerID", true, map[string]bool{})
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"records": [{
						"_recordType": "note",
						"_recordID": "note1",
						"_type": "record",
						"_access": null,
						"_ownerID": "ownerID",
						"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
						"city": {"$recordType":"city","$recordID":"beautiful","$type":"ref"},
						"secret":{"$recordType":"secret","$recordID":"secretID","$type":"ref"},
						"_transient": {
							"secret": {"_access":[{"level":"write","relation":"$direct","user_id":"ownerID"}],"_recordType":"secret","_recordID":"secretID","_type":"record","_ownerID":"ownerID", "content": "Secret of the note"}
						}
					}]
				}
			}`)
		})
	})

	getRecordStoreNullRef := func() (recordStore *referencedRecordStore) {
		return &referencedRecordStore{
			note: record.Record{
				ID:      record.NewRecordID("note", "note1"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"category": record.NewReference("category", "important"),
					"city":     nil,
				},
			},
			category: record.Record{
				ID:      record.NewRecordID("category", "important"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"title": "This is important.",
				},
			},
			city: record.Record{
				ID:      record.NewRecordID("city", "beautiful"),
				OwnerID: "ownerID",
				Data: map[string]interface{}{
					"name": "This is beautiful.",
				},
			},
			MockStore: record.NewMockStore(),
		}
	}

	Convey("Test QueryHandler with eager load but null reference in DB", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStoreNullRef()
		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithAPIKey()
		qh.Logger = logging.LoggerEntry("handler")
		qh.TxContext = db.NewMockTxContext()

		Convey("query record with eager load", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"include": {"city": {"$type": "keypath", "$val": "city"}}
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
					"result": {
						"records": [{
							"_recordType": "note",
							"_recordID": "note1",
							"_type": "record",
							"_access": null,
							"_ownerID": "ownerID",
							"category": {"$recordType":"category","$recordID":"important","$type":"ref"},
							"city": null,
							"_transient": {
								"city": null
							}
						}]
					}
				}`)
		})
	})
}

// a very naive Database that alway returns the single record set onto it
type referencedRecordStore struct {
	note     record.Record
	category record.Record
	city     record.Record
	user     record.Record
	secret   record.Record
	*record.MockStore
}

func (s *referencedRecordStore) UserRecordType() string { return "user" }

func (s *referencedRecordStore) Get(id record.ID, record *record.Record) error {
	switch id.String() {
	case "note/note1":
		*record = s.note
	case "category/important":
		*record = s.category
	case "city/beautiful":
		*record = s.city
	case "user/ownerID":
		*record = s.user
	}
	return nil
}

func (s *referencedRecordStore) GetByIDs(ids []record.ID, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	records := []record.Record{}
	for _, id := range ids {
		var record *record.Record
		switch id.String() {
		case "note/note1":
			record = &s.note
		case "category/important":
			record = &s.category
		case "city/beautiful":
			record = &s.city
		case "user/ownerID":
			record = &s.user
		case "secret/secretID":
			record = &s.secret
		}

		// mock the acl query
		// it will only consider direct record acl entry
		if record != nil {
			if record.ACL == nil || len(record.ACL) == 0 {
				records = append(records, *record)
				continue
			}
			for _, aclEntry := range record.ACL {
				if aclEntry.Relation == "$direct" &&
					aclEntry.UserID == accessControlOptions.ViewAsUser.ID {
					records = append(records, s.secret)
					continue
				}
			}
		}
	}
	return record.NewRows(record.NewMemoryRows(records)), nil
}

func (s *referencedRecordStore) Save(record *record.Record) error {
	return nil
}

func (s *referencedRecordStore) QueryCount(query *record.Query, accessControlOptions *record.AccessControlOptions) (uint64, error) {
	return uint64(1), nil
}

func (s *referencedRecordStore) Query(query *record.Query, accessControlOptions *record.AccessControlOptions) (*record.Rows, error) {
	return record.NewRows(record.NewMemoryRows([]record.Record{s.note})), nil
}

func (s *referencedRecordStore) Extend(recordType string, schema record.Schema) (bool, error) {
	return false, nil
}

func (s *referencedRecordStore) GetSchema(recordType string) (record.Schema, error) {
	typemap := map[string]record.Schema{
		"note": record.Schema{
			"category": record.FieldType{
				Type:          record.TypeReference,
				ReferenceType: "category",
			},
			"city": record.FieldType{
				Type:          record.TypeReference,
				ReferenceType: "city",
			},
		},
		"category": record.Schema{
			"title": record.FieldType{
				Type: record.TypeString,
			},
		},
		"city": record.Schema{
			"name": record.FieldType{
				Type: record.TypeString,
			},
		},
		"user": record.Schema{
			"name": record.FieldType{
				Type: record.TypeString,
			},
		},
	}
	return typemap[recordType], nil
}

func TestRecordQueryWithCount(t *testing.T) {
	getRecordStore := func() (recordStore *queryResultsRecordStore) {
		return &queryResultsRecordStore{
			records: []record.Record{
				record.Record{
					ID: record.NewRecordID("note", "1"),
				},
				record.Record{
					ID: record.NewRecordID("note", "0"),
				},
				record.Record{
					ID: record.NewRecordID("note", "2"),
				},
			},
			typemap: map[string]record.Schema{
				"note": record.Schema{},
			},
			MockStore: record.NewMockStore(),
		}
	}

	Convey("Test QueryHandler count", t, func() {
		qh := &QueryHandler{}
		recordStore := getRecordStore()
		qh.RecordStore = recordStore
		qh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		qh.Logger = logging.LoggerEntry("handler")
		qh.TxContext = db.NewMockTxContext()

		Convey("get count of records", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
				{
					"record_type": "note",
					"count": true
				}
			`))
			resp := httptest.NewRecorder()
			h := handler.APIHandlerToHandler(qh, qh.TxContext)
			h.ServeHTTP(resp, req)
			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": {
					"info": {
						"count": 3
					},
					"records": [{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "1",
						"_access": null
					},
					{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "0",
						"_access": null
					},
					{
						"_type": "record",
						"_recordType": "note",
						"_recordID": "2",
						"_access": null
					}]
				}
			}`)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}
