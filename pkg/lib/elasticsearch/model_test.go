package elasticsearch

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	. "github.com/authgear/authgear-server/pkg/util/testing"
)

func TestRawToSource(t *testing.T) {
	Convey("RawToSource", t, func() {
		lastLoginAt := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)
		raw := &model.ElasticsearchUserRaw{
			ID:                "ID",
			AppID:             "APP_ID",
			CreatedAt:         time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:         time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC),
			LastLoginAt:       &lastLoginAt,
			IsDisabled:        true,
			Email:             []string{"user@example.com"},
			PreferredUsername: []string{"user"},
			PhoneNumber:       []string{"+85298765432"},
		}

		source := RawToSource(raw)
		sourceBytes, err := json.Marshal(source)
		So(err, ShouldBeNil)

		So(sourceBytes, ShouldEqualJSON, `{
			"app_id": "APP_ID",
			"id": "ID",
			"created_at": "2006-01-02T03:04:05Z",
			"updated_at": "2006-01-02T03:04:05Z",
			"last_login_at": "2006-01-02T03:04:05Z",
			"is_disabled": true,
			"email": [
				"user@example.com"
			],
			"email_text": [
				"user@example.com"
			],
			"email_local_part": [
				"user"
			],
			"email_local_part_text": [
				"user"
			],
			"email_domain": [
				"example.com"
			],
			"email_domain_text": [
				"example.com"
			],
			"preferred_username": [
				"user"
			],
			"preferred_username_text": [
				"user"
			],
			"phone_number": [
				"+85298765432"
			],
			"phone_number_text": [
				"+85298765432"
			],
			"phone_number_country_code": [
				"852"
			],
			"phone_number_national_number": [
				"98765432"
			],
			"phone_number_national_number_text": [
				"98765432"
			]
		}`)
	})
}
