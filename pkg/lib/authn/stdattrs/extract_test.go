package stdattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func makeFull() T {
	return T{
		"name":                  "John Doe",
		"given_name":            "John",
		"family_name":           "Doe",
		"middle_name":           "Wick",
		"nickname":              "John",
		"preferred_username":    "johndoe",
		"profile":               "https://example.com/profile",
		"picture":               "https://example.com/picture",
		"website":               "https://example.com/website",
		"email":                 "johndoe@example.com",
		"email_verified":        true,
		"gender":                "other",
		"birthdate":             "1990-01-01",
		"zoneinfo":              "Asia/Hong_Kong",
		"locale":                "zh-Hant-HK",
		"phone_number":          "+85298765432",
		"phone_number_verified": true,
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

func TestExtract(t *testing.T) {
	Convey("Extract", t, func() {
		test := func(input T, expected T) {
			actual := Extract(input)
			So(actual, ShouldResemble, expected)
		}

		test(T{}, T{})

		full := makeFull()
		test(full, full)

		fullWithExtra := makeFull()
		fullWithExtra["foobar"] = "foobar"
		test(fullWithExtra, full)

		test(T{"name": ""}, T{})
		test(T{"address": map[string]interface{}{}}, T{})
		test(T{
			"address": map[string]interface{}{
				"formatted": "some address",
			},
		}, T{
			Address: map[string]interface{}{
				Formatted: "some address",
			},
		})
	})
}
