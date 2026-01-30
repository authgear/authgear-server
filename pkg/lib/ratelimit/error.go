package ratelimit

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const rateLimitKey = "rate_limit"
const DEPRECATED_rateLimitNameKey = "rate_limit_name"
const DEPRECATED_bucketNameKey = "bucket_name"

var RateLimited = apierrors.TooManyRequest.WithReason("RateLimited")

func ErrRateLimited(rl RateLimitName, rlgroup RateLimitGroup, bucketName BucketName) error {
	details := apierrors.Details{
		// Deprecated field. Do not use.
		// Use rate_limit_name instead.
		DEPRECATED_bucketNameKey: bucketName,
	}
	// Some buckets do not have a rate limit name, so do not add the key if it is empty
	if rl != "" {
		details[rateLimitKey] = model.RateLimit{
			Name:  string(rl),
			Group: string(rlgroup),
		}
		details[DEPRECATED_rateLimitNameKey] = rlgroup
	}
	return RateLimited.NewWithInfo("request rate limited", details)
}

func IsRateLimitErrorWithBucketName(err error, bucketName BucketName) bool {
	return apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
		return e.Kind == RateLimited && e.Info_ReadOnly[DEPRECATED_bucketNameKey] == bucketName
	})
}
