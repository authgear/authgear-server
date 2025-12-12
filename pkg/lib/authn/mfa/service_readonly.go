package mfa

import "context"

// RateLimiter depends on EventService
// EventService depends on UserInfoService
// So finally depends on mfa.Service causing circular dependency
// This service was created for read only methods and do not depends on RateLimiter to break this circular dependency
type ReadOnlyService struct {
	RecoveryCodes StoreRecoveryCode
}

func (s *ReadOnlyService) ListRecoveryCodes(ctx context.Context, userID string) ([]*RecoveryCode, error) {
	return s.RecoveryCodes.List(ctx, userID)
}
