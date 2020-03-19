package urlprefix

import "github.com/google/wire"

var DependencySet = wire.NewSet(NewProvider)
