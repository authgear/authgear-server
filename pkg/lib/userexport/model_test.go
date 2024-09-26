package userexport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestRequest(t *testing.T) {
	requestBody := `
{
	"format": "ndjson"
}
	`
	Convey("Test serialization of Request", t, func() {
		var request Request
		err := json.Unmarshal([]byte(requestBody), &request)
		So(err, ShouldBeNil)

		serialized, err := json.Marshal(request)
		So(err, ShouldBeNil)

		So(string(serialized), ShouldEqualJSON, `{"csv":{"fields":null},"format":"ndjson"}`)
	})

	Convey("Request JSON Schema", t, func() {
		test := func(requestBody string, errorString string) {
			var request Request
			r, _ := http.NewRequest("POST", "/", strings.NewReader(requestBody))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RequestSchema.Validator(), &request)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}

		test(requestBody, "")
		test("{}", `invalid request body:
<root>: required
  map[actual:<nil> expected:[format] missing:[format]]`)
		test(`
{
	"format": "ndjson"
}
		`, "")
		test(`
{
	"format": "csv"
}
		`, "")
		test(`
{
	"format": "unknown format"
}
		`, `invalid request body:
/format: enum
  map[actual:unknown format expected:[ndjson csv]]`)

		test(`
{
	"format": "csv",
	"csv": {
		"fields": []
	}
}
		`, "")
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "unknown_pointer": "/sub" }]
	}
}
		`, `invalid request body:
/csv/fields/0: required
  map[actual:[unknown_pointer] expected:[pointer] missing:[pointer]]`)
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "pointer": "/sub" }]
	}
}
		`, "")
		test(`
{
	"format": "csv",
	"csv": {
		"fields": [{ "pointer": "/sub", "field_name": "user_id" }]
	}
}
		`, "")
	})

	Convey("CSV fields valid", t, func() {
		var fields []*FieldPointer
		fields = make([]*FieldPointer, 0)
		fields = append(fields, &FieldPointer{
			Pointer: "/sub",
		})
		fields = append(fields, &FieldPointer{
			Pointer: "/address/country",
		})
		fields = append(fields, &FieldPointer{
			Pointer: "/claims/0/email",
		})
		fields = append(fields, &FieldPointer{
			Pointer:   "/give_name",
			FieldName: "my name",
		})
		extractedFields, err := ExtractCSVHeaderField(fields)

		So(err, ShouldBeNil)
		So(len(fields) == len(extractedFields.FieldNames), ShouldBeTrue)

		So(extractedFields.FieldNames[0] == "sub", ShouldBeTrue)
		So(extractedFields.FieldNames[1] == "address.country", ShouldBeTrue)
		So(extractedFields.FieldNames[2] == "claims.0.email", ShouldBeTrue)
		So(extractedFields.FieldNames[3] == "my name", ShouldBeTrue)
	})

	Convey("CSV fields duplicated error", t, func() {
		var fields []*FieldPointer
		fields = make([]*FieldPointer, 0)
		fields = append(fields, &FieldPointer{
			Pointer: "/sub",
		})
		fields = append(fields, &FieldPointer{
			Pointer: "/sub",
		})
		_, err := ExtractCSVHeaderField(fields)

		So(err, ShouldBeError, `{"field_names":["sub","sub"]}`)

		fields = make([]*FieldPointer, 0)
		fields = append(fields, &FieldPointer{
			Pointer: "/sub",
		})
		fields = append(fields, &FieldPointer{
			Pointer:   "/username",
			FieldName: "sub",
		})
		_, err = ExtractCSVHeaderField(fields)

		So(err, ShouldBeError, `{"field_names":["sub","sub"]}`)

		fields = make([]*FieldPointer, 0)
		fields = append(fields, &FieldPointer{
			Pointer: "/claims/0/email",
		})
		fields = append(fields, &FieldPointer{
			Pointer: "/sub",
		})
		fields = append(fields, &FieldPointer{
			Pointer:   "/username",
			FieldName: "claims.0.email",
		})
		_, err = ExtractCSVHeaderField(fields)

		So(err, ShouldBeError, `{"field_names":["claims.0.email","sub","claims.0.email"]}`)
	})

	Convey("Traverse record json with pointers", t, func() {
		record := `
		{
  "sub": "opaque_user_id",

  "preferred_username": "dummy",
  "email": "dummy@dummy.com",
  "phone_number": "+85298765432",

  "email_verified": true,
  "phone_number_verified": false,

  "name": "Dummy Dum",
  "given_name": "Dummy",
  "family_name": "Dum",
  "middle_name": "",
  "nickname": "Lou",
  "profile": "https://example.com",
  "picture": "https://example.com",
  "website": "https://example.com",
  "gender": "male",
  "birthdate": "1990-01-01",
  "zoneinfo": "Asia/Hong_Kong",
  "locale": "zh-Hant-HK",
  "address": {
    "formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
    "street_address": "1 Unnamed Road",
    "locality": "Central",
    "region": "Hong Kong",
    "postal_code": "N/A",
    "country": "HK"
  },

  "custom_attributes": {
    "member_id": "123456789"
  },

  "roles": ["role_a", "role_b"],
  "groups": ["group_a"],

  "disabled": false,

  "identities": [
    {
      "type": "login_id",
      "login_id": {
        "type": "username",
        "key": "username",
        "value": "dummydum",
        "original_value": "DUMMYDUM_login_id1"
      },
      "claims": {
        "preferred_username": "dummydum"
      }
    },
    {
      "type": "login_id",
      "login_id": {
        "type": "email",
        "key": "email",
        "value": "dummy@dummy.com",
        "original_value": "DUMMY@dummy.com"
      },
      "claims": {
        "email": "dummy@dummy.com"
      }
    },
    {
      "type": "login_id",
      "login_id": {
        "type": "phone",
        "key": "phone",
        "value": "+85298765432",
        "original_value": "+85298765432"
      },
      "claims": {
        "phone_number": "+85298765432"
      }
    },
    {
      "type": "oauth",
      "oauth": {
        "provider_alias": "google",
        "provider_type": "google",
        "provider_subject_id": "blahblahblah",
        "user_profile": {
          "email": "dummy@dummy.com"
        }
      },
      "claims": {
        "email": "dummy@dummy.com"
      }
    },
    {
      "type": "ldap",
      "ldap": {
        "server_name": "myldap",
        "last_login_username": "dummydum",
        "user_id_attribute_name": "uid",
        "user_id_attribute_value": "blahblahblah",
        "attributes": {
          "dn": "the DN"
        }
      },
      "claims": {
        "preferred_username": "dummydum"
      }
    }
  ],

  "mfa": {
    "emails": ["dummy@dummy.com"],
    "phone_numbers": ["+85298765432"],
    "totps": [
      {
        "secret": "the-secret",
        "uri": "otpauth://totp...."
      }
    ]
  },

  "biometric_count": 0,
  "passkey_count": 0
}
		`
		var recordJson interface{}
		_ = json.Unmarshal([]byte(record), &recordJson)

		stringValue, _ := TraverseRecordValue(recordJson, "/sub")
		So(stringValue == "opaque_user_id", ShouldBeTrue)

		numberValue, _ := TraverseRecordValue(recordJson, "/biometric_count")
		So(numberValue == "0", ShouldBeTrue)

		traverseDownValue, _ := TraverseRecordValue(recordJson, "/identities/0/login_id/original_value")
		So(traverseDownValue == "DUMMYDUM_login_id1", ShouldBeTrue)

		mapValue, _ := TraverseRecordValue(recordJson, "/identities/3")
		So(mapValue, ShouldEqualJSON, `{"type":"oauth","oauth":{"provider_alias":"google","provider_type":"google","provider_subject_id":"blahblahblah","user_profile":{"email":"dummy@dummy.com"}},"claims":{"email":"dummy@dummy.com"}}`)

		sliceValue, _ := TraverseRecordValue(recordJson, "/roles")
		So(sliceValue, ShouldEqualJSON, `["role_a","role_b"]`)

		trueValue, _ := TraverseRecordValue(recordJson, "/email_verified")
		So(trueValue == "true", ShouldBeTrue)

		falseValue, _ := TraverseRecordValue(recordJson, "/disabled")
		So(falseValue == "false", ShouldBeTrue)

		notFoundValue, _ := TraverseRecordValue(recordJson, "/dummy_cursor")
		So(notFoundValue == "", ShouldBeTrue)
	})

}
