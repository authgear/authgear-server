package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrRoleNotFound = apierrors.NotFound.WithReason("RoleNotFound").New("role not found")

var ErrRoleDuplicateKey = apierrors.BadRequest.WithReason("RoleDuplicateKey").New("duplicate role key")

var ErrGroupNotFound = apierrors.NotFound.WithReason("GroupNotFound").New("group not found")

var ErrGroupDuplicateKey = apierrors.BadRequest.WithReason("GroupDuplicateKey").New("duplicate group key")

var GroupUnknownKeys = apierrors.NotFound.WithReason("GroupUnknownKeys")
