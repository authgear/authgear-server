package facade

import "github.com/authgear/authgear-server/pkg/lib/authn/mfa"

type MFAFacade struct {
	Coordinator *Coordinator
}

func (f *MFAFacade) GenerateDeviceToken() string {
	return f.Coordinator.MFAGenerateDeviceToken()
}

func (f *MFAFacade) CreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error) {
	return f.Coordinator.MFACreateDeviceToken(userID, token)
}

func (f *MFAFacade) VerifyDeviceToken(userID string, token string) error {
	return f.Coordinator.MFAVerifyDeviceToken(userID, token)
}

func (f *MFAFacade) InvalidateAllDeviceTokens(userID string) error {
	return f.Coordinator.MFAInvalidateAllDeviceTokens(userID)
}

func (f *MFAFacade) VerifyRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAVerifyRecoveryCode(userID, code)
}

func (f *MFAFacade) ConsumeRecoveryCode(rc *mfa.RecoveryCode) error {
	return f.Coordinator.MFAConsumeRecoveryCode(rc)
}

func (f *MFAFacade) GenerateRecoveryCodes() []string {
	return f.Coordinator.MFAGenerateRecoveryCodes()
}

func (f *MFAFacade) ReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAReplaceRecoveryCodes(userID, codes)
}

func (f *MFAFacade) ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error) {
	return f.Coordinator.MFAListRecoveryCodes(userID)
}
