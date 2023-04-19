package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const bucketNameKey = "bucket_name"

// ErrUsageLimitExceeded is deprecated; see usage.ErrUsageLimitExceeded.
var ErrUsageLimitExceeded = apierrors.ServiceUnavailable.WithReason("UsageLimitExceeded").
	New("usage limit exceeded")

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrTooManyRequestsFrom(bucket Bucket) error {
	errMsg := "request rate limited"
	if bucket.Name == "" {
		return RateLimited.New(errMsg)
	} else {
		return RateLimited.NewWithInfo(errMsg, apierrors.Details{
			bucketNameKey: bucket.Name,
		})
	}
}

func IsRateLimitErrorWithBucketName(err error, bucketName string) bool {
	if !apierrors.IsKind(err, RateLimited) {
		return false
	}

	apiError := apierrors.AsAPIError(err)
	if apiError == nil {
		return false
	}

	return apiError.Info[bucketNameKey] == bucketName
}
