package userimport

import (
	"encoding/json"
	"fmt"
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
	"upsert": true,
	"identifier": "preferred_username",
	"records": [
		{
			"preferred_username": "johndoe",
			"email": "johndoe@example.com",
			"phone_number": "+85298765432",

			"disabled": true,

			"email_verified": true,
			"phone_number_verified": true,

			"name": "John Doe",
			"given_name": "John",
			"family_name": "Doe",
			"middle_name": "middle",
			"nickname": "JohnDoe",
			"profile": "https://example.com/profile",
			"picture": "https://example.com/picture",
			"website": "https://example.com/website",
			"gender": "male",
			"birthdate": "1970-01-01",
			"zoneinfo": "Asia/Hong_Kong",
			"locale": "zh-HK",
			"address": {
				"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
				"street_address": "1 Unnamed Road",
				"locality": "Central",
				"region": "Hong Kong Island",
				"postal_code": "N/A",
				"country": "HK"
			},

			"custom_attributes": {
				"member_id": "123456789"
			},

			"roles": ["role_a", "role_b"],
			"groups": ["group_a", "group_b"],

			"password": {
				"type": "bcrypt",
				"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
			},

			"mfa": {
				"email": "johndoe@example.com",
				"phone_number": "+85298765432",
				"password": {
					"type": "bcrypt",
					"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
				},
				"totp": {
					"secret": "JBSWY3DPEHPK3PXP"
				}
			}
		}
	]
}
	`
	Convey("Test serialization of Request", t, func() {
		var request Request
		err := json.Unmarshal([]byte(requestBody), &request)
		So(err, ShouldBeNil)

		serialized, err := json.Marshal(request)
		So(err, ShouldBeNil)

		So(string(serialized), ShouldEqualJSON, requestBody)
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
  map[actual:<nil> expected:[identifier records] missing:[identifier records]]`)
		test(`
{
	"identifier": "foobar",
	"records": []
}
		`, `invalid request body:
/identifier: enum
  map[actual:foobar expected:[preferred_username email phone_number]]
/records: minItems
  map[actual:0 expected:1]`)
	})
}

func TestRecord(t *testing.T) {
	Convey("Record", t, func() {
		Convey("preferred_username", func() {
			r := Record{}

			_, ok := r.PreferredUsername()
			So(ok, ShouldBeFalse)

			r = Record{
				"preferred_username": nil,
			}

			v, ok := r.PreferredUsername()
			So(ok, ShouldBeTrue)
			So(v, ShouldBeNil)

			r = Record{
				"preferred_username": "a",
			}

			v, ok = r.PreferredUsername()
			So(ok, ShouldBeTrue)
			So(*v, ShouldEqual, "a")
		})

		Convey("email", func() {
			r := Record{}

			_, ok := r.Email()
			So(ok, ShouldBeFalse)

			r = Record{
				"email": nil,
			}

			v, ok := r.Email()
			So(ok, ShouldBeTrue)
			So(v, ShouldBeNil)

			r = Record{
				"email": "a",
			}

			v, ok = r.Email()
			So(ok, ShouldBeTrue)
			So(*v, ShouldEqual, "a")
		})

		Convey("phone_number", func() {
			r := Record{}

			_, ok := r.PhoneNumber()
			So(ok, ShouldBeFalse)

			r = Record{
				"phone_number": nil,
			}

			v, ok := r.PhoneNumber()
			So(ok, ShouldBeTrue)
			So(v, ShouldBeNil)

			r = Record{
				"phone_number": "a",
			}

			v, ok = r.PhoneNumber()
			So(ok, ShouldBeTrue)
			So(*v, ShouldEqual, "a")
		})

		Convey("disabled", func() {
			r := Record{}

			_, ok := r.Disabled()
			So(ok, ShouldBeFalse)

			r = Record{
				"disabled": nil,
			}

			So(func() {
				r.Disabled()
			}, ShouldPanicWith, fmt.Errorf("disabled is expected to be non-null"))

			r = Record{
				"disabled": true,
			}

			v, ok := r.Disabled()
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, true)
		})

		Convey("email_verified", func() {
			r := Record{}

			_, ok := r.EmailVerified()
			So(ok, ShouldBeFalse)

			r = Record{
				"email_verified": nil,
			}

			So(func() {
				r.EmailVerified()
			}, ShouldPanicWith, fmt.Errorf("email_verified is expected to be non-null"))

			r = Record{
				"email_verified": true,
			}

			v, ok := r.EmailVerified()
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, true)
		})

		Convey("phone_number_verified", func() {
			r := Record{}

			_, ok := r.PhoneNumberVerified()
			So(ok, ShouldBeFalse)

			r = Record{
				"phone_number_verified": nil,
			}

			So(func() {
				r.PhoneNumberVerified()
			}, ShouldPanicWith, fmt.Errorf("phone_number_verified is expected to be non-null"))

			r = Record{
				"phone_number_verified": true,
			}

			v, ok := r.PhoneNumberVerified()
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, true)
		})

		Convey("standard_attributes", func() {
			r := Record{}

			l := r.StandardAttributesList()
			So(l, ShouldBeEmpty)

			recordString := `
{
	"preferred_username": "johndoe",
	"email": "johndoe@example.com",
	"phone_number": "+85298765432",

	"disabled": true,

	"email_verified": true,
	"phone_number_verified": true,

	"name": "John Doe",
	"given_name": "John",
	"family_name": "Doe",
	"middle_name": "middle",
	"nickname": "JohnDoe",
	"profile": "https://example.com/profile",
	"picture": "https://example.com/picture",
	"website": "https://example.com/website",
	"gender": "male",
	"birthdate": "1970-01-01",
	"zoneinfo": "Asia/Hong_Kong",
	"locale": "zh-HK",
	"address": {
		"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
		"street_address": "1 Unnamed Road",
		"locality": "Central",
		"region": "Hong Kong Island",
		"postal_code": "N/A",
		"country": "HK"
	},

	"custom_attributes": {
		"member_id": "123456789"
	},

	"roles": ["role_a", "role_b"],
	"groups": ["group_a", "group_b"],

	"password": {
		"type": "bcrypt",
		"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	},

	"mfa": {
		"email": "johndoe@example.com",
		"phone_number": "+85298765432",
		"password": {
			"type": "bcrypt",
			"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
		},
		"totp": {
			"secret": "JBSWY3DPEHPK3PXP"
		}
	}
}
		`

			r = Record{}
			err := json.Unmarshal([]byte(recordString), &r)
			So(err, ShouldBeNil)

			l = r.StandardAttributesList()
			So(len(l), ShouldEqual, 16)
		})

		Convey("custom_attributes", func() {
			r := Record{}

			l := r.CustomAttributesList()
			So(l, ShouldBeEmpty)

			r = Record{
				"custom_attributes": nil,
			}

			So(func() {
				r.CustomAttributesList()
			}, ShouldPanicWith, fmt.Errorf("custom_attributes is expected to be non-null"))

			r = Record{
				"custom_attributes": map[string]interface{}{},
			}

			l = r.CustomAttributesList()
			So(l, ShouldBeEmpty)

			r = Record{
				"custom_attributes": map[string]interface{}{
					"a": "b",
				},
			}

			l = r.CustomAttributesList()
			So(len(l), ShouldEqual, 1)
		})

		Convey("roles", func() {
			r := Record{}
			_, ok := r.Roles()
			So(ok, ShouldBeFalse)

			r = Record{
				"roles": nil,
			}

			So(func() {
				r.Roles()
			}, ShouldPanicWith, fmt.Errorf("roles is expected to be of type []string, but was <nil>"))

			r = Record{
				"roles": []interface{}{"a"},
			}

			v, ok := r.Roles()
			So(ok, ShouldBeTrue)
			So(v, ShouldResemble, []string{"a"})
		})

		Convey("groups", func() {
			r := Record{}
			_, ok := r.Groups()
			So(ok, ShouldBeFalse)

			r = Record{
				"groups": nil,
			}

			So(func() {
				r.Groups()
			}, ShouldPanicWith, fmt.Errorf("groups is expected to be of type []string, but was <nil>"))

			r = Record{
				"groups": []interface{}{"a"},
			}

			v, ok := r.Groups()
			So(ok, ShouldBeTrue)
			So(v, ShouldResemble, []string{"a"})
		})

		Convey("password", func() {
			r := Record{}
			_, ok := r.Password()
			So(ok, ShouldBeFalse)

			r = Record{
				"password": nil,
			}

			So(func() {
				r.Password()
			}, ShouldPanicWith, fmt.Errorf("password is expected to be non-null"))

			r = Record{
				"password": map[string]interface{}{},
			}

			v, ok := r.Password()
			So(ok, ShouldBeTrue)
			So(v, ShouldNotBeNil)
			So(len(v), ShouldEqual, 0)
		})

		Convey("mfa", func() {
			r := Record{}
			_, ok := r.MFA()
			So(ok, ShouldBeFalse)

			r = Record{
				"mfa": nil,
			}

			So(func() {
				r.MFA()
			}, ShouldPanicWith, fmt.Errorf("mfa is expected to be non-null"))

			r = Record{
				"mfa": map[string]interface{}{},
			}

			v, ok := r.MFA()
			So(ok, ShouldBeTrue)
			So(v, ShouldNotBeNil)
			So(len(v), ShouldEqual, 0)
		})

		Convey("Redact", func() {
			recordString := `
{
	"preferred_username": "johndoe",
	"email": "johndoe@example.com",
	"phone_number": "+85298765432",

	"disabled": true,

	"email_verified": true,
	"phone_number_verified": true,

	"name": "John Doe",
	"given_name": "John",
	"family_name": "Doe",
	"middle_name": "middle",
	"nickname": "JohnDoe",
	"profile": "https://example.com/profile",
	"picture": "https://example.com/picture",
	"website": "https://example.com/website",
	"gender": "male",
	"birthdate": "1970-01-01",
	"zoneinfo": "Asia/Hong_Kong",
	"locale": "zh-HK",
	"address": {
		"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
		"street_address": "1 Unnamed Road",
		"locality": "Central",
		"region": "Hong Kong Island",
		"postal_code": "N/A",
		"country": "HK"
	},

	"custom_attributes": {
		"member_id": "123456789"
	},

	"roles": ["role_a", "role_b"],
	"groups": ["group_a", "group_b"],

	"password": {
		"type": "bcrypt",
		"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	},

	"mfa": {
		"email": "johndoe@example.com",
		"phone_number": "+85298765432",
		"password": {
			"type": "bcrypt",
			"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
		},
		"totp": {
			"secret": "JBSWY3DPEHPK3PXP"
		}
	}
}
		`

			expected := `
{
	"preferred_username": "johndoe",
	"email": "johndoe@example.com",
	"phone_number": "+85298765432",

	"disabled": true,

	"email_verified": true,
	"phone_number_verified": true,

	"name": "John Doe",
	"given_name": "John",
	"family_name": "Doe",
	"middle_name": "middle",
	"nickname": "JohnDoe",
	"profile": "https://example.com/profile",
	"picture": "https://example.com/picture",
	"website": "https://example.com/website",
	"gender": "male",
	"birthdate": "1970-01-01",
	"zoneinfo": "Asia/Hong_Kong",
	"locale": "zh-HK",
	"address": {
		"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
		"street_address": "1 Unnamed Road",
		"locality": "Central",
		"region": "Hong Kong Island",
		"postal_code": "N/A",
		"country": "HK"
	},

	"custom_attributes": {
		"member_id": "123456789"
	},

	"roles": ["role_a", "role_b"],
	"groups": ["group_a", "group_b"],

	"password": {
		"type": "bcrypt",
		"password_hash": "REDACTED"
	},

	"mfa": {
		"email": "johndoe@example.com",
		"phone_number": "+85298765432",
		"password": {
			"type": "bcrypt",
			"password_hash": "REDACTED"
		},
		"totp": {
			"secret": "REDACTED"
		}
	}
}
		`

			r := Record{}
			err := json.Unmarshal([]byte(recordString), &r)
			So(err, ShouldBeNil)
			r.Redact()

			redacted, err := json.Marshal(r)
			So(err, ShouldBeNil)

			So(string(redacted), ShouldEqualJSON, expected)
		})
	})
}

func TestPassword(t *testing.T) {
	Convey("Password", t, func() {
		p := Password{
			"type":          "bcrypt",
			"password_hash": "hash",
		}

		So(p.Type(), ShouldEqual, PasswordTypeBcrypt)
		So(p.PasswordHash(), ShouldEqual, "hash")
	})
}

func TestTOTP(t *testing.T) {
	Convey("TOTP", t, func() {
		p := TOTP{
			"secret": "secret",
		}

		So(p.Secret(), ShouldEqual, "secret")
	})
}

func TestMFA(t *testing.T) {
	Convey("MFA", t, func() {
		Convey("email", func() {
			r := MFA{}

			_, ok := r.Email()
			So(ok, ShouldBeFalse)

			r = MFA{
				"email": nil,
			}

			v, ok := r.Email()
			So(ok, ShouldBeTrue)
			So(v, ShouldBeNil)

			r = MFA{
				"email": "a",
			}

			v, ok = r.Email()
			So(ok, ShouldBeTrue)
			So(*v, ShouldEqual, "a")
		})

		Convey("phone_number", func() {
			r := MFA{}

			_, ok := r.PhoneNumber()
			So(ok, ShouldBeFalse)

			r = MFA{
				"phone_number": nil,
			}

			v, ok := r.PhoneNumber()
			So(ok, ShouldBeTrue)
			So(v, ShouldBeNil)

			r = MFA{
				"phone_number": "a",
			}

			v, ok = r.PhoneNumber()
			So(ok, ShouldBeTrue)
			So(*v, ShouldEqual, "a")
		})

		Convey("password", func() {
			r := MFA{}
			_, ok := r.Password()
			So(ok, ShouldBeFalse)

			r = MFA{
				"password": nil,
			}

			So(func() {
				r.Password()
			}, ShouldPanicWith, fmt.Errorf("password is expected to be non-null"))

			r = MFA{
				"password": map[string]interface{}{},
			}

			v, ok := r.Password()
			So(ok, ShouldBeTrue)
			So(v, ShouldNotBeNil)
			So(len(v), ShouldEqual, 0)
		})

		Convey("totp", func() {
			r := MFA{}
			_, ok := r.TOTP()
			So(ok, ShouldBeFalse)

			r = MFA{
				"totp": nil,
			}

			So(func() {
				r.TOTP()
			}, ShouldPanicWith, fmt.Errorf("totp is expected to be non-null"))

			r = MFA{
				"totp": map[string]interface{}{},
			}

			v, ok := r.TOTP()
			So(ok, ShouldBeTrue)
			So(v, ShouldNotBeNil)
			So(len(v), ShouldEqual, 0)
		})
	})
}

func TestRecordSchema(t *testing.T) {
	recordString := `
{
	"preferred_username": "johndoe",
	"email": "johndoe@example.com",
	"phone_number": "+85298765432",

	"disabled": true,

	"email_verified": true,
	"phone_number_verified": true,

	"name": "John Doe",
	"given_name": "John",
	"family_name": "Doe",
	"middle_name": "middle",
	"nickname": "JohnDoe",
	"profile": "https://example.com/profile",
	"picture": "https://example.com/picture",
	"website": "https://example.com/website",
	"gender": "male",
	"birthdate": "1970-01-01",
	"zoneinfo": "Asia/Hong_Kong",
	"locale": "zh-HK",
	"address": {
		"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
		"street_address": "1 Unnamed Road",
		"locality": "Central",
		"region": "Hong Kong Island",
		"postal_code": "N/A",
		"country": "HK"
	},

	"custom_attributes": {
		"member_id": "123456789"
	},

	"roles": ["role_a", "role_b"],
	"groups": ["group_a", "group_b"],

	"password": {
		"type": "bcrypt",
		"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	},

	"mfa": {
		"email": "johndoe@example.com",
		"phone_number": "+85298765432",
		"password": {
			"type": "bcrypt",
			"password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
		},
		"totp": {
			"secret": "JBSWY3DPEHPK3PXP"
		}
	}
}
		`

	Convey("Record JSON schema for password", t, func() {
		test := func(recordString string, errorString string) {
			var record Record
			r, _ := http.NewRequest("POST", "/", strings.NewReader(recordString))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RecordSchemaForIdentifierEmail.Validator(), &record)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}

		test(`{
			"email": "user@example.com",
			"password": {
				"unknown": 1
			}
		}`, `invalid request body:
/password: required
  map[actual:[unknown] expected:[password_hash type] missing:[password_hash type]]
/password/unknown: `)
	})

	Convey("Record JSON schema for mfa", t, func() {
		test := func(recordString string, errorString string) {
			var record Record
			r, _ := http.NewRequest("POST", "/", strings.NewReader(recordString))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RecordSchemaForIdentifierEmail.Validator(), &record)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}

		test(`{
			"email": "user@example.com",
			"mfa": {
				"unknown": 1
			}
		}`, `invalid request body:
/mfa/unknown: `)

		test(`{
			"email": "user@example.com",
			"mfa": {
				"totp": {
					"secret": "a",
					"unknown": 1
				}
			}
		}`, `invalid request body:
/mfa/totp/unknown: `)
	})

	Convey("Record JSON Schema for email", t, func() {
		test := func(recordString string, errorString string) {
			var record Record
			r, _ := http.NewRequest("POST", "/", strings.NewReader(recordString))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RecordSchemaForIdentifierEmail.Validator(), &record)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}
		test(recordString, "")

		test(`{}`, `invalid request body:
<root>: required
  map[actual:<nil> expected:[email] missing:[email]]`)

		allNulls := `
		{
			"email": "user@example.com",

				"name": null,
				"given_name": null,
				"family_name": null,
				"middle_name": null,
				"nickname": null,
				"profile": null,
				"picture": null,
				"website": null,
				"gender": null,
				"birthdate": null,
				"zoneinfo": null,
				"locale": null,
				"address": null,

				"mfa": {
					"email": null,
					"phone_number": null
				}
		}
		`
		test(allNulls, "")

		test(`{
			"email": "user@example.com",
			"password": {},
			"mfa": {
				"password": {},
				"totp": {}
			}
		}`, `invalid request body:
/mfa/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]
/mfa/totp: required
  map[actual:<nil> expected:[secret] missing:[secret]]
/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]`)
	})

	Convey("Record JSON Schema for phone_number", t, func() {
		test := func(recordString string, errorString string) {
			var record Record
			r, _ := http.NewRequest("POST", "/", strings.NewReader(recordString))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RecordSchemaForIdentifierPhoneNumber.Validator(), &record)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}
		test(recordString, "")

		test(`{}`, `invalid request body:
<root>: required
  map[actual:<nil> expected:[phone_number] missing:[phone_number]]`)

		allNulls := `
		{
			"phone_number": "+85298765432",

				"name": null,
				"given_name": null,
				"family_name": null,
				"middle_name": null,
				"nickname": null,
				"profile": null,
				"picture": null,
				"website": null,
				"gender": null,
				"birthdate": null,
				"zoneinfo": null,
				"locale": null,
				"address": null,

				"mfa": {
					"email": null,
					"phone_number": null
				}
		}
		`
		test(allNulls, "")

		test(`{
			"phone_number": "+85298765432",
			"password": {},
			"mfa": {
				"password": {},
				"totp": {}
			}
		}`, `invalid request body:
/mfa/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]
/mfa/totp: required
  map[actual:<nil> expected:[secret] missing:[secret]]
/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]`)
	})

	Convey("Record JSON Schema for preferred_username", t, func() {
		test := func(recordString string, errorString string) {
			var record Record
			r, _ := http.NewRequest("POST", "/", strings.NewReader(recordString))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			err := httputil.BindJSONBody(r, w, RecordSchemaForIdentifierPreferredUsername.Validator(), &record)
			if errorString == "" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, errorString)
			}
		}
		test(recordString, "")

		test(`{}`, `invalid request body:
<root>: required
  map[actual:<nil> expected:[preferred_username] missing:[preferred_username]]`)

		allNulls := `
		{
			"preferred_username": "johndoe",

				"name": null,
				"given_name": null,
				"family_name": null,
				"middle_name": null,
				"nickname": null,
				"profile": null,
				"picture": null,
				"website": null,
				"gender": null,
				"birthdate": null,
				"zoneinfo": null,
				"locale": null,
				"address": null,

				"mfa": {
					"email": null,
					"phone_number": null
				}
		}
		`
		test(allNulls, "")

		test(`{
			"preferred_username": "johndoe",
			"password": {},
			"mfa": {
				"password": {},
				"totp": {}
			}
		}`, `invalid request body:
/mfa/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]
/mfa/totp: required
  map[actual:<nil> expected:[secret] missing:[secret]]
/password: required
  map[actual:<nil> expected:[password_hash type] missing:[password_hash type]]`)
	})
}
