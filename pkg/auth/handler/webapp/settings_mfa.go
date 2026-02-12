package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
)

type SettingsMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
}
