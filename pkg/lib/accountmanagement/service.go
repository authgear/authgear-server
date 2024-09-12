package accountmanagement

import (
	"encoding/json"
	"fmt"

	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type StartAddingInput struct {
	UserID                                          string
	Alias                                           string
	RedirectURI                                     string
	IncludeStateAuthorizationURLAndBindStateToToken bool
}

type StartAddingOutput struct {
	Token            string `json:"token,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type FinishAddingInput struct {
	UserID string
	Token  string
	Query  string
}

type FinishAddingOutput struct {
	// It is intentionally empty.
}
type ResendOTPCodeInput struct {
	Channel      model.AuthenticatorOOBChannel
	LoginID      string
	isSwitchPage bool
}

type sendOTPCodeInput struct {
	Channel  model.AuthenticatorOOBChannel
	Target   string
	isResend bool
}

type StartCreateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Channel    model.AuthenticatorOOBChannel
}

type StartUpdateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Channel    model.AuthenticatorOOBChannel
}

type startIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Channel    model.AuthenticatorOOBChannel
	isUpdate   bool
}

type StartIdentityWithVerificationOutput struct {
	Token            string `json:"token,omitempty"`
	SkipVerification bool
}

type ResumeAddingIdentityWithVerificationInput struct {
	Token string
}

type ResumeAddingIdentityWithVerificationOutput struct {
	Token          string `json:"token,omitempty"`
	LoginID        string
	LoginIDKeyType model.LoginIDKeyType
	IdentityID     string
}

type verifyIdentityInput struct {
	UserID       string
	Token        string
	Channel      model.AuthenticatorOOBChannel
	Code         string
	IdentityInfo *identity.Info
}

type CreateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type CreateIdentityWithVerificationOutput struct {
	// It is intentionally empty.
}

type UpdateIdentityWithVerificationInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
	Code       string
	Token      string
	Channel    model.AuthenticatorOOBChannel
}

type UpdateIdentityWithVerificationOutput struct {
	// It is intentionally empty.
}

type ChangePasswordInput struct {
	OAuthSessionID string
	RedirectURI    string
	OldPassword    string
	NewPassword    string
}

type ChangePasswordOutput struct {
	RedirectURI string
}

type CreateAdditionalPasswordInput struct {
	NewAuthenticatorID string
	UserID             string
	Password           string
}

type AddPasskeyInput struct {
	CreationResponse *protocol.CredentialCreationResponse
}

type AddPasskeyOutput struct {
	// It is intentionally empty.
}

func NewCreateAdditionalPasswordInput(userID string, password string) CreateAdditionalPasswordInput {
	return CreateAdditionalPasswordInput{
		NewAuthenticatorID: uuid.New(),
		UserID:             userID,
		Password:           password,
	}
}

type RemovePasskeyInput struct {
	IdentityID string
}

type RemovePasskeyOutput struct {
	// It is intentionally empty.
}

type AddBiometricInput struct {
	JWTToken string
}

type AddBiometricOutput struct {
	// It is intentionally empty.
}

type RemoveBiometricInput struct {
	IdentityID string
}

type RemoveBiometricOuput struct {
	// It is intentionally empty.
}

type AddUsernameInput struct {
	LoginID    string
	LoginIDKey string
}

type AddUsernameOutput struct {
	// It is intentionally empty.
}

type UpdateUsernameInput struct {
	LoginID    string
	LoginIDKey string
	IdentityID string
}

type UpdateUsernameOutput struct {
	// It is intentionally empty.
}

type RemoveUsernameInput struct {
	IdentityID string
}

type RemoveUsernameOutput struct {
	// It is intentionally empty.
}

type RemoveEmailInput struct {
	IdentityID string
}

type RemoveEmailOutput struct {
	// It is intentionally empty.
}

type RemovePhoneNumberInput struct {
	IdentityID string
}

type RemovePhoneNumberOutput struct {
	// It is intentionally empty.
}

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
}

type UserService interface {
	Get(id string, role accesscontrol.Role) (*model.User, error)
	UpdateMFAEnrollment(userID string, t *time.Time) error
}

type Store interface {
	GenerateToken(options GenerateTokenOptions) (string, error)
	GetToken(tokenStr string) (*Token, error)
	ConsumeToken(tokenStr string) (*Token, error)
}

type OAuthProvider interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)
	GetAuthorizationURL(alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
	GetUserProfile(alias string, options oauthrelyingparty.GetUserProfileOptions) (oauthrelyingparty.UserProfile, error)
}

type BiometricIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*biometric.Request, error)
}

type IdentityService interface {
	Get(id string) (*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	CheckDuplicated(info *identity.Info) (dupe *identity.Info, err error)
	Create(info *identity.Info) error
	Update(oldInfo *identity.Info, newInfo *identity.Info) error
	Delete(is *identity.Info) error
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type AuthenticatorService interface {
	NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	Update(authenticatorInfo *authenticator.Info) error
	UpdatePassword(authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error)
	VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *facade.VerifyOptions) (verifyResult *service.VerifyResult, err error)
}

type AuthenticationInfoService interface {
	Save(entry *authenticationinfo.Entry) error
}

type PasskeyService interface {
	ConsumeAttestationResponse(attestationResponse []byte) (err error)
}

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ otp.MessageType) (*otp.PreparedMessage, error)
	Send(msg *otp.PreparedMessage, opts otp.SendOptions) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
}

type VerificationService interface {
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
}

type Service struct {
	Database                  *appdb.Handle
	Config                    *config.AppConfig
	Challenges                ChallengeProvider
	Users                     UserService
	Store                     Store
	OAuthProvider             OAuthProvider
	BiometricProvider         BiometricIdentityProvider
	Identities                IdentityService
	Events                    EventService
	Authenticators            AuthenticatorService
	AuthenticationInfoService AuthenticationInfoService
	PasskeyService            PasskeyService
	UIInfoResolver            SettingsDeleteAccountSuccessUIInfoResolver
	OTPSender                 OTPSender
	OTPCodeService            OTPCodeService
	Verification              VerificationService
}

func (s *Service) verifyIdentity(input *verifyIdentityInput) (err error) {
	defer func() {
		if err == nil {
			_, err = s.Store.ConsumeToken(input.Token)
		}
	}()

	token, err := s.Store.GetToken(input.Token)
	err = token.CheckUser(input.UserID)
	if err != nil {
		return err
	}

	var loginIDValue string
	var loginIDType model.LoginIDKeyType
	switch {
	case token.Email != "":
		loginIDValue = token.Email
		loginIDType = model.LoginIDKeyTypeEmail
	case token.PhoneNumber != "":
		loginIDValue = token.PhoneNumber
		loginIDType = model.LoginIDKeyTypePhone
	default:
		return ErrAccountManagementTokenInvalid
	}

	err = s.OTPCodeService.VerifyOTP(
		otp.KindVerification(s.Config, input.Channel),
		loginIDValue,
		input.Code,
		&otp.VerifyOptions{UserID: input.UserID},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return verification.ErrInvalidVerificationCode
	} else if err != nil {
		return err
	}

	var claimName model.ClaimName
	claimName, ok := model.GetLoginIDKeyTypeClaim(loginIDType)
	if !ok {
		panic(fmt.Errorf("accountmanagement: unexpected login ID key"))
	}

	verifiedClaim := s.Verification.NewVerifiedClaim(input.UserID, string(claimName), loginIDValue)

	// NodeDoVerifyIdentity GetEffects()
	err = s.Verification.MarkClaimVerified(verifiedClaim)
	if err != nil {
		return err
	}

	identityInfo := input.IdentityInfo
	var e event.Payload
	if payload, ok := nonblocking.NewIdentityVerifiedEventPayload(
		model.UserRef{
			Meta: model.Meta{
				ID: identityInfo.UserID,
			},
		},
		identityInfo.ToModel(),
		string(verifiedClaim.Name),
		false,
	); ok {
		e = payload
	}

	if e != nil {
		if err := s.Events.DispatchEventOnCommit(e); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) makeAndCheckCreateLoginIDIdentityInfo(loginKey string, loginID string, userID string) (*identity.Info, error) {
	// EdgeUseIdentityLoginID
	identitySpec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, err
	}

	// EdgeCreateIdentityEnd
	identityInfo, err := s.Identities.New(userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, err
	}
	// EdgeDoCreateIdentity
	createDisabled := identityInfo.CreateDisabled(s.Config.Identity)
	if createDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	// NodeDoCreateIdentity GetEffects() -> EffectOnCommit()
	if _, err := s.Identities.CheckDuplicated(identityInfo); err != nil {
		return nil, err
	}

	return identityInfo, nil
}

func (s *Service) createAndDispatchEvent(identityInfo *identity.Info) (err error) {
	// Create identity after verification
	err = s.Identities.Create(identityInfo)
	if err != nil {
		return err
	}

	// Dispatch event
	if err = s.dispatchEnableIdentityEvent(identityInfo); err != nil {
		return err
	}

	return nil
}

func (s *Service) updateAndDispatchEvent(oldInfo *identity.Info, newInfo *identity.Info) (err error) {
	// Update identity after verification
	if err := s.Identities.Update(oldInfo, newInfo); err != nil {
		return err
	}

	// Dispatch event
	if err = s.dispatchUpdatedIdentityEvent(oldInfo, newInfo); err != nil {
		return err
	}
	return nil
}

func (s *Service) removeIdentity(identityID string, userID string) (identityInfo *identity.Info, err error) {
	// EdgeRemoveIdentity
	identityInfo, err = s.Identities.Get(identityID)
	if err != nil {
		return nil, err
	}

	if identityInfo.UserID != userID {
		return nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	// EdgeDoRemoveIdentity
	deleteDisabled := identityInfo.DeleteDisabled(s.Config.Identity)
	if deleteDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}

	err = s.Identities.Delete(identityInfo)
	if err != nil {
		return nil, err
	}

	// NodeDoRemoveIdentity GetEffects() -> EffectOnCommit()
	err = s.dispatchDisabledIdentityEvent(identityInfo)
	if err != nil {
		return nil, err
	}

	return identityInfo, nil
}

func (s *Service) makeAndCheckUpdateLoginIDIdentityInfo(identityID string, userID string, loginKey string, loginID string) (oldInfo *identity.Info, newInfo *identity.Info, err error) {
	// EdgeUseIdentityLoginID
	identitySpec, err := s.makeLoginIDSpec(loginKey, loginID)
	if err != nil {
		return nil, nil, err
	}

	oldInfo, err = s.Identities.Get(identityID)
	if err != nil {
		return nil, nil, err
	}

	if oldInfo.UserID != userID {
		return nil, nil, ErrAccountManagementIdentityNotOwnedbyToUser
	}

	newInfo, err = s.Identities.UpdateWithSpec(oldInfo, identitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
	if err != nil {
		return nil, nil, err
	}

	// EdgeDoUpdateIdentity
	updateDisabled := oldInfo.UpdateDisabled(s.Config.Identity)
	if updateDisabled {
		return nil, nil, api.ErrIdentityModifyDisabled
	}

	// NodeDoUpdateIdentity GetEffects() -> EffectRun()
	if _, err := s.Identities.CheckDuplicated(newInfo); err != nil {
		if identity.IsErrDuplicatedIdentity(err) {
			s1 := oldInfo.ToSpec()
			s2 := newInfo.ToSpec()
			return nil, nil, identity.NewErrDuplicatedIdentity(&s2, &s1)
		}
		return nil, nil, err
	}

	return oldInfo, newInfo, nil

}

func (s *Service) dispatchEnableIdentityEvent(identityInfo *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityInfo.UserID,
		},
	}

	var e event.Payload
	switch identityInfo.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityInfo.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDAddedEventPayload(
			userRef,
			identityInfo.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthConnectedEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricEnabledEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchUpdatedIdentityEvent(identityAfterUpdate *identity.Info, identityBeforeUpdate *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityAfterUpdate.UserID,
		},
	}

	var e event.Payload
	switch identityAfterUpdate.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityAfterUpdate.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDUpdatedEventPayload(
			userRef,
			identityAfterUpdate.ToModel(),
			identityBeforeUpdate.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) dispatchDisabledIdentityEvent(identityInfo *identity.Info) (err error) {
	userRef := model.UserRef{
		Meta: model.Meta{
			ID: identityInfo.UserID,
		},
	}

	var e event.Payload
	switch identityInfo.Type {
	case model.IdentityTypeLoginID:
		loginIDType := identityInfo.LoginID.LoginIDType
		if payload, ok := nonblocking.NewIdentityLoginIDRemovedEventPayload(
			userRef,
			identityInfo.ToModel(),
			string(loginIDType),
			false,
		); ok {
			e = payload
		}
	case model.IdentityTypeOAuth:
		e = &nonblocking.IdentityOAuthDisconnectedEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	case model.IdentityTypeBiometric:
		e = &nonblocking.IdentityBiometricDisabledEventPayload{
			UserRef:  userRef,
			Identity: identityInfo.ToModel(),
			AdminAPI: false,
		}
	}

	if e != nil {
		err = s.Events.DispatchEventOnCommit(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) StartAdding(input *StartAddingInput) (*StartAddingOutput, error) {
	state := ""
	if input.IncludeStateAuthorizationURLAndBindStateToToken {
		state = GenerateRandomState()
	}

	param := oauthrelyingparty.GetAuthorizationURLOptions{
		RedirectURI: input.RedirectURI,
		State:       state,
	}

	authorizationURL, err := s.OAuthProvider.GetAuthorizationURL(input.Alias, param)
	if err != nil {
		return nil, err
	}

	token, err := s.Store.GenerateToken(GenerateTokenOptions{
		UserID:      input.UserID,
		Alias:       input.Alias,
		RedirectURI: input.RedirectURI,
		MaybeState:  state,
	})
	if err != nil {
		return nil, err
	}

	return &StartAddingOutput{
		Token:            token,
		AuthorizationURL: authorizationURL,
	}, nil
}

func (s *Service) FinishAdding(input *FinishAddingInput) (*FinishAddingOutput, error) {
	token, err := s.Store.ConsumeToken(input.Token)
	if err != nil {
		return nil, err
	}

	err = token.CheckOAuthUser(input.UserID)
	if err != nil {
		return nil, err
	}

	state, err := ExtractStateFromQuery(input.Query)
	if err != nil {
		return nil, err
	}

	err = token.CheckState(state)
	if err != nil {
		return nil, err
	}

	providerConfig, err := s.OAuthProvider.GetProviderConfig(token.Alias)
	if err != nil {
		return nil, err
	}

	emptyNonce := ""
	userProfile, err := s.OAuthProvider.GetUserProfile(token.Alias, oauthrelyingparty.GetUserProfileOptions{
		Query:       input.Query,
		RedirectURI: token.RedirectURI,
		Nonce:       emptyNonce,
	})
	if err != nil {
		return nil, err
	}

	providerID := providerConfig.ProviderID()
	spec := &identity.Spec{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuthSpec{
			ProviderID:     providerID,
			SubjectID:      userProfile.ProviderUserID,
			RawProfile:     userProfile.ProviderRawProfile,
			StandardClaims: userProfile.StandardAttributes,
		},
	}

	info, err := s.Identities.New(
		token.UserID,
		spec,
		// We are not adding Login ID here so the options is irrelevant.
		identity.NewIdentityOptions{},
	)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		_, err = s.Identities.CheckDuplicated(info)
		if err != nil {
			return err
		}

		err = s.Identities.Create(info)
		if err != nil {
			return err
		}

		evt := &nonblocking.IdentityOAuthConnectedEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: info.UserID,
				},
			},
			Identity: info.ToModel(),
			AdminAPI: false,
		}

		err = s.Events.DispatchEventOnCommit(evt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &FinishAddingOutput{}, nil
}

func (s *Service) sendOTPCode(input *sendOTPCodeInput) error {
	var msgType otp.MessageType
	switch input.Channel {
	case model.AuthenticatorOOBChannelWhatsapp:
		msgType = otp.MessageTypeWhatsappCode
	case model.AuthenticatorOOBChannelSMS:
		msgType = otp.MessageTypeVerification
	case model.AuthenticatorOOBChannelEmail:
		msgType = otp.MessageTypeVerification
	default:
		panic(fmt.Errorf("accountmanagement: unknown channel"))
	}

	msg, err := s.OTPSender.Prepare(input.Channel, input.Target, otp.FormCode, msgType)
	if !input.isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodeService.GenerateOTP(
		otp.KindVerification(s.Config, input.Channel),
		input.Target,
		otp.FormCode,
		&otp.GenerateOptions{},
	)
	// If it is not resend (switch between page), we should not send and return rate limit error to the caller.
	if !input.isResend && apierrors.IsKind(err, ratelimit.RateLimited) {
		return nil
	} else if err != nil {
		return err
	}

	err = s.OTPSender.Send(msg, otp.SendOptions{OTP: code})
	if err != nil {
		return err
	}

	return nil

}

func (s *Service) StartCreateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartCreateIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	return s.startIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Channel:    input.Channel,
		isUpdate:   false,
	})
}

func (s *Service) StartUpdateIdentityWithVerification(resolvedSession session.ResolvedSession, input *StartUpdateIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	return s.startIdentityWithVerification(resolvedSession, &startIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Channel:    input.Channel,
		IdentityID: input.IdentityID,
		isUpdate:   true,
	})
}

func (s *Service) startIdentityWithVerification(resolvedSession session.ResolvedSession, input *startIdentityWithVerificationInput) (output *StartIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	var token string

	var newInfo *identity.Info
	var oldInfo *identity.Info

	// Do Identity Checking for before sending OTP
	switch {
	case input.isUpdate:
		oldInfo, newInfo, err = s.makeAndCheckUpdateLoginIDIdentityInfo(input.IdentityID, userID, input.LoginIDKey, input.LoginID)
	case !input.isUpdate:
		newInfo, err = s.makeAndCheckCreateLoginIDIdentityInfo(input.LoginIDKey, input.LoginID, userID)
	}
	if err != nil {
		return nil, err
	}

	claims, err := s.Verification.GetIdentityVerificationStatus(newInfo)
	if err != nil {
		return nil, err
	}

	// If the identity is already verified, skip verification and create/update immediately.
	if len(claims) > 0 && claims[0].Verified {
		switch {
		case input.isUpdate:
			err = s.updateAndDispatchEvent(oldInfo, newInfo)
		case !input.isUpdate:
			err = s.createAndDispatchEvent(newInfo)
		}
		if err != nil {
			return nil, err
		}
		return &StartIdentityWithVerificationOutput{SkipVerification: true}, nil
	}

	loginIDType := newInfo.LoginID.LoginIDType
	switch loginIDType {
	case model.LoginIDKeyTypeEmail:
		token, err = s.Store.GenerateToken(GenerateTokenOptions{
			UserID:     userID,
			Email:      input.LoginID,
			IdentityID: input.IdentityID,
		})
	case model.LoginIDKeyTypePhone:
		token, err = s.Store.GenerateToken(GenerateTokenOptions{
			UserID:      userID,
			PhoneNumber: input.LoginID,
			IdentityID:  input.IdentityID,
		})
	}
	if err != nil {
		return nil, err
	}

	err = s.sendOTPCode(&sendOTPCodeInput{
		Channel:  input.Channel,
		Target:   input.LoginID,
		isResend: false,
	})
	if err != nil {
		return nil, err
	}

	return &StartIdentityWithVerificationOutput{
		Token:            token,
		SkipVerification: false,
	}, nil
}

func (s *Service) ResumeAddingIdentityWithVerification(resolvedSession session.ResolvedSession, input *ResumeAddingIdentityWithVerificationInput) (output *ResumeAddingIdentityWithVerificationOutput, err error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	token, err := s.Store.GetToken(input.Token)
	if err != nil {
		return nil, err
	}
	err = token.CheckUser(userID)
	if err != nil {
		return nil, err
	}

	var loginID string
	var loginIDKeyType model.LoginIDKeyType
	identityID := token.IdentityID

	switch {
	case token.Email != "":
		loginID = token.Email
		loginIDKeyType = model.LoginIDKeyTypeEmail
	case token.PhoneNumber != "":
		loginID = token.PhoneNumber
		loginIDKeyType = model.LoginIDKeyTypePhone
	default:
		return nil, ErrAccountManagementTokenInvalid
	}

	return &ResumeAddingIdentityWithVerificationOutput{
		Token:          input.Token,
		LoginID:        loginID,
		LoginIDKeyType: loginIDKeyType,
		IdentityID:     identityID,
	}, nil
}

func (s *Service) ResendOTPCode(input *ResendOTPCodeInput) (err error) {
	// Either it is a switch page or resend
	isResend := !input.isSwitchPage
	err = s.sendOTPCode(&sendOTPCodeInput{
		Channel:  input.Channel,
		Target:   input.LoginID,
		isResend: isResend,
	})
	if err != nil {
		return err
	}

	return nil
}

// If have OAuthSessionID, it means the user is changing password after login with SDK.
// Then do special handling such as authenticationInfo
func (s *Service) ChangePassword(resolvedSession session.ResolvedSession, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID

	ais, err := s.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindPrimary),
	)
	if err != nil {
		return &ChangePasswordOutput{}, err
	}
	if len(ais) == 0 {
		return &ChangePasswordOutput{}, api.ErrNoPassword
	}
	oldInfo := ais[0]
	_, err = s.Authenticators.VerifyWithSpec(oldInfo, &authenticator.Spec{
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.OldPassword,
		},
	}, nil)
	if err != nil {
		err = api.ErrInvalidCredentials
		return &ChangePasswordOutput{}, err
	}
	changed, newInfo, err := s.Authenticators.UpdatePassword(oldInfo, &service.UpdatePasswordOptions{
		SetPassword:    true,
		PlainPassword:  input.NewPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return &ChangePasswordOutput{}, err
	}
	if changed {
		err = s.Database.WithTx(func() error {
			err = s.Authenticators.Update(newInfo)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return &ChangePasswordOutput{}, err
		}

	}
	redirectURI := input.RedirectURI
	// If is changing password with SDK.
	if input.OAuthSessionID != "" {
		authInfo := resolvedSession.GetAuthenticationInfo()
		authenticationInfoEntry := authenticationinfo.NewEntry(authInfo, input.OAuthSessionID, "")

		err = s.Database.WithTx(func() error {
			err = s.AuthenticationInfoService.Save(authenticationInfoEntry)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return &ChangePasswordOutput{}, err
		}
		redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(input.RedirectURI, authenticationInfoEntry)
	}
	return &ChangePasswordOutput{RedirectURI: redirectURI}, nil

}

func (s *Service) CreateAdditionalPassword(input CreateAdditionalPasswordInput) error {
	spec := &authenticator.Spec{
		UserID:    input.UserID,
		IsDefault: false,
		Kind:      model.AuthenticatorKindSecondary,
		Type:      model.AuthenticatorTypePassword,
		Password: &authenticator.PasswordSpec{
			PlainPassword: input.Password,
		},
	}
	info, err := s.Authenticators.NewWithAuthenticatorID(input.NewAuthenticatorID, spec)
	if err != nil {
		return err
	}
	return s.CreateAuthenticator(info)
}

func (s *Service) CreateAuthenticator(authenticatorInfo *authenticator.Info) error {
	err := s.Database.WithTx(func() error {
		err := s.Authenticators.Create(authenticatorInfo, false)
		if err != nil {
			return err
		}
		if authenticatorInfo.Kind == authenticator.KindSecondary {
			err = s.Users.UpdateMFAEnrollment(authenticatorInfo.UserID, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) AddPasskey(resolvedSession session.ResolvedSession, input *AddPasskeyInput) (*AddPasskeyOutput, error) {
	// NodePromptCreatePasskey ReactTo
	// case inputNodePromptCreatePasskey.IsCreationResponse()
	userID := resolvedSession.GetAuthenticationInfo().UserID
	creationResponse := input.CreationResponse
	creationResponseBytes, err := json.Marshal(creationResponse)
	if err != nil {
		return nil, err
	}

	authenticatorSpec := &authenticator.Spec{
		UserID: userID,
		Kind:   authenticator.KindPrimary,
		Type:   model.AuthenticatorTypePasskey,
		Passkey: &authenticator.PasskeySpec{
			AttestationResponse: creationResponseBytes,
		},
	}

	authenticatorID := uuid.New()
	authenticatorInfo, err := s.Authenticators.NewWithAuthenticatorID(authenticatorID, authenticatorSpec)
	if err != nil {
		return nil, err
	}

	identitySpec := &identity.Spec{
		Type: model.IdentityTypePasskey,
		Passkey: &identity.PasskeySpec{
			AttestationResponse: creationResponseBytes,
		},
	}
	identityInfo, err := s.Identities.New(userID, identitySpec, identity.NewIdentityOptions{})
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		err := s.Identities.Create(identityInfo)
		if err != nil {
			return err
		}
		// NodeDoCreatePasskey GetEffects()
		err = s.Authenticators.Create(authenticatorInfo, false)
		if err != nil {
			return err
		}
		err = s.PasskeyService.ConsumeAttestationResponse(creationResponseBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddPasskeyOutput{}, nil
}

func (s *Service) RemovePasskey(resolvedSession session.ResolvedSession, input *RemovePasskeyInput) (*RemovePasskeyOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	err := s.Database.WithTx(func() (err error) {
		// case *nodes.NodeDoUseUser: (Passkey skip DeleteDisabled check)
		_, err = s.removeIdentity(identityID, userID)
		if err != nil {
			return err
		}
		// NodeDoRemoveIdentity GetEffects() -> EffectOnCommit()
		// Passkey no PayloadEvent for EffectOnCommit

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemovePasskeyOutput{}, nil
}

func (s *Service) AddBiometric(resolvedSession session.ResolvedSession, input *AddBiometricInput) (*AddBiometricOutput, error) {

	// EdgeUseIdentityBiometric
	enabled := false
	for _, t := range s.Config.Authentication.Identities {
		if t == model.IdentityTypeBiometric {
			enabled = true
			break
		}
	}

	if !enabled {
		return nil, api.NewInvariantViolated(
			"BiometricDisallowed",
			"biometric is not allowed",
			nil,
		)
	}

	jwt := input.JWTToken

	request, err := s.BiometricProvider.ParseRequestUnverified(jwt)
	if err != nil {
		return nil, api.ErrInvalidCredentials
	}

	purpose, err := s.Challenges.Consume(request.Challenge)
	if err != nil || *purpose != challenge.PurposeBiometricRequest {
		return nil, api.ErrInvalidCredentials
	}

	// request.Action case: identitybiometric.RequestActionSetup
	displayName := deviceinfo.DeviceModel(request.DeviceInfo)
	if displayName == "" {
		return nil, api.ErrInvalidCredentials
	}
	if request.Key == nil {
		return nil, api.ErrInvalidCredentials
	}

	key, err := json.Marshal(request.Key)
	if err != nil {
		return nil, err
	}

	userID := resolvedSession.GetAuthenticationInfo().UserID

	// IsCreating: true
	identitySpec := &identity.Spec{
		Type: model.IdentityTypeBiometric,
		Biometric: &identity.BiometricSpec{
			KeyID:      request.KeyID,
			Key:        string(key),
			DeviceInfo: request.DeviceInfo,
		},
	}

	// EdgeCreateIdentityEnd
	identityInfo, err := s.Identities.New(userID, identitySpec, identity.NewIdentityOptions{LoginIDEmailByPassBlocklistAllowlist: false})
	if err != nil {
		return nil, err
	}

	// EdgeDoCreateIdentity
	err = s.Database.WithTx(func() error {
		user, err := s.Users.Get(userID, accesscontrol.RoleGreatest)
		if err != nil {
			return err
		}

		if user.IsAnonymous {
			return api.NewInvariantViolated(
				"AnonymousUserAddIdentity",
				"anonymous user cannot add identity",
				nil,
			)
		}

		if _, err := s.Identities.CheckDuplicated(identityInfo); err != nil {
			return err
		}

		if err := s.Identities.Create(identityInfo); err != nil {
			return err
		}

		// EdgeEnsureVerificationBegin (shouldVerify: false)
		// -> EdgeDoVerifyIdentity (NewVerifiedClaim = nil, SkipVerificationEvent = true)
		// -> EdgeDoUseIdentity (e.UserIDHint == "")
		// -> EdgeEnsureRemoveAnonymousIdentity (anonymousIdentities == nil as checked above)
		// -> EdgeCreateAuthenticatorBegin (n.Stage: authn.AuthenticationStagePrimary -> identityRequiresPrimaryAuthentication = false)
		// -> EdgeCreateAuthenticatorEnd (Authenticators: nil so do nothing)
		// -> NodeDoCreateAuthenticator (node.Stage: authn.AuthenticationStagePrimary -> authenticators == nil so do nothing)

		if err := s.dispatchEnableIdentityEvent(identityInfo); err != nil {
			return err
		}
		// NodeDoVerifyIdentity (EffectOnCommit return nil as n.SkipVerificationEvent == true)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AddBiometricOutput{}, nil
}

func (s *Service) RemoveBiometric(resolvedSession session.ResolvedSession, input *RemoveBiometricInput) (*RemoveBiometricOuput, error) {
	identityID := input.IdentityID
	userID := resolvedSession.GetAuthenticationInfo().UserID

	err := s.Database.WithTx(func() (err error) {
		// case *nodes.NodeDoUseUser: (Biometric skip DeleteDisabled check)
		_, err = s.removeIdentity(identityID, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveBiometricOuput{}, nil
}

func (s *Service) makeLoginIDSpec(loginIDKey string, loginID string) (*identity.Spec, error) {
	// EdgeUseIdentityLoginID
	matchedLoginIDConfig, ok := s.Config.Identity.LoginID.GetKeyConfig(loginIDKey)
	if !ok {
		return nil, api.NewInvariantViolated(
			"InvalidLoginIDKey",
			"invalid login ID key",
			nil,
		)
	}
	typ := matchedLoginIDConfig.Type
	identitySpec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Key:   loginIDKey,
			Type:  typ,
			Value: loginID,
		},
	}
	return identitySpec, nil
}

func (s *Service) AddUsername(resolvedSession session.ResolvedSession, input *AddUsernameInput) (*AddUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID

	identityInfo, err := s.makeAndCheckCreateLoginIDIdentityInfo(loginKey, loginID, userID)
	if err != nil {
		return nil, err
	}

	err = s.Database.WithTx(func() error {
		err := s.Identities.Create(identityInfo)
		if err != nil {
			return err
		}

		// No Need to Verify shouldVerify: false
		// Skip EdgeEnsureVerificationBegin, EdgeDoVerifyIdentity, EdgeCreateAuthenticatorBegin, EdgeDoCreateAuthenticator

		// NodeDoCreateIdentity GetEffects() -> EffectOnCommit()
		if err := s.dispatchEnableIdentityEvent(identityInfo); err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		return nil, err
	}

	return &AddUsernameOutput{}, nil

}

func (s *Service) UpdateUsername(resolvedSession session.ResolvedSession, input *UpdateUsernameInput) (*UpdateUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginKey := input.LoginIDKey
	loginID := input.LoginID
	identityID := input.IdentityID

	err := s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.makeAndCheckUpdateLoginIDIdentityInfo(identityID, userID, loginKey, loginID)
		if err != nil {
			return err
		}
		if err := s.updateAndDispatchEvent(oldInfo, newInfo); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateUsernameOutput{}, nil
}

func (s *Service) RemoveUsername(resolvedSession session.ResolvedSession, input *RemoveUsernameInput) (*RemoveUsernameOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	err := s.Database.WithTx(func() (err error) {
		// case *nodes.NodeDoUseUser: (Username skip DeleteDisabled check)
		_, err = s.removeIdentity(identityID, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveUsernameOutput{}, nil
}

func (s *Service) createIdentityWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginID := input.LoginID
	loginIDKey := input.LoginIDKey

	err := s.Database.WithTx(func() error {
		identityInfo, err := s.makeAndCheckCreateLoginIDIdentityInfo(loginIDKey, loginID, userID)
		if err != nil {
			return err
		}

		err = s.verifyIdentity(&verifyIdentityInput{
			UserID:       userID,
			Token:        input.Token,
			Channel:      input.Channel,
			Code:         input.Code,
			IdentityInfo: identityInfo,
		})
		if err != nil {
			return err
		}

		// Create identity after verification
		err = s.Identities.Create(identityInfo)
		if err != nil {
			return err
		}

		// Dispatch event
		if err = s.dispatchEnableIdentityEvent(identityInfo); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreateIdentityWithVerificationOutput{}, nil
}

func (s *Service) updateIdentityWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	loginID := input.LoginID
	loginIDKey := input.LoginIDKey
	identityID := input.IdentityID

	err := s.Database.WithTx(func() error {
		oldInfo, newInfo, err := s.makeAndCheckUpdateLoginIDIdentityInfo(identityID, userID, loginIDKey, loginID)
		if err != nil {
			return err
		}

		err = s.verifyIdentity(&verifyIdentityInput{
			UserID:       userID,
			Token:        input.Token,
			Channel:      input.Channel,
			Code:         input.Code,
			IdentityInfo: newInfo,
		})
		if err != nil {
			return err
		}

		// Update identity after verification
		if err = s.updateAndDispatchEvent(oldInfo, newInfo); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateIdentityWithVerificationOutput{}, nil
}

func (s *Service) AddEmailWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	return s.createIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) UpdateEmailWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	return s.updateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		IdentityID: input.IdentityID,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) RemoveEmail(resolvedSession session.ResolvedSession, input *RemoveEmailInput) (*RemoveEmailOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	err := s.Database.WithTx(func() (err error) {
		_, err = s.removeIdentity(identityID, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemoveEmailOutput{}, nil
}

func (s *Service) AddPhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *CreateIdentityWithVerificationInput) (*CreateIdentityWithVerificationOutput, error) {
	return s.createIdentityWithVerification(resolvedSession, &CreateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) UpdatePhoneNumberWithVerification(resolvedSession session.ResolvedSession, input *UpdateIdentityWithVerificationInput) (*UpdateIdentityWithVerificationOutput, error) {
	return s.updateIdentityWithVerification(resolvedSession, &UpdateIdentityWithVerificationInput{
		LoginID:    input.LoginID,
		LoginIDKey: input.LoginIDKey,
		IdentityID: input.IdentityID,
		Code:       input.Code,
		Channel:    input.Channel,
		Token:      input.Token,
	})
}

func (s *Service) RemovePhoneNumber(resolvedSession session.ResolvedSession, input *RemovePhoneNumberInput) (*RemovePhoneNumberOutput, error) {
	userID := resolvedSession.GetAuthenticationInfo().UserID
	identityID := input.IdentityID

	err := s.Database.WithTx(func() (err error) {
		_, err = s.removeIdentity(identityID, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RemovePhoneNumberOutput{}, nil
}
