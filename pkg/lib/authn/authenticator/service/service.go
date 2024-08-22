package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.Password, error)
	GetMany(ids []string) ([]*authenticator.Password, error)
	List(userID string) ([]*authenticator.Password, error)
	New(id string, userID string, passwordSpec *authenticator.PasswordSpec, isDefault bool, kind string) (*authenticator.Password, error)
	UpdatePassword(a *authenticator.Password, options *password.UpdatePasswordOptions) (bool, *authenticator.Password, error)
	Create(*authenticator.Password) error
	Update(*authenticator.Password) error
	Delete(*authenticator.Password) error
	Authenticate(a *authenticator.Password, password string) (verifyResult *password.VerifyResult, err error)
}

type PasskeyAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.Passkey, error)
	GetMany(ids []string) ([]*authenticator.Passkey, error)
	List(userID string) ([]*authenticator.Passkey, error)
	New(
		id string,
		userID string,
		attestationResponse []byte,
		isDefault bool,
		kind string,
	) (*authenticator.Passkey, error)
	Create(*authenticator.Passkey) error
	Update(*authenticator.Passkey) error
	Delete(*authenticator.Passkey) error
	Authenticate(a *authenticator.Passkey, assertionResponse []byte) (requireUpdate bool, err error)
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.TOTP, error)
	GetMany(ids []string) ([]*authenticator.TOTP, error)
	List(userID string) ([]*authenticator.TOTP, error)
	New(id string, userID string, totpSpec *authenticator.TOTPSpec, isDefault bool, kind string) (*authenticator.TOTP, error)
	Create(*authenticator.TOTP) error
	Delete(*authenticator.TOTP) error
	Authenticate(a *authenticator.TOTP, code string) error
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.OOBOTP, error)
	GetMany(ids []string) ([]*authenticator.OOBOTP, error)
	List(userID string) ([]*authenticator.OOBOTP, error)
	New(id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) (*authenticator.OOBOTP, error)
	UpdateTarget(a *authenticator.OOBOTP, option oob.UpdateTargetOption) (*authenticator.OOBOTP, bool)
	Create(*authenticator.OOBOTP) error
	Update(*authenticator.OOBOTP) error
	Delete(*authenticator.OOBOTP) error
}

type OTPCodeService interface {
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type Service struct {
	Store          *Store
	Config         *config.AppConfig
	Password       PasswordAuthenticatorProvider
	Passkey        PasskeyAuthenticatorProvider
	TOTP           TOTPAuthenticatorProvider
	OOBOTP         OOBOTPAuthenticatorProvider
	OTPCodeService OTPCodeService
	RateLimits     RateLimits
	Lockout        Lockout
}

func (s *Service) Get(id string) (*authenticator.Info, error) {
	ref, err := s.Store.GetRefByID(id)
	if err != nil {
		return nil, err
	}
	switch ref.Type {
	case model.AuthenticatorTypePassword:
		p, err := s.Password.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypePasskey:
		p, err := s.Passkey.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypeTOTP:
		t, err := s.TOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return t.ToInfo(), nil

	// FIXME(oob): handle getting different channel
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		o, err := s.OOBOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	}

	panic("authenticator: unknown authenticator type " + ref.Type)
}

func (s *Service) GetMany(ids []string) ([]*authenticator.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	var passwordIDs, passkeyIDs, totpIDs, oobIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case model.AuthenticatorTypePassword:
			passwordIDs = append(passwordIDs, ref.ID)
		case model.AuthenticatorTypePasskey:
			passkeyIDs = append(passkeyIDs, ref.ID)
		case model.AuthenticatorTypeTOTP:
			totpIDs = append(totpIDs, ref.ID)
		case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
			oobIDs = append(oobIDs, ref.ID)
		default:
			panic("authenticator: unknown authenticator type " + ref.Type)
		}
	}

	var infos []*authenticator.Info

	{
		p, err := s.Password.GetMany(passwordIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range p {
			infos = append(infos, a.ToInfo())
		}
	}
	{
		passkeys, err := s.Passkey.GetMany(passkeyIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range passkeys {
			infos = append(infos, a.ToInfo())
		}
	}

	{
		t, err := s.TOTP.GetMany(totpIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range t {
			infos = append(infos, a.ToInfo())
		}
	}

	{
		o, err := s.OOBOTP.GetMany(oobIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range o {
			infos = append(infos, a.ToInfo())
		}
	}

	return infos, nil
}

// nolint:gocognit
func (s *Service) ListByUserIDs(userIDs []string, filters ...authenticator.Filter) (map[string][]*authenticator.Info, error) {
	refs, err := s.Store.ListRefsByUsers(userIDs, nil, nil)
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
		passwords, err := s.Password.GetMany(extractIDs(passwordRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passwords {
			infos = append(infos, i.ToInfo())
		}
	}

	// passkey
	if passkeyRefs, ok := refsByType[model.AuthenticatorTypePasskey]; ok && len(passkeyRefs) > 0 {
		passkeys, err := s.Passkey.GetMany(extractIDs(passkeyRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passkeys {
			infos = append(infos, i.ToInfo())
		}
	}

	// totp
	if totpRefs, ok := refsByType[model.AuthenticatorTypeTOTP]; ok && len(totpRefs) > 0 {
		totps, err := s.TOTP.GetMany(extractIDs(totpRefs))
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
		oobotps, err := s.OOBOTP.GetMany(extractIDs(oobotpRefs))
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

func (s *Service) List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	infosByUserID, err := s.ListByUserIDs([]string{userID}, filters...)
	if err != nil {
		return nil, err
	}

	infos, ok := infosByUserID[userID]

	if !ok || len(infos) == 0 {
		return []*authenticator.Info{}, nil
	}

	return infos, nil
}

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string, authenticatorType *model.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error) {
	return s.Store.ListRefsByUsers(userIDs, authenticatorType, authenticatorKind)
}

func (s *Service) New(spec *authenticator.Spec) (*authenticator.Info, error) {
	return s.NewWithAuthenticatorID("", spec)
}

func (s *Service) NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	switch spec.Type {
	case model.AuthenticatorTypePassword:
		p, err := s.Password.New(authenticatorID, spec.UserID, spec.Password, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypePasskey:
		attestationResponse := spec.Passkey.AttestationResponse

		p, err := s.Passkey.New(
			authenticatorID,
			spec.UserID,
			attestationResponse,
			spec.IsDefault,
			string(spec.Kind),
		)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypeTOTP:
		t, err := s.TOTP.New(authenticatorID, spec.UserID, spec.TOTP, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return t.ToInfo(), nil

	case model.AuthenticatorTypeOOBEmail:
		email := spec.OOBOTP.Email
		o, err := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBEmail, email, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}

		return o.ToInfo(), nil
	case model.AuthenticatorTypeOOBSMS:
		phone := spec.OOBOTP.Phone
		o, err := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBSMS, phone, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil

	}

	panic("authenticator: unknown authenticator type " + spec.Type)
}

type UpdateOOBOTPTargetOption struct {
	Phone string
	Email string
}

func (o UpdateOOBOTPTargetOption) toProviderOptions() oob.UpdateTargetOption {
	return oob.UpdateTargetOption{
		Phone: o.Phone,
		Email: o.Email,
	}
}

func (s *Service) UpdateOOBOTPTarget(ai *authenticator.Info, option UpdateOOBOTPTargetOption) (*authenticator.Info, bool) {
	switch ai.Type {
	case model.AuthenticatorTypeOOBEmail:
		fallthrough
	case model.AuthenticatorTypeOOBSMS:
		a := ai.OOBOTP
		newAuth, changed := s.OOBOTP.UpdateTarget(a, option.toProviderOptions())
		return newAuth.ToInfo(), changed
	}

	panic("authenticator: update authenticator is not supported for type " + ai.Type)
}

type UpdatePasswordOptions struct {
	SetPassword    bool
	PlainPassword  string
	SetExpireAfter bool
	ExpireAfter    *time.Time
}

func (options *UpdatePasswordOptions) toProviderOptions() *password.UpdatePasswordOptions {
	return &password.UpdatePasswordOptions{
		SetPassword:    options.SetPassword,
		PlainPassword:  options.PlainPassword,
		SetExpireAfter: options.SetExpireAfter,
		ExpireAfter:    options.ExpireAfter,
	}
}

func (s *Service) UpdatePassword(ai *authenticator.Info, options *UpdatePasswordOptions) (bool, *authenticator.Info, error) {
	if ai.Type != model.AuthenticatorTypePassword {
		panic("authenticator: update password is not supported for type " + ai.Type)
	}

	a := ai.Password
	changed, newAuth, err := s.Password.UpdatePassword(a, options.toProviderOptions())
	if err != nil {
		return false, nil, err
	}
	return changed, newAuth.ToInfo(), nil
}

func (s *Service) Create(info *authenticator.Info) error {
	ais, err := s.List(info.UserID)
	if err != nil {
		return err
	}

	for _, a := range ais {
		if info.Equal(a) {
			err = authenticator.NewErrDuplicatedAuthenticator(info.Type)
			return err
		}
	}

	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypeTOTP:
		a := info.TOTP
		if err := s.TOTP.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		a := info.OOBOTP
		if err := s.OOBOTP.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	default:
		panic("authenticator: unknown authenticator type " + info.Type)
	}

	return nil
}

func (s *Service) Update(info *authenticator.Info) error {
	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.Update(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Update(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypeOOBEmail:
		a := info.OOBOTP
		if err := s.OOBOTP.Update(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypeOOBSMS:
		a := info.OOBOTP
		if err := s.OOBOTP.Update(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	default:
		panic("authenticator: unknown authenticator type for update" + info.Type)
	}

	return nil
}

func (s *Service) Delete(info *authenticator.Info) error {
	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeTOTP:
		a := info.TOTP
		if err := s.TOTP.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		a := info.OOBOTP
		if err := s.OOBOTP.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	default:
		panic("authenticator: delete authenticator is not supported yet for type " + info.Type)
	}

	return nil
}

func (s *Service) verifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (verifyResult *VerifyResult, err error) {
	verifyResult = &VerifyResult{}
	switch info.Type {
	case model.AuthenticatorTypePassword:
		plainPassword := spec.Password.PlainPassword
		a := info.Password
		verifyResult.Password, err = s.Password.Authenticate(a, plainPassword)
		if err != nil {
			err = api.ErrInvalidCredentials
			return nil, err
		}
		*info = *a.ToInfo()
		return
	case model.AuthenticatorTypePasskey:
		assertionResponse := spec.Passkey.AssertionResponse
		a := info.Passkey
		verifyResult.Passkey, err = s.Passkey.Authenticate(a, assertionResponse)
		if err != nil {
			err = api.ErrInvalidCredentials
			return nil, err
		}
		*info = *a.ToInfo()

		return
	case model.AuthenticatorTypeTOTP:
		code := spec.TOTP.Code
		a := info.TOTP
		if s.TOTP.Authenticate(a, code) != nil {
			err = api.ErrInvalidCredentials
			return nil, err
		}
		// Do not update info because by definition TOTP does not update itself during verification.

		return
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		var channel model.AuthenticatorOOBChannel
		if options.OOBChannel != nil {
			channel = *options.OOBChannel
		} else {
			switch info.Type {
			case model.AuthenticatorTypeOOBEmail:
				channel = model.AuthenticatorOOBChannelEmail
			case model.AuthenticatorTypeOOBSMS:
				channel = model.AuthenticatorOOBChannelSMS
			}
		}

		var kind otp.Kind

		if options.Form != "" {
			kind = otp.KindOOBOTPWithForm(s.Config, channel, options.Form)
		} else {
			panic("authenticator: form is required for OOBOTP")
		}

		code := spec.OOBOTP.Code
		a := info.OOBOTP
		err = s.OTPCodeService.VerifyOTP(kind, a.ToTarget(), code, &otp.VerifyOptions{
			UseSubmittedCode: options.UseSubmittedValue,
			UserID:           info.UserID,
		})
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			err = api.ErrInvalidCredentials
			return nil, err
		} else if err != nil {
			return nil, err
		}
		// Do not update info because by definition OOBOTP does not update itself during verification.

		return verifyResult, nil
	}

	panic("authenticator: unhandled authenticator type " + info.Type)
}

// Given a list of authenticators, try to verify one of them
func (s *Service) VerifyOneWithSpec(
	userID string,
	authenticatorType model.AuthenticatorType,
	infos []*authenticator.Info,
	spec *authenticator.Spec,
	options *VerifyOptions) (info *authenticator.Info, verifyResult *VerifyResult, err error) {
	if options == nil {
		options = &VerifyOptions{}
	}

	r := s.RateLimits.Reserve(userID, authenticatorType)
	defer s.RateLimits.Cancel(r)

	if err = r.Error(); err != nil {
		return
	}

	// Check if it is already locked
	err = s.Lockout.Check(userID)
	if err != nil {
		return
	}

	for _, thisInfo := range infos {
		if thisInfo.UserID != userID || thisInfo.Type != authenticatorType {
			// Ensure all authenticators are in same type of the same user
			err = fmt.Errorf("only authenticators with same type of same user can be verified together")
			return
		}
		verifyResult, err = s.verifyWithSpec(thisInfo, spec, options)
		if errors.Is(err, api.ErrInvalidCredentials) {
			continue
		}
		// unexpected errors or no error
		// For both cases we should break the loop and return
		if err == nil {
			info = thisInfo
		}
		break
	}

	switch {
	case info == nil && err == nil:
		// If we reach here, it means infos is empty.
		// Here is one case that infos is empty.
		// The end-user remove their passkey in Authgear, but keep the passkey in their browser.
		// Authgear will see an passkey that it does not know.
		err = api.ErrInvalidCredentials
	case info != nil && err == nil:
		// Authenticated.
		break
	case info == nil && err != nil:
		// Some error.
		break
	default:
		panic(fmt.Errorf("unexpected post condition: info != nil && err != nil"))
	}

	// If error is ErrInvalidCredentials, consume rate limit token and increment lockout attempt
	if errors.Is(err, api.ErrInvalidCredentials) {
		r.Consume()
		lockErr := s.Lockout.MakeAttempt(userID, authenticatorType)
		if lockErr != nil {
			err = errors.Join(lockErr, err)
			return
		}
		return
	}
	// else, simply return the error if any
	return
}

func (s *Service) UpdateOrphans(oldInfo *identity.Info, newInfo *identity.Info) error {
	authenticators, err := s.List(oldInfo.UserID)
	if err != nil {
		return err
	}

	for _, a := range authenticators {
		if a.IsDependentOf(oldInfo) {
			newAuth, changed := s.UpdateOOBOTPTarget(a, UpdateOOBOTPTargetOption{
				Email: newInfo.LoginID.LoginID,
				Phone: newInfo.LoginID.LoginID,
			})
			if changed {
				err = s.Update(newAuth)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) RemoveOrphans(identities []*identity.Info) error {
	if len(identities) == 0 {
		return nil
	}

	authenticators, err := s.List(identities[0].UserID)
	if err != nil {
		return err
	}

	for _, a := range authenticators {
		if a.IsIndependent() {
			continue
		}

		orphaned := true
		for _, i := range identities {
			if a.IsDependentOf(i) {
				orphaned = false
			}
		}

		if orphaned {
			err = s.Delete(a)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	return s.Lockout.ClearAttempts(userID, usedMethods)
}
