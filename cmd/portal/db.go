package main

import (
	"errors"
	"os"
)

var DatabaseURL string
var DatabaseSchema string

func loadDBCredentials() (dbURL string, dbSchema string, err error) {
	if DatabaseURL == "" {
		DatabaseURL = os.Getenv("DATABASE_URL")
	}
	if DatabaseSchema == "" {
		DatabaseSchema = os.Getenv("DATABASE_SCHEMA")
	}

	if DatabaseURL == "" {
		return "", "", errors.New("missing database URL")
	}
	if DatabaseSchema == "" {
		return "", "", errors.New("missing database schema")
	}
	return DatabaseURL, DatabaseSchema, nil
}
