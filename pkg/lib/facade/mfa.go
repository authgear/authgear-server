package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
)

type MFAFacade struct {
	Coordinator *Coordinator
}

func (f *MFAFacade) GenerateDeviceToken(ctx context.Context) string {
	return f.Coordinator.MFAGenerateDeviceToken(ctx)
}

func (f *MFAFacade) CreateDeviceToken(ctx context.Context, userID string, token string) (*mfa.DeviceToken, error) {
	return f.Coordinator.MFACreateDeviceToken(ctx, userID, token)
}

func (f *MFAFacade) VerifyDeviceToken(ctx context.Context, userID string, token string) error {
	return f.Coordinator.MFAVerifyDeviceToken(ctx, userID, token)
}

func (f *MFAFacade) InvalidateAllDeviceTokens(ctx context.Context, userID string) error {
	return f.Coordinator.MFAInvalidateAllDeviceTokens(ctx, userID)
}

func (f *MFAFacade) VerifyRecoveryCode(ctx context.Context, userID string, code string) (*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAVerifyRecoveryCode(ctx, userID, code)
}

func (f *MFAFacade) ConsumeRecoveryCode(ctx context.Context, rc *mfa.RecoveryCode) error {
	return f.Coordinator.MFAConsumeRecoveryCode(ctx, rc)
}

func (f *MFAFacade) GenerateRecoveryCodes(ctx context.Context) []string {
	return f.Coordinator.MFAGenerateRecoveryCodes(ctx)
}

func (f *MFAFacade) ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAReplaceRecoveryCodes(ctx, userID, codes)
}

func (f *MFAFacade) ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAListRecoveryCodes(ctx, userID)
}
