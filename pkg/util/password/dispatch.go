package password

import "errors"

var latestFormat passwordFormat

var defaultFormat passwordFormat
var supportedFormats map[string]passwordFormat

var ErrTooLong = errors.New("password is too long")

func init() {
	latestFormat = bcryptSHA512Password{}

	defaultFormat = bcryptPassword{}
	supportedFormats = map[string]passwordFormat{}
	for _, fmt := range []passwordFormat{
		bcryptSHA512Password{},
	} {
		supportedFormats[fmt.ID()] = fmt
	}
}

func resolveFormat(hash []byte) (passwordFormat, error) {
	id, _, err := parsePasswordFormat(hash)
	if err != nil {
		return nil, err
	}

	fmt, ok := supportedFormats[string(id)]
	if ok {
		return fmt, nil
	}
	return defaultFormat, nil
}

func Hash(password []byte) ([]byte, error) {
	// Reject if new password is too long
	if len(password) > MaxLength {
		return nil, ErrTooLong
	}

	return latestFormat.Hash(password)
}

func Compare(password, hash []byte) error {
	// Do not enforce password length limit: we do not want users
	// to be locked out.

	fmt, err := resolveFormat(hash)
	if err != nil {
		return err
	}
	return fmt.Compare(password, hash)
}

func TryMigrate(password []byte, hash *[]byte) (migrated bool, err error) {
	// Do not enforce password length limit: migration of old password should
	// not fail due to length limit

	fmt, err := resolveFormat(*hash)
	if err != nil {
		return
	}
	if fmt.ID() == latestFormat.ID() {
		return
	}
	newHash, err := latestFormat.Hash(password)
	if err != nil {
		return
	}

	*hash = newHash
	migrated = true
	return
}
