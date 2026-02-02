package password

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestTranslateBcryptError(t *testing.T) {
	Convey("TranslateBcryptError", t, func() {
		test := func(err error, expected *apierrors.APIError) {
			So(apierrors.AsAPIErrorWithContext(context.Background(), TranslateBcryptError(err)), ShouldResemble, expected)
		}

		test(nil, nil)
		test(bcrypt.ErrHashTooShort, &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:          400,
			Message:       "crypto/bcrypt: hashedSecret too short to be a bcrypted password",
			Info_ReadOnly: make(map[string]interface{}),
		})
		test(bcrypt.ErrPasswordTooLong, &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:          400,
			Message:       "bcrypt: password length exceeds 72 bytes",
			Info_ReadOnly: make(map[string]interface{}),
		})
		test(bcrypt.HashVersionTooNewError('3'), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:          400,
			Message:       "crypto/bcrypt: bcrypt algorithm version '3' requested is newer than current version '2'",
			Info_ReadOnly: make(map[string]interface{}),
		})
		test(bcrypt.InvalidHashPrefixError('#'), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:          400,
			Message:       "crypto/bcrypt: bcrypt hashes must start with '$', but hashedSecret started with '#'",
			Info_ReadOnly: make(map[string]interface{}),
		})
		test(bcrypt.InvalidCostError(100), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "Invalid",
				Reason: "InvalidBcryptHash",
			},
			Code:          400,
			Message:       "crypto/bcrypt: cost 100 is outside allowed inclusive range 4..31",
			Info_ReadOnly: make(map[string]interface{}),
		})
		test(fmt.Errorf("something else"), &apierrors.APIError{
			Kind: apierrors.Kind{
				Name:   "InternalError",
				Reason: "UnexpectedError",
			},
			Code:          500,
			Message:       "unexpected error occurred",
			Info_ReadOnly: make(map[string]interface{}),
		})
	})
}
