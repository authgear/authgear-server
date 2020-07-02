package flows

import "github.com/authgear/authgear-server/pkg/core/skyerr"

var UnsupportedConfiguration = skyerr.Forbidden.WithReason("UnsupportedConfiguration")

var ErrUnsupportedConfiguration = UnsupportedConfiguration.New(
	"this operation is not supported by app configuration",
)

var ErrAnonymousDisabled = UnsupportedConfiguration.New(
	"anonymous user is disabled by configuration",
)
