package handler

import "github.com/skygeario/skygear-server/pkg/core/config"

type ScopesValidator func(client config.OAuthClientConfiguration, scopes []string) error
type TokenGenerator func() string
