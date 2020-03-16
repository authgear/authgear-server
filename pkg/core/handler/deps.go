package handler

import "github.com/google/wire"

var DependencySet = wire.NewSet(NewRequireAuthzFactory)
