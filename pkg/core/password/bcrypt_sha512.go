package password

import (
	"crypto/sha512"

	"golang.org/x/crypto/bcrypt"
)

type bcryptSHA512Password struct{}

var _ passwordFormat = bcryptSHA512Password{}

func (p bcryptSHA512Password) ID() string {
	return "bcrypt-sha512"
}

func (p bcryptSHA512Password) Hash(password []byte) ([]byte, error) {
	shaHash := sha512.Sum512(password)
	h, err := bcrypt.GenerateFromPassword(shaHash[:], bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return constructPasswordFormat([]byte(p.ID()), h), nil
}

func (p bcryptSHA512Password) Compare(password, hash []byte) error {
	_, data, err := parsePasswordFormat(hash)
	if err != nil {
		return err
	}
	shaHash := sha512.Sum512(password)
	return bcrypt.CompareHashAndPassword(data, shaHash[:])
}
