package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrRoleDuplicateKey = apierrors.BadRequest.WithReason("RoleDuplicateKey").New("duplicate role key")
