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
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

// TODO: TestRecordQueryResult

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
					ID:    "faseng.cat.id",
					Roles: []string{"user"},
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
					ID:    "faseng.cat.id",
					Roles: []string{"user"},
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
