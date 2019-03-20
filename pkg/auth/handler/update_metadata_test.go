package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateMetadataHandler(t *testing.T) {
	var zeroTime time.Time
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test UpdateMetadataHandler", t, func() {
		// fixture
		userID := "john.doe.id"

		uh := &UpdateMetadataHandler{}
		uh.AuthContext = auth.NewMockContextGetterWithUser(userID, true, map[string]bool{})
		profileData := map[string]map[string]interface{}{
			"john.doe.id": map[string]interface{}{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			},
		}
		uh.UserProfileStore = userprofile.NewMockUserProfileStoreByData(profileData)

		loginIDMetadataKeys := [][]string{[]string{"email"}, []string{"username"}}
		passwordAuthProvider := password.NewMockProviderWithPrincipalMap(
			loginIDMetadataKeys,
			map[string]password.Principal{
				"john.doe.principal.id0": password.Principal{
					ID:     "john.doe.principal.id0",
					UserID: "john.doe.id",
					AuthData: map[string]interface{}{
						"username": "john.doe",
					},
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
				"john.doe.principal.id1": password.Principal{
					ID:     "john.doe.principal.id1",
					UserID: "john.doe.id",
					AuthData: map[string]interface{}{
						"email": "john.doe@example.com",
					},
					HashedPassword: []byte("$2a$10$/jm/S1sY6ldfL6UZljlJdOAdJojsJfkjg/pqK47Q8WmOLE19tGWQi"), // 123456
				},
			},
		)
		uh.PasswordAuthProvider = passwordAuthProvider
		uh.TxContext = db.NewMockTxContext()
		h := handler.APIHandlerToHandler(uh, uh.TxContext)

		Convey("should update metadata", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"metadata": {
					"username": "john.doe",
					"email":    "john.doe@example.com",
					"age": 24
				}
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "faseng_access_token",
					"verified":true,
					"verify_info":{},
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s",
					"metadata": {
						"username": "john.doe",
						"email": "john.doe@example.com",
						"age": 24
					}
				}
			}`,
				userID,
				userID,
				userID))
		})

		Convey("shouldn't update metadata in LOGIN_ID_METADATA_KEYS", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"metadata": {
					"username": "mary.jane",
					"email":    "mary.jane@example.com",
					"age": 25
				}
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{
				"result": {
					"user_id": "%s",
					"access_token": "faseng_access_token",
					"verified":true,
					"verify_info":{},
					"created_at": "0001-01-01T00:00:00Z",
					"created_by": "%s",
					"updated_at": "0001-01-01T00:00:00Z",
					"updated_by": "%s",
					"metadata": {
						"username": "john.doe",
						"email": "john.doe@example.com",
						"age": 25
					}
				}
			}`,
				userID,
				userID,
				userID))
		})
	})
}
