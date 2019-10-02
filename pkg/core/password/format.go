package password

import (
	"bytes"
	"errors"
)

type passwordFormat interface {
	ID() string
	Hash(password []byte) ([]byte, error)
	Compare(password, hash []byte) error
}

var errInvalidPasswordFormat = errors.New("invalid password format")

func parsePasswordFormat(h []byte) (id []byte, data []byte, err error) {
	i := bytes.IndexByte(h, '$')
	if i != 0 {
		err = errInvalidPasswordFormat
		return
	}
	h = h[i+1:]

	i = bytes.IndexByte(h, '$')
	if i == -1 {
		err = errInvalidPasswordFormat
		return
	}

	id = h[:i]
	data = h[i+1:]
	return
}
