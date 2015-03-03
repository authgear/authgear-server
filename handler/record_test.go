package handler

import (
	"os"
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
