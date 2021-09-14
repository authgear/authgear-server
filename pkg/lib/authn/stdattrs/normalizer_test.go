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

		// Omit invalid locale
		test(T{
			"locale": "NONSENSE",
		}, T{})
		test(T{
			"locale": "en",
		}, T{
			"locale": "en",
		})

		// Omit empty names
		test(T{
			"name":        "",
			"family_name": "",
			"given_name":  "",
		}, T{})
		test(T{
			"name":        "John Doe",
			"family_name": "Doe",
			"given_name":  "John",
		}, T{
			"name":        "John Doe",
			"family_name": "Doe",
			"given_name":  "John",
		})

		// Omit invalid URLs
		test(T{
			"picture": "NONSENSE",
			"profile": "NONSENSE",
		}, T{})
		test(T{
			"picture": "http://example.com",
			"profile": "http://example.com",
		}, T{
			"picture": "http://example.com",
			"profile": "http://example.com",
		})
	})
}
