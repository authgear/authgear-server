package handler

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/oursky/ourd/auth"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/fs"
	"github.com/oursky/ourd/router"
)

func TestNewRecordPayload(t *testing.T) {
	rpayload := router.Payload{
		Data: map[string]interface{}{
			"database_id": "somedbid",
		},
	}

	payload := newRecordPayload(&rpayload)

	if payload.Payload != &rpayload {
		t.Errorf("got payload = %v, want %v", payload.Payload, rpayload)
	}

	if payload.DatabaseID != "somedbid" {
		t.Errorf("got DatabaseID = %v, want somedbid", payload.DatabaseID)
	}
}

func TestRecordPayloadIsValidDB(t *testing.T) {
	payload := recordPayload{}

	payload.DatabaseID = "_public"
	if !payload.IsValidDB() {
		t.Error("got IsValidDB() = false, want true")
	}

	payload.DatabaseID = "_private"
	if !payload.IsValidDB() {
		t.Error("got IsValidDB() = false, want true")
	}

	payload.DatabaseID = "invaliddbid"
	if payload.IsValidDB() {
		t.Error("got IsValidDB() = true, want false")
	}

}

func TestRecordPayloadIsPublicDB(t *testing.T) {
	payload := recordPayload{}

	payload.DatabaseID = "_public"
	if !payload.IsPublicDB() {
		t.Error("got IsPublicDB() = false, want true")
	}

	payload.DatabaseID = "_private"
	if payload.IsPublicDB() {
		t.Error("got IsPublicDB() = true, want false")
	}
}

func TestRecordPayloadIsReadOnly(t *testing.T) {
	readonlytests := []struct {
		action string
		result bool
	}{
		{"record:save", false},
		{"record:fetch", true},
		{"record:query", true},
		{"record:delete", false},
	}

	payload := recordPayload{
		Payload: &router.Payload{Data: map[string]interface{}{}},
	}

	for _, tt := range readonlytests {
		payload.Payload.Data["action"] = tt.action
		isReadonly := payload.IsReadOnly()
		if isReadonly != tt.result {
			t.Errorf("got {action: %#v}.IsReadOnly() = %v, want %v", tt.action, isReadonly, tt.result)
		}
	}
}

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
		"_id":       "recordkey",
		"_type":     "recordtype",
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
		"_id": "recordkey",
		"_type": "recordtype",
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

// recordHandlerParam holds the parameters being passed to a RecordHandler
type calledHandler bool

func (h *calledHandler) SetCalled(p *recordPayload, r *router.Response, db oddb.Database) {
	*h = true
}

// TODO(limouren): refactor TokenStores commonly used in testing to
// a separate package

// trueStore is a TokenStore that always noop on Put and assign itself on Get
type trueStore auth.Token

func (store *trueStore) Get(id string, token *auth.Token) error {
	*token = auth.Token(*store)
	return nil
}

func (store *trueStore) Put(token *auth.Token) error {
	return nil
}

// errStore is a TokenStore that always noop and returns itself as error
// on both Get and Put
type errStore auth.TokenNotFoundError

func (store *errStore) Get(id string, token *auth.Token) error {
	return (*auth.TokenNotFoundError)(store)
}

func (store *errStore) Put(token *auth.Token) error {
	return (*auth.TokenNotFoundError)(store)
}

func TestInjectRecordHandler(t *testing.T) {
	validTokenStore := &trueStore{
		AccessToken: "recordaccesstoken",
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		UserInfoID:  "recorduserinfoid",
	}

	notfoundTokenStore := &errStore{}

	dbinjecttests := []struct {
		action     string
		dbID       string
		tokenStore auth.TokenStore
		called     bool
		resultCode int
	}{
		{"record:fetch", "_public", validTokenStore, true, 0},
		{"record:query", "_public", notfoundTokenStore, true, 0},
		{"record:save", "_public", validTokenStore, true, 0},
		{"record:delete", "_public", notfoundTokenStore, false, 104},
		{"record:save", "invaliddbid", notfoundTokenStore, false, 202},
		{"record:fetch", "_private", validTokenStore, true, 0},
		{"record:query", "_private", notfoundTokenStore, false, 104},
		{"record:save", "_private", validTokenStore, true, 0},
		{"record:delete", "_private", notfoundTokenStore, false, 104},
	}

	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.oddb.test", dir)
	if err != nil {
		panic(err)
	}

	recordService := RecordService{}
	rpayload := router.Payload{
		Data:   map[string]interface{}{},
		DBConn: conn,
	}

	for i, tt := range dbinjecttests {
		called := false
		response := router.Response{}
		rpayload.Data["action"] = tt.action
		rpayload.Data["database_id"] = tt.dbID

		recordService.TokenStore = tt.tokenStore
		recordService.injectRecordHandler(((*calledHandler)(&called)).SetCalled)(&rpayload, &response)

		if called != tt.called {
			t.Errorf("row %v: got called = %v, want %v", i, called, tt.called)
		}

		if !called && response.Result.(genericError).Code != ErrCode(tt.resultCode) {
			t.Errorf("row %v: got response.Result.(genericError).Code = %v, want %v", i, response.Result.(genericError).Code, tt.resultCode)
		}
	}
}

func TestRecordSaveHandler(t *testing.T) {
	payload := recordPayload{
		Payload: &router.Payload{
			Data: map[string]interface{}{
				"action": "record:save",
				"records": []map[string]interface{}{
					map[string]interface{}{
						"_id":   "id1",
						"_type": "type1",
						"k1":    "v1",
						"k2":    "v2",
					},
					map[string]interface{}{
						"_id":   "id2",
						"_type": "type2",
						"k3":    "v3",
						"k4":    "v4",
					},
				},
			},
		},
	}

	dir := tempDir()
	defer os.RemoveAll(dir)

	conn, err := fs.Open("com.oursky.oddb.test", dir)
	if err != nil {
		panic(err)
	}

	db := conn.PublicDB()

	response := router.Response{}
	RecordSaveHandler(&payload, &response, db)

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

	expectedResult := []interface{}{
		transportRecord(expectedRecord1),
		transportRecord(expectedRecord2),
	}

	if !reflect.DeepEqual(response.Result, expectedResult) {
		t.Fatalf("got response.Result = %#v, want %#v", response.Result, expectedResult)
	}
}
