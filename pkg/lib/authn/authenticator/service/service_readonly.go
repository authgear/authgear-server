package service

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

// RateLimiter depends on EventService
// EventService depends on UserInfoService
// So finally depends on authenticator.Service causing circular dependency
// This service was created for read only methods and do not depends on RateLimiter to break this circular dependency
type ReadOnlyService struct {
	Store    *Store
	Password PasswordAuthenticatorProvider
	Passkey  PasskeyAuthenticatorProvider
	TOTP     TOTPAuthenticatorProvider
	OOBOTP   OOBOTPAuthenticatorProvider
}

func (s *ReadOnlyService) List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	infosByUserID, err := s.ListByUserIDs(ctx, []string{userID}, filters...)
	if err != nil {
		return nil, err
	}

	infos, ok := infosByUserID[userID]

	if !ok || len(infos) == 0 {
		return []*authenticator.Info{}, nil
	}

	return infos, nil
}

// nolint:gocognit
func (s *ReadOnlyService) ListByUserIDs(ctx context.Context, userIDs []string, filters ...authenticator.Filter) (map[string][]*authenticator.Info, error) {
	refs, err := s.Store.ListRefsByUsers(ctx, userIDs, nil, nil)
	if err != nil {
		return nil, err
	}
	refsByType := map[model.AuthenticatorType]([]*authenticator.Ref){}

	for _, ref := range refs {
		arr := refsByType[ref.Type]
		arr = append(arr, ref)
		refsByType[ref.Type] = arr
	}

	extractIDs := func(authenticatorRefs []*authenticator.Ref) []string {
		ids := []string{}
		for _, r := range authenticatorRefs {
			ids = append(ids, r.ID)
		}
		return ids
	}

	infos := []*authenticator.Info{}

	// password
	if passwordRefs, ok := refsByType[model.AuthenticatorTypePassword]; ok && len(passwordRefs) > 0 {
		passwords, err := s.Password.GetMany(ctx, extractIDs(passwordRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passwords {
			infos = append(infos, i.ToInfo())
		}
	}

	// passkey
	if passkeyRefs, ok := refsByType[model.AuthenticatorTypePasskey]; ok && len(passkeyRefs) > 0 {
		passkeys, err := s.Passkey.GetMany(ctx, extractIDs(passkeyRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passkeys {
			infos = append(infos, i.ToInfo())
		}
	}

	// totp
	if totpRefs, ok := refsByType[model.AuthenticatorTypeTOTP]; ok && len(totpRefs) > 0 {
		totps, err := s.TOTP.GetMany(ctx, extractIDs(totpRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range totps {
			infos = append(infos, i.ToInfo())
		}
	}

	// oobotp
	oobotpRefs := []*authenticator.Ref{}
	if oobotpSMSRefs, ok := refsByType[model.AuthenticatorTypeOOBSMS]; ok && len(oobotpSMSRefs) > 0 {
		oobotpRefs = append(oobotpRefs, oobotpSMSRefs...)
	}
	if oobotpEmailRefs, ok := refsByType[model.AuthenticatorTypeOOBEmail]; ok && len(oobotpEmailRefs) > 0 {
		oobotpRefs = append(oobotpRefs, oobotpEmailRefs...)
	}
	if len(oobotpRefs) > 0 {
		oobotps, err := s.OOBOTP.GetMany(ctx, extractIDs(oobotpRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range oobotps {
			infos = append(infos, i.ToInfo())
		}
	}

	var filteredInfos []*authenticator.Info
	for _, a := range infos {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			filteredInfos = append(filteredInfos, a)
		}
	}

	infosByUserID := map[string][]*authenticator.Info{}
	for _, info := range filteredInfos {
		arr := infosByUserID[info.UserID]
		arr = append(arr, info)
		infosByUserID[info.UserID] = arr
	}

	return infosByUserID, nil
}
