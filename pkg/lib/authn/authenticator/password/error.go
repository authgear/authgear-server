package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidBcryptHash = apierrors.Invalid.WithReason("InvalidBcryptHash")

func TranslateBcryptError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, bcrypt.ErrHashTooShort) {
		return InvalidBcryptHash.New(err.Error())
	}

	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return InvalidBcryptHash.New(err.Error())
	}

	var version bcrypt.HashVersionTooNewError
	if errors.As(err, &version) {
		return InvalidBcryptHash.New(version.Error())
	}

	var prefix bcrypt.InvalidHashPrefixError
	if errors.As(err, &prefix) {
		return InvalidBcryptHash.New(prefix.Error())
	}

	var cost bcrypt.InvalidCostError
	if errors.As(err, &cost) {
		return InvalidBcryptHash.New(cost.Error())
	}

	// Otherwise it is not a bcrypt error.
	return err
}
