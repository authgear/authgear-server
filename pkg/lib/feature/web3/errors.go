package web3

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidEndpoint = apierrors.NotFound.WithReason("InvalidURL").New("invalid endpoint")
