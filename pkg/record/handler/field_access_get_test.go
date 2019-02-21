package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFieldAccessGetHandler(t *testing.T) {
	getRecordStore := func() (recordStore *record.MockStore) {
		return record.NewMockStore()
	}

	Convey("Test FieldAccessGetHandler", t, func() {
		// fixture
		fh := &FieldAccessGetHandler{}
		recordStore := getRecordStore()
		fh.RecordStore = recordStore
		fh.Logger = logging.LoggerEntry("handler")
		fh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(fh, fh.TxContext)

		Convey("should return empty array for no Field ACL settings", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"access": []
				}
			}`)
		})

		Convey("should return Field ACL settings", func() {
			recordStore.FieldAccess = record.NewFieldACL(record.FieldACLEntryList{
				{
					RecordType:   "*",
					RecordField:  "*",
					UserRole:     record.FieldUserRole{record.PublicFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
				{
					RecordType:   "note",
					RecordField:  "content",
					UserRole:     record.FieldUserRole{record.SpecificUserFieldUserRoleType, "johndoe"},
					Writable:     false,
					Readable:     true,
					Comparable:   false,
					Discoverable: false,
				},
				{
					RecordType:   "note",
					RecordField:  "*",
					UserRole:     record.FieldUserRole{record.OwnerFieldUserRoleType, ""},
					Writable:     true,
					Readable:     true,
					Comparable:   true,
					Discoverable: true,
				},
			})
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"access": [
						{
							"record_type":"note",
							"record_field":"content",
							"user_role":"_user_id:johndoe",
							"writable":false,
							"readable":true,
							"comparable":false,
							"discoverable":false
						},
						{
							"record_type":"note",
							"record_field":"*",
							"user_role":"_owner",
							"writable":true,
							"readable":true,
							"comparable":true,
							"discoverable":true
						},
						{
							"record_type":"*",
							"record_field":"*",
							"user_role":"_public",
							"writable":true,
							"readable":true,
							"comparable":true,
							"discoverable":true
						}
					]
				}
			}`)
		})
	})
}
