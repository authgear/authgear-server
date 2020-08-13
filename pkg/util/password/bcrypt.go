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
