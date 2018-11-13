package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFieldAccessUpdateHandler(t *testing.T) {
	getRecordStore := func() (recordStore *mockFieldAccessUpdateRecordStore) {
		return &mockFieldAccessUpdateRecordStore{
			Store: record.NewMockStore(),
		}
	}

	Convey("Test FieldAccessUpdateHandler", t, func(c C) {
		// fixture
		fh := &FieldAccessUpdateHandler{}
		recordStore := getRecordStore()
		fh.RecordStore = recordStore
		fh.Logger = logging.LoggerEntry("handler")
		fh.TxContext = db.NewMockTxContext()

		h := handler.APIHandlerToHandler(fh, fh.TxContext)

		c.Convey("should set empty Field ACL", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access": []
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"access": []
				}
			}`)
			So(recordStore.acl, ShouldResemble, record.NewFieldACL(record.FieldACLEntryList{}))
		})

		c.Convey("should set Field ACL", func() {
			fieldACLJSON := `[
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
			]`

			req, _ := http.NewRequest(
				"POST",
				"",
				strings.NewReader(fmt.Sprintf(`{"access": %s}`, fieldACLJSON)),
			)
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(
				resp.Body.Bytes(),
				ShouldEqualJSON,
				fmt.Sprintf(`{
					"result": {
						"access": %s
					}
				}`, fieldACLJSON),
			)
			So(recordStore.acl, ShouldResemble, record.NewFieldACL(record.FieldACLEntryList{
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
			}))
		})
	})
}

type mockFieldAccessUpdateRecordStore struct {
	acl record.FieldACL

	record.Store
}

func (c *mockFieldAccessUpdateRecordStore) SetRecordFieldAccess(acl record.FieldACL) (err error) {
	c.acl = acl
	return c.Store.SetRecordFieldAccess(acl)
}
