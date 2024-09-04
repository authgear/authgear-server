package stdattrs

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func makeValid() T {
	return T{
		"name":               "John Doe",
		"given_name":         "John",
		"family_name":        "Doe",
		"middle_name":        "Wick",
		"nickname":           "John",
		"preferred_username": "johndoe",
		"profile":            "https://example.com/profile",
		"picture":            "https://example.com/picture",
		"website":            "https://example.com/website",
		"email":              "johndoe@example.com",
		"gender":             "other",
		"birthdate":          "1990-01-01",
		"zoneinfo":           "Asia/Hong_Kong",
		"locale":             "zh-Hant-HK",
		"phone_number":       "+85298765432",
		"address": map[string]interface{}{
			"formatted":      "1 Unnamed Road, Central, Hong Kong Island, HK",
			"street_address": "1 Unnamed Road",
			"locality":       "Central",
			"region":         "Hong Kong Island",
			"postal_code":    "N/A",
			"country":        "HK",
		},
	}
}

func TestSchema(t *testing.T) {
	Convey("Schema", t, func() {
		expected := `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": {
			"type": "string",
			"format": "email"
		},
		"phone_number": {
			"type": "string",
			"format": "phone"
		},
		"preferred_username": {
			"type": "string",
			"minLength": 1
		},
		"family_name": {
			"type": "string",
			"minLength": 1
		},
		"given_name": {
			"type": "string",
			"minLength": 1
		},
		"middle_name": {
			"type": "string",
			"minLength": 1
		},
		"name": {
			"type": "string",
			"minLength": 1
		},
		"nickname": {
			"type": "string",
			"minLength": 1
		},
		"picture": {
			"type": "string",
			"format": "x_picture"
		},
		"profile": {
			"type": "string",
			"format": "uri"
		},
		"website": {
			"type": "string",
			"format": "uri"
		},
		"gender": {
			"type": "string",
			"minLength": 1
		},
		"birthdate": {
			"type": "string",
			"format": "birthdate"
		},
		"zoneinfo": {
			"type": "string",
			"format": "timezone"
		},
		"locale": {
			"type": "string",
			"format": "bcp47"
		},
		"address": {
			"type": "object",
			"properties": {
				"formatted": {
					"type": "string",
					"minLength": 1
				},
				"street_address": {
					"type": "string",
					"minLength": 1
				},
				"locality": {
					"type": "string",
					"minLength": 1
				},
				"region": {
					"type": "string",
					"minLength": 1
				},
				"postal_code": {
					"type": "string",
					"minLength": 1
				},
				"country": {
					"type": "string",
					"minLength": 1
				}
			}
		}
	}
}
`
		bytes, err := json.Marshal(schemaBuilder)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, expected)
	})
}

func TestValidate(t *testing.T) {
	Convey("Validate", t, func() {
		test := func(input T, expected error) {
			err := Validate(input)
			if expected == nil {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldResemble, expected)
			}
		}

		test(makeValid(), nil)

		// Extra properties
		extra := makeValid()
		extra["foobar"] = 42
		test(extra, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/foobar",
				},
			},
		})

		// Empty attrs is valid.
		test(T{}, nil)

		// Empty string is invalid.
		test(T{"name": ""}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/name",
					Keyword:  "minLength",
					Info: map[string]interface{}{
						"actual":   0.0,
						"expected": 1.0,
					},
				},
			},
		})

		// invalid email
		test(T{"email": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/email",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  "invalid email address: mail: missing '@' or angle-addr",
						"format": "email",
					},
				},
			},
		})

		// invalid phone_number
		test(T{"phone_number": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/phone_number",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  "not in E.164 format",
						"format": "phone",
					},
				},
			},
		})

		// invalid URL
		test(T{"picture": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/picture",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  "invalid scheme: ",
						"format": "x_picture",
					},
				},
			},
		})

		// invalid birthdate
		test(T{"birthdate": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/birthdate",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  `invalid birthdate: "invalid"`,
						"format": "birthdate",
					},
				},
			},
		})

		// invalid zoneinfo
		test(T{"zoneinfo": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/zoneinfo",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  `valid timezone name has at least 1 slash: "invalid"`,
						"format": "timezone",
					},
				},
			},
		})

		// invalid locale
		test(T{"locale": "invalid"}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/locale",
					Keyword:  "format",
					Info: map[string]interface{}{
						"error":  "invalid BCP 47 tag: language: tag is not well-formed",
						"format": "bcp47",
					},
				},
			},
		})

		// invalid address
		test(T{"address": 1}, &validation.AggregatedError{
			Message: "invalid value",
			Errors: []validation.Error{
				{
					Location: "/address",
					Keyword:  "type",
					Info: map[string]interface{}{
						"actual":   []interface{}{"integer", "number"},
						"expected": []interface{}{"object"},
					},
				},
			},
		})
	})
}
