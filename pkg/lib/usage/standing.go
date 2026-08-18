package usage

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// EffectiveStandingUsageLimit is the standing (period-less) counterpart to
// EffectiveUsageLimit — its "usage" is a live count (e.g. COUNT(*) of DCR
// clients), not a periodically-reset Redis counter.
type EffectiveStandingUsageLimit struct {
	Name   model.UsageName
	Quota  int
	Action model.UsageLimitAction
}

func (l *Limiter) effectiveStandingUsageLimits(name model.UsageName) []EffectiveStandingUsageLimit {
	var limits []EffectiveStandingUsageLimit
	if l.EffectiveConfig.FeatureConfig.Usage != nil && l.EffectiveConfig.FeatureConfig.Usage.Limits != nil {
		for _, limit := range l.EffectiveConfig.FeatureConfig.Usage.Limits.StandingLimits(name) {
			limits = append(limits, EffectiveStandingUsageLimit{Name: name, Quota: limit.Quota, Action: limit.Action})
		}
	}
	// Deliberately no AppConfig.Usage.Limits lookup here: standing limits are
	// never project-editable (docs/plans/dcr/2026-08-17-05-client-usage-limit.md §2.3).
	return limits
}

func (l *Limiter) minBlockStandingQuota(limits []EffectiveStandingUsageLimit) (int, bool) {
	minQuota := 0
	found := false
	for _, limit := range limits {
		if limit.Action != model.UsageLimitActionBlock {
			continue
		}
		if !found || limit.Quota < minQuota {
			minQuota = limit.Quota
			found = true
		}
	}
	return minQuota, found
}

func crossedStandingUsageLimits(before, after int, limits []EffectiveStandingUsageLimit) []EffectiveStandingUsageLimit {
	var crossed []EffectiveStandingUsageLimit
	for _, limit := range limits {
		if before < limit.Quota && after >= limit.Quota {
			crossed = append(crossed, limit)
		}
	}
	return crossed
}

// CheckStanding returns ErrStandingUsageLimitExceeded if creating one more
// record of this usage name, on top of currentCount, would exceed the
// configured block quota. It does not create or reserve anything — the
// caller creates the record itself only if this returns nil.
func (l *Limiter) CheckStanding(ctx context.Context, name model.UsageName, currentCount int) error {
	limits := l.effectiveStandingUsageLimits(name)
	if blockQuota, ok := l.minBlockStandingQuota(limits); ok && currentCount+1 > blockQuota {
		return ErrStandingUsageLimitExceeded(name)
	}
	return nil
}

// ReportStandingCreated fires alert/hook/event triggers for any quota
// crossed by the creation that just succeeded. Call only after the new
// record has actually been committed, with the count observed immediately
// before that creation — mirrors Reserve's before/after crossing detection
// against a live COUNT(*) instead of a Redis counter.
func (l *Limiter) ReportStandingCreated(ctx context.Context, name model.UsageName, countBeforeCreate int) {
	limits := l.effectiveStandingUsageLimits(name)
	for _, standing := range crossedStandingUsageLimits(countBeforeCreate, countBeforeCreate+1, limits) {
		// Period is empty for a standing limit; UsageAlertPayload.Period
		// becomes the zero value model.UsageLimitPeriod(""), which any
		// consumer of the usage.alert.triggered payload must treat as "not
		// applicable" for this usage name.
		_ = l.maybeDispatchUsageAlert(ctx, EffectiveUsageLimit{
			Name:   standing.Name,
			Quota:  standing.Quota,
			Period: "",
			Action: standing.Action,
		}, countBeforeCreate+1)
	}
}
