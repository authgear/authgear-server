package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const rateLimitNameKey = "rate_limit_name"
const rateLimitGroupKey = "rate_limit_group"
const bucketNameKey = "bucket_name"

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrRateLimited(rl RateLimitName, rlgroup RateLimitGroup, bucketName BucketName) error {
	details := apierrors.Details{
		// Deprecated field. Do not use.
		// Use rate_limit_name instead.
		bucketNameKey: bucketName,
	}
	// Some buckets do not have a rate limit name, so do not add the key if it is empty
	if rl != "" {
		details[rateLimitNameKey] = rl
		details[rateLimitGroupKey] = rlgroup
	}
	return RateLimited.NewWithInfo("request rate limited", details)
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
