package handler

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

func TestTransportRecordMarshalJSON(t *testing.T) {
	r := transportRecord{
		Key:  "recordkey",
		Type: "recordtype",
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
		Key:  "recordkey",
		Type: "recordtype",
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
		Key:  "recordkey",
		Type: "recordtype",
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

// recordHandlerParam holds the parameters being passed to a RecordHandler
type calledHandler bool

func (h *calledHandler) SetCalled(p *recordPayload, r *router.Response, db oddb.Database) {
	*h = true
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
	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.oddb.test", dir)
	if err != nil {
		panic(err)
	}

	db := conn.PublicDB()

	payload := router.Payload{
		Data: map[string]interface{}{
			"action": "record:save",
			"records": []interface{}{
				map[string]interface{}{
					"_id": "type1/id1",
					"k1":  "v1",
					"k2":  "v2",
				},
				map[string]interface{}{
					"_id": "type2/id2",
					"k3":  "v3",
					"k4":  "v4",
				},
			},
		},
		Database: db,
		UserInfo: &oddb.UserInfo{},
	}

	response := router.Response{}
	RecordSaveHandler(&payload, &response)

	// check for DB persistences

	record1 := oddb.Record{}
	if err := db.Get("id1", &record1); err != nil {
		t.Fatalf("got err = %v, want err = nil", err)
	}

	expectedRecord1 := oddb.Record{
		Type: "type1",
		Key:  "id1",
		Data: map[string]interface{}{
			"k1": "v1",
			"k2": "v2",
		},
	}

	if !reflect.DeepEqual(record1, expectedRecord1) {
		t.Errorf("got record1 = %#v, want %#v", record1, expectedRecord1)
	}

	record2 := oddb.Record{}
	if err := db.Get("id2", &record2); err != nil {
		t.Fatalf("got err = %v, want err = nil", err)
	}

	expectedRecord2 := oddb.Record{
		Type: "type2",
		Key:  "id2",
		Data: map[string]interface{}{
			"k3": "v3",
			"k4": "v4",
		},
	}

	if !reflect.DeepEqual(record2, expectedRecord2) {
		t.Errorf("got record2 = %#v, want %#v", record2, expectedRecord2)
	}

	// check for Response

	expectedResult := []responseItem{
		newResponseItem((*transportRecord)(&expectedRecord1)),
		newResponseItem((*transportRecord)(&expectedRecord2)),
	}

	if !reflect.DeepEqual(response.Result, expectedResult) {
		t.Fatalf("got response.Result = %#v, want %#v", response.Result, expectedResult)
	}
}

type mapDB struct {
	Map map[string]oddb.Record
	oddb.Database
}

func (db mapDB) Get(key string, record *oddb.Record) error {
	r, ok := db.Map[key]
	if !ok {
		return oddb.ErrRecordNotFound
	}

	*record = r

	return nil
}

func TestRecordFetch(t *testing.T) {
	record1 := transportRecord{Key: "1", Type: "record"}
	record2 := transportRecord{Key: "2", Type: "record"}
	db := mapDB{
		Map: map[string]oddb.Record{
			"1": oddb.Record(record1),
			"2": oddb.Record(record2),
		},
	}

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
				newResponseItem(&record1),
				newResponseItem(&record2),
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
				newResponseItem(&record1),
				newResponseItemErr("type/not-exist", oderr.ErrRecordNotFound),
				newResponseItem(&record2),
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
