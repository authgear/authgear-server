package password

import (
	"golang.org/x/crypto/bcrypt"
)

type bcryptPassword struct{}

var _ passwordFormat = bcryptPassword{}

func (bcryptPassword) ID() string {
	return "bcrypt"
}

func (bcryptPassword) Hash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func (bcryptPassword) Compare(password, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}

func (bcryptPassword) CheckHash(hash []byte) error {
	// The package bcrypt only exposes 3 functions.
	// The only functions that can be used to implement CheckHash is Cost.
	_, err := bcrypt.Cost(hash)
	return err
}
