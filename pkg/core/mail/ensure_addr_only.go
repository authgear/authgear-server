package mail

import (
	"errors"

	gomail "net/mail"
)

var ErrAddressWithName = errors.New("address must not have name")
var ErrAddressNotSameAsInput = errors.New("formatted address is not the same as input")

// EnsureAddressOnly ensures the given string is address only.
func EnsureAddressOnly(s string) error {
	addr, err := gomail.ParseAddress(s)
	if err != nil {
		return err
	}
	if addr.Name != "" {
		return ErrAddressWithName
	}
	ss := addr.String()
	// Remove <>
	ss = ss[1 : len(ss)-1]
	if s != ss {
		return ErrAddressNotSameAsInput
	}
	return nil
}
