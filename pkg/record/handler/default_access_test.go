package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaultAccessHandler(t *testing.T) {
	getRecordStore := func() (recordStore *mockDefaultAccessRecordStore) {
		return &mockDefaultAccessRecordStore{
			Store: record.NewMockStore(),
		}
	}

	Convey("Test DefaultAccessHandler", t, func() {
		// fixture
		sh := &DefaultAccessHandler{}
		recordStore := getRecordStore()
		sh.RecordStore = recordStore
		sh.Logger = logging.LoggerEntry("handler")
		sh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("set default access", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"type": "script",
				"default_access": [
					{"public": true, "level": "read"},
					{"role": "admin", "level": "write"}
				]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"type": "script",
					"default_access": [
						{"public": true, "level": "read"},
						{"role": "admin", "level": "write"}
					]
				}
			}`)

			So(recordStore.recordType, ShouldEqual, "script")

			admin := authinfo.AuthInfo{
				Roles: []string{"admin"},
			}
			So(recordStore.acl.Accessible(&admin, record.WriteLevel), ShouldEqual, true)
			So(recordStore.acl.Accessible(nil, record.ReadLevel), ShouldEqual, true)
		})
	})
}

type mockDefaultAccessRecordStore struct {
	recordType string
	acl        record.ACL

	record.Store
}

func (s *mockDefaultAccessRecordStore) SetRecordDefaultAccess(recordType string, acl record.ACL) error {
	s.recordType = recordType
	s.acl = acl
	return s.Store.SetRecordDefaultAccess(recordType, acl)
}
