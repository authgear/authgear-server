package lockout

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const bucketNameKey = "bucket_name"
const untilKey = "until"

var ErrLocked = apierrors.TooManyRequest.WithReason("Locked")

func NewErrLocked(bucketName BucketName, until time.Time) error {
	return ErrLocked.NewWithInfo("resource locked", apierrors.Details{
		bucketNameKey: bucketName,
		untilKey:      until,
	})
}
