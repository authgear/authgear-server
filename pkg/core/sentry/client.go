package sentry

import (
	"os"

	"github.com/getsentry/sentry-go"
)

var DefaultClient *Client

func init() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		DefaultClient = &Client{
			Hub:   sentry.CurrentHub(),
			Scope: sentry.CurrentHub().Scope(),
		}
		return
	}

	client, err := NewClient(dsn)
	if err != nil {
		panic(err)
	}
	DefaultClient = client
}

type Client struct {
	Scope *sentry.Scope
	Hub   *sentry.Hub
}

func NewClient(dsn string) (*Client, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		return nil, err
	}
	scope := sentry.NewScope()
	hub := sentry.NewHub(client, sentry.NewScope())
	return &Client{Scope: scope, Hub: hub}, nil
}
