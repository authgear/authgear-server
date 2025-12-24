package ratelimit

import (
	"context"
	"fmt"
)

type ratelimitWeightsContextKeyType struct{}

type Weights map[RateLimitGroup]float64

var ratelimitWeightsContextKey = ratelimitWeightsContextKeyType{}

type ratelimitWeightsContext struct {
	weights Weights
}

func WithRateLimitWeights(ctx context.Context) context.Context {
	return context.WithValue(ctx, ratelimitWeightsContextKey, &ratelimitWeightsContext{
		weights: nil,
	})
}

func SetRateLimitWeights(ctx context.Context, weights Weights) {
	wCtx, ok := ctx.Value(ratelimitWeightsContextKey).(*ratelimitWeightsContext)
	if !ok || wCtx == nil {
		panic(fmt.Errorf("trying to set rate limit weights but the context does not exist"))
	}
	wCtx.weights = weights
}

func getRateLimitWeights(ctx context.Context) Weights {
	wCtx, ok := ctx.Value(ratelimitWeightsContextKey).(*ratelimitWeightsContext)
	if !ok || wCtx == nil {
		return nil
	}
	return wCtx.weights
}
