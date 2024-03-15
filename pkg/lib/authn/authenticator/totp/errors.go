package totp

import (
	"encoding/base32"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidCode = errors.New("invalid code")

var InvalidTOTPSecret = apierrors.Invalid.WithReason("InvalidTOTPSecret")

func TranslateTOTPError(err error) error {
	if err == nil {
		return err
	}

	var corruptInputError base32.CorruptInputError
	if errors.As(err, &corruptInputError) {
		return InvalidTOTPSecret.New(corruptInputError.Error())
	}

	// Otherwise it is not a TOTP error.
	return err
}
