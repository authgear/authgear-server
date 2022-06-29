package verification

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type CodeStore interface {
	Create(code *Code) error
	Get(codeKey *CodeKey) (*Code, error)
	Delete(codeKey *CodeKey) error
}

type ClaimStore interface {
	ListByUser(userID string) ([]*Claim, error)
	ListByClaimName(userID string, claimName string) ([]*Claim, error)
	Get(userID string, claimName string, claimValue string) (*Claim, error)
	Create(claim *Claim) error
	Delete(id string) error
	DeleteAll(userID string) error
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("verification")} }

type Service struct {
	RemoteIP          httputil.RemoteIP
	Logger            Logger
	Config            *config.VerificationConfig
	UserProfileConfig *config.UserProfileConfig

	Clock       clock.Clock
	CodeStore   CodeStore
	ClaimStore  ClaimStore
	RateLimiter RateLimiter
}

func (s *Service) claimVerificationConfig(claimName string) *config.VerificationClaimConfig {
	switch claimName {
	case identity.StandardClaimEmail:
		return s.Config.Claims.Email
	case identity.StandardClaimPhoneNumber:
		return s.Config.Claims.PhoneNumber
	default:
		return nil
	}
}

func (s *Service) IsClaimVerifiable(claimName string) bool {
	if c := s.claimVerificationConfig(claimName); c != nil && *c.Enabled {
		return true
	}
	return false
}

func (s *Service) GetClaimVerificationStatus(userID string, name string, value string) (Status, error) {
	c := s.claimVerificationConfig(name)
	if c == nil || !*c.Enabled {
		return StatusDisabled, nil
	}

	_, err := s.ClaimStore.Get(userID, name, value)
	if errors.Is(err, ErrClaimUnverified) {
		if *c.Required {
			return StatusRequired, nil
		}
		return StatusPending, nil
	} else if err != nil {
		return "", err
	}

	return StatusVerified, nil
}

func (s *Service) getVerificationStatus(i *identity.Info, verifiedClaims map[claim]struct{}) []ClaimStatus {
	var statuses []ClaimStatus
	for claimName, claimValue := range i.Claims {
		c := s.claimVerificationConfig(claimName)
		if c == nil {
			continue
		}

		value, ok := claimValue.(string)
		if !ok {
			continue
		}

		isEnabled := *c.Enabled

		var status Status
		if _, verified := verifiedClaims[claim{claimName, value}]; verified {
			status = StatusVerified
		} else if *c.Required {
			status = StatusRequired
		} else {
			status = StatusPending
		}

		if status != StatusDisabled {
			statuses = append(statuses, ClaimStatus{
				Name:      claimName,
				Status:    status,
				IsEnabled: isEnabled,
			})
		}
	}
	return statuses
}

func (s *Service) GetIdentityVerificationStatus(i *identity.Info) ([]ClaimStatus, error) {
	claims, err := s.ClaimStore.ListByUser(i.UserID)
	if err != nil {
		return nil, err
	}

	verifiedClaims := make(map[claim]struct{})
	for _, c := range claims {
		verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
	}

	return s.getVerificationStatus(i, verifiedClaims), nil
}

func (s *Service) GetVerificationStatuses(is []*identity.Info) (map[string][]ClaimStatus, error) {
	if len(is) == 0 {
		return nil, nil
	}

	// Assuming user ID of all identities is same
	userID := is[0].UserID
	claims, err := s.ClaimStore.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	verifiedClaims := make(map[claim]struct{})
	for _, c := range claims {
		verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
	}

	statuses := map[string][]ClaimStatus{}
	for _, i := range is {
		if i.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		statuses[i.ID] = s.getVerificationStatus(i, verifiedClaims)
	}
	return statuses, nil
}

func (s *Service) GetAuthenticatorVerificationStatus(a *authenticator.Info) (AuthenticatorStatus, error) {
	if a.Type != model.AuthenticatorTypeOOBEmail && a.Type != model.AuthenticatorTypeOOBSMS {
		panic("verification: incompatible authenticator type: " + a.Type)
	}

	var claimName string
	var claimValue string
	aClaims := a.StandardClaims()
	switch a.Type {
	case model.AuthenticatorTypeOOBEmail:
		claimName = string(model.ClaimEmail)
		claimValue = aClaims[model.ClaimEmail]
	case model.AuthenticatorTypeOOBSMS:
		claimName = string(model.ClaimPhoneNumber)
		claimValue = aClaims[model.ClaimPhoneNumber]
	}

	_, err := s.ClaimStore.Get(a.UserID, claimName, claimValue)
	if errors.Is(err, ErrClaimUnverified) {
		return AuthenticatorStatusUnverified, nil
	} else if err != nil {
		return "", err
	}

	return AuthenticatorStatusVerified, nil
}

func (s *Service) GetClaims(userID string) ([]*Claim, error) {
	return s.ClaimStore.ListByUser(userID)
}

func (s *Service) IsUserVerified(identities []*identity.Info) (bool, error) {
	statuses, err := s.GetVerificationStatuses(identities)
	if err != nil {
		return false, err
	}

	numVerifiable := 0
	numVerified := 0
	for _, claimStatuses := range statuses {
		for _, claim := range claimStatuses {
			switch claim.Status {
			case StatusVerified:
				numVerifiable++
				numVerified++
			case StatusPending, StatusRequired:
				numVerifiable++
			case StatusDisabled:
				break
			default:
				panic("verification: unknown status:" + claim.Status)
			}
		}
	}

	switch s.Config.Criteria {
	case config.VerificationCriteriaAny:
		return numVerifiable > 0 && numVerified >= 1, nil
	case config.VerificationCriteriaAll:
		return numVerifiable > 0 && numVerified == numVerifiable, nil
	default:
		panic("verification: unknown criteria " + s.Config.Criteria)
	}
}

func (s *Service) CreateNewCode(info *identity.Info, webSessionID string, requestedByUser bool) (*Code, error) {
	if info.Type != model.IdentityTypeLoginID {
		panic("verification: expect login ID identity")
	}

	loginIDType := config.LoginIDKeyType(info.Claims[identity.IdentityClaimLoginIDType].(string))

	code := secretcode.OOBOTPSecretCode.Generate()
	codeModel := &Code{
		UserID:          info.UserID,
		IdentityID:      info.ID,
		IdentityType:    string(info.Type),
		LoginIDType:     string(loginIDType),
		LoginID:         info.Claims[identity.IdentityClaimLoginIDValue].(string),
		Code:            code,
		ExpireAt:        s.Clock.NowUTC().Add(s.Config.CodeExpiry.Duration()),
		WebSessionID:    webSessionID,
		RequestedByUser: requestedByUser,
	}

	err := s.CodeStore.Create(codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) GetCode(webSessionID string, info *identity.Info) (*Code, error) {
	loginIDType := info.Claims[identity.IdentityClaimLoginIDType].(string)
	loginID := info.Claims[identity.IdentityClaimLoginIDValue].(string)
	return s.CodeStore.Get(&CodeKey{
		WebSessionID: webSessionID,
		LoginIDType:  loginIDType,
		LoginID:      loginID,
	})
}

func (s *Service) VerifyCode(webSessionID string, info *identity.Info, code string) (*Code, error) {
	loginIDType := info.Claims[identity.IdentityClaimLoginIDType].(string)
	loginID := info.Claims[identity.IdentityClaimLoginIDValue].(string)
	codeKey := &CodeKey{
		WebSessionID: webSessionID,
		LoginIDType:  loginIDType,
		LoginID:      loginID,
	}

	err := s.RateLimiter.TakeToken(AutiBruteForceVerifyBucket(string(s.RemoteIP)))
	if err != nil {
		return nil, err
	}

	codeModel, err := s.CodeStore.Get(codeKey)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	if !secretcode.OOBOTPSecretCode.Compare(code, codeModel.Code) {
		return nil, ErrInvalidVerificationCode
	}

	if err = s.CodeStore.Delete(codeKey); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after verification")
	}

	return codeModel, nil
}

func (s *Service) NewVerifiedClaim(userID string, claimName string, claimValue string) *Claim {
	return &Claim{
		ID:     uuid.New(),
		UserID: userID,
		Name:   claimName,
		Value:  claimValue,
	}
}

func (s *Service) MarkClaimVerified(claim *Claim) error {
	claims, err := s.GetClaims(claim.UserID)
	if err != nil {
		return err
	}
	for _, c := range claims {
		if c.Name == claim.Name && c.Value == claim.Value {
			return nil
		}
	}
	claim.CreatedAt = s.Clock.NowUTC()
	return s.ClaimStore.Create(claim)
}

func (s *Service) DeleteClaim(claimID string) error {
	return s.ClaimStore.Delete(claimID)
}

func (s *Service) ResetVerificationStatus(userID string) error {
	return s.ClaimStore.DeleteAll(userID)
}

func (s *Service) RemoveOrphanedClaims(identities []*identity.Info, authenticators []*authenticator.Info) error {
	// Assuming user ID of all identities is same
	userID := identities[0].UserID
	claims, err := s.ClaimStore.ListByUser(userID)
	if err != nil {
		return err
	}

	orphans := make(map[claim]*Claim)
	for _, c := range claims {
		orphans[claim{c.Name, c.Value}] = c
	}

	for _, i := range identities {
		if i.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		for name, value := range i.StandardClaims() {
			delete(orphans, claim{Name: string(name), Value: value})
		}
	}

	for _, a := range authenticators {
		if a.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		for name, value := range a.StandardClaims() {
			delete(orphans, claim{Name: string(name), Value: value})
		}
	}

	for _, claim := range orphans {
		err = s.ClaimStore.Delete(claim.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
