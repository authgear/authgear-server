package handler

import "github.com/authgear/authgear-server/pkg/lib/config"

type ScopesValidator func(client *config.OAuthClientConfig, scopes []string) error
type TokenGenerator func() string
