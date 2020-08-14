package identity

import "errors"

var ErrIdentityNotFound = errors.New("identity not found")

var ErrIdentityAlreadyExists = errors.New("identity already exists")
