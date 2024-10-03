package userimport

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrJobNotFound = apierrors.NotFound.WithReason("JobNotFound").New("job not found")
