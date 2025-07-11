package databaseutil

import (
	"errors"

	"github.com/lib/pq"
)

func IsDuplicateKeyError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) {
		// 23505 is unique_violation
		if pqError.Code == "23505" {
			return true
		}
	}
	return false
}
