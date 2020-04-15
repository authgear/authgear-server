package welcemail

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewDefaultSender,
)
