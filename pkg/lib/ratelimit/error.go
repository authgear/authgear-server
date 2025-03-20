package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const bucketNameKey = "bucket_name"

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrRateLimited(bucketName BucketName) error {
	return RateLimited.NewWithInfo("request rate limited", apierrors.Details{
		bucketNameKey: bucketName,
	})
}

func IsRateLimitErrorWithBucketName(err error, bucketName BucketName) bool {
	if !apierrors.IsKind(err, RateLimited) {
		return false
	}

	apiError := apierrors.AsAPIError(err)
	if apiError == nil {
		return false
	}

	return apiError.Info_ReadOnly[bucketNameKey] == bucketName
}
