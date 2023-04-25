package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const bucketNameKey = "bucket_name"

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrRateLimited(bucketName string) error {
	return RateLimited.NewWithInfo("request rate limited", apierrors.Details{
		bucketNameKey: bucketName,
	})
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
