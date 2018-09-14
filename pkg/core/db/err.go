package db

import (
	"net"

	"github.com/lib/pq"
)

func IsForeignKeyViolated(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
		return true
	}

	return false
}

func IsUniqueViolated(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}

	return false
}

func IsInvalidInputSyntax(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && (pqErr.Code == "22P02" || pqErr.Code == "22P03")
}

func IsUndefinedTable(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P01" {
		return true
	}

	return false
}

func IsNetworkError(err error) bool {
	_, ok := err.(*net.OpError)
	return ok
}
