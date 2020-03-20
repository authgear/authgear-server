package oidc

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
)

var DependencySet = wire.NewSet(
	wire.Value(handler.ScopesValidator(ValidateScopes)),
	wire.Struct(new(MetadataProvider)),
)
