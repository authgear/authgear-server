package redisqueue

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrTaskNotFound = apierrors.NotFound.WithReason("TaskNotFound").New("task not found")
