package password

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestTranslateBcryptError(t *testing.T) {
	Convey("TranslateBcryptError", t, func() {
		test := func(err error, expected *apierrors.APIError) {
			So(apierrors.AsAPIError(TranslateBcryptError(err)), ShouldResemble, expected)
		}

		test(nil, nil)
		test(bcrypt.ErrHashTooShort, &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:    400,
			Message: "crypto/bcrypt: hashedSecret too short to be a bcrypted password",
			Info:    make(map[string]interface{}),
		})
		test(bcrypt.ErrPasswordTooLong, &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:    400,
			Message: "bcrypt: password length exceeds 72 bytes",
			Info:    make(map[string]interface{}),
		})
		test(bcrypt.HashVersionTooNewError('3'), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:    400,
			Message: "crypto/bcrypt: bcrypt algorithm version '3' requested is newer than current version '2'",
			Info:    make(map[string]interface{}),
		})
		test(bcrypt.InvalidHashPrefixError('#'), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:    400,
			Message: "crypto/bcrypt: bcrypt hashes must start with '$', but hashedSecret started with '#'",
			Info:    make(map[string]interface{}),
		})
		test(bcrypt.InvalidCostError(100), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:    400,
			Message: "crypto/bcrypt: cost 100 is outside allowed range (4,31)",
			Info:    make(map[string]interface{}),
		})
		test(fmt.Errorf("something else"), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "InternalError",
				Reason: "UnexpectedError",
			},
			Code:    500,
			Message: "unexpected error occurred",
			Info:    make(map[string]interface{}),
		})
	})
}
