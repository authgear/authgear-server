package stdattrs

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

type mockNormalizer struct{}

func (n mockNormalizer) Normalize(loginID string) (string, error) {
	return loginID, nil
}

func (n mockNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

func TestNormalizer(t *testing.T) {
	Convey("Normalizer", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		factory := NewMockLoginIDNormalizerFactory(ctrl)
		n := &Normalizer{
			LoginIDNormalizerFactory: factory,
		}

		factory.EXPECT().NormalizerWithLoginIDType(gomock.Any()).AnyTimes().Return(mockNormalizer{})

		test := func(input T, output T) {
			err := n.Normalize(input)
			So(err, ShouldBeNil)
			So(input, ShouldResemble, output)
		}

		// Normalize email
		test(T{
			"email": "user@example.com",
		}, T{
			"email": "user@example.com",
		})

		// Normalize phone_number
		test(T{
			"phone_number": "+85298765432",
		}, T{
			"phone_number": "+85298765432",
		})

		// Normalize strings
		test(T{
			"name":               "",
			"given_name":         "",
			"family_name":        "",
			"middle_name":        "",
			"nickname":           "",
			"preferred_username": "",
			"gender":             "",
		}, T{})
		test(T{
			"name":               0,
			"given_name":         0,
			"family_name":        0,
			"middle_name":        0,
			"nickname":           0,
			"preferred_username": 0,
			"gender":             0,
		}, T{})
		test(T{
			"name":               "John Doe",
			"given_name":         "John",
			"family_name":        "Doe",
			"middle_name":        "Wick",
			"nickname":           "John",
			"preferred_username": "johndoe",
			"gender":             "other",
		}, T{
			"name":               "John Doe",
			"given_name":         "John",
			"family_name":        "Doe",
			"middle_name":        "Wick",
			"nickname":           "John",
			"preferred_username": "johndoe",
			"gender":             "other",
		})

		// Normalize bools
		test(T{
			"email_verified":        true,
			"phone_number_verified": true,
		}, T{
			"email_verified":        true,
			"phone_number_verified": true,
		})
		test(T{
			"email_verified":        "",
			"phone_number_verified": "",
		}, T{})

		// Normalize URLs
		test(T{
			"picture": "NONSENSE",
			"profile": "NONSENSE",
			"website": "NONSENSE",
		}, T{})
		test(T{
			"picture": "http://example.com",
			"profile": "http://example.com",
			"website": "http://example.com",
		}, T{
			"picture": "http://example.com",
			"profile": "http://example.com",
			"website": "http://example.com",
		})

		// Normalize birthdate
		test(T{
			"birthdate": "1990-01-01",
		}, T{
			"birthdate": "1990-01-01",
		})
		test(T{
			"birthdate": "0000-01-01",
		}, T{
			"birthdate": "0000-01-01",
		})
		test(T{
			"birthdate": "--01-01",
		}, T{
			"birthdate": "--01-01",
		})
		test(T{
			"birthdate": "1990",
		}, T{
			"birthdate": "1990",
		})
		test(T{
			"birthdate": "1990-01-01T00:00:00Z",
		}, T{})

		// Normalize locale
		test(T{
			"locale": "en",
		}, T{
			"locale": "en",
		})
		test(T{
			"locale": 1,
		}, T{})
		test(T{
			"locale": "NONSENSE",
		}, T{})

		// Normalize address
		test(T{
			"address": 1,
		}, T{})
		test(T{
			"address": map[string]interface{}{},
		}, T{})
		test(T{
			"address": map[string]interface{}{
				"formatted": "",
			},
		}, T{})
		test(T{
			"address": map[string]interface{}{
				"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
			},
		}, T{
			"address": map[string]interface{}{
				"formatted": "1 Unnamed Road, Central, Hong Kong Island, HK",
			},
		})
		test(T{
			"address": map[string]interface{}{
				"formatted":      "1 Unnamed Road, Central, Hong Kong Island, HK",
				"street_address": "1 Unnamed Road",
				"locality":       "Central",
				"region":         "Hong Kong Island",
				"postal_code":    "N/A",
				"country":        "HK",
			},
		}, T{
			"address": map[string]interface{}{
				"formatted":      "1 Unnamed Road, Central, Hong Kong Island, HK",
				"street_address": "1 Unnamed Road",
				"locality":       "Central",
				"region":         "Hong Kong Island",
				"postal_code":    "N/A",
				"country":        "HK",
			},
		})
	})
}
