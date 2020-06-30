package handler

import "github.com/authgear/authgear-server/pkg/auth/config"

type ScopesValidator func(client config.OAuthClientConfig, scopes []string) error
type TokenGenerator func() string
