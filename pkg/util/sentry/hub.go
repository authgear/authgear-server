package sentry

import (
	"github.com/getsentry/sentry-go"
)

func NewHub(dsn string) (*sentry.Hub, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: string(dsn), Debug: true})
	if err != nil {
		return nil, err
	}
	scope := sentry.NewScope()
	hub := sentry.NewHub(client, scope)
	return hub, nil
}
