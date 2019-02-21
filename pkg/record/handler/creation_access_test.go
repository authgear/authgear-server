package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreationAccessHandler(t *testing.T) {
	getRecordStore := func() (recordStore *mockCreationAccessRecordStore) {
		return &mockCreationAccessRecordStore{
			Store: record.NewMockStore(),
		}
	}

	Convey("Test CreationAccessHandler", t, func() {
		// fixture
		sh := &CreationAccessHandler{}
		recordStore := getRecordStore()
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("throw error on invalid Data", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"create_roles": ["Admin", "Writer"]
			}`))
			payload, err := sh.DecodeRequest(req)
			So(err, ShouldBeNil)
			err = payload.Validate()
			So(err, ShouldResemble,
				skyerr.NewInvalidArgument("missing required fields", []string{"type"}))

			req, _ = http.NewRequest("POST", "", strings.NewReader(`{
				"type":         "script",
				"create_roles": "Admin"
			}`))
			_, err = sh.DecodeRequest(req)
			So(err, ShouldHaveSameTypeAs, (*json.UnmarshalTypeError)(nil))
		})

		Convey("set creation access", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"type": "script",
				"create_roles": ["Admin", "Writer"]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"type": "script",
					"create_roles": ["Admin", "Writer"]
				}
			}`)
			So(recordStore.recordType, ShouldEqual, "script")

			roleNames := []string{}
			for _, perACE := range recordStore.acl {
				if perACE.Role != "" {
					roleNames = append(roleNames, perACE.Role)
				}
			}

			So(roleNames, ShouldContain, "Admin")
			So(roleNames, ShouldContain, "Writer")
		})
	})
}

type mockCreationAccessRecordStore struct {
	recordType string
	acl        record.ACL

	record.Store
}

func (c *mockCreationAccessRecordStore) SetRecordAccess(recordType string, acl record.ACL) error {
	c.recordType = recordType
	c.acl = acl
	return c.Store.SetRecordAccess(recordType, acl)
}
