package verification

import (
	"errors"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
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
	Get(id string) (*Code, error)
	Delete(id string) error
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
	Request    *http.Request
	Logger     Logger
	Config     *config.VerificationConfig
	TrustProxy config.TrustProxy

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
		if c == nil || !*c.Enabled {
			continue
		}

		value, ok := claimValue.(string)
		if !ok {
			continue
		}

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
				Name:   claimName,
				Status: status,
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
	if a.Type != authn.AuthenticatorTypeOOBEmail && a.Type != authn.AuthenticatorTypeOOBSMS {
		panic("verification: incompatible authenticator type: " + a.Type)
	}

	var claimName string
	var claimValue string
	aClaims := a.StandardClaims()
	switch a.Type {
	case authn.AuthenticatorTypeOOBEmail:
		claimName = string(authn.ClaimEmail)
		claimValue = aClaims[authn.ClaimEmail]
	case authn.AuthenticatorTypeOOBSMS:
		claimName = string(authn.ClaimPhoneNumber)
		claimValue = aClaims[authn.ClaimPhoneNumber]
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

func (s *Service) CreateNewCode(id string, info *identity.Info, webSessionID string, requestedByUser bool) (*Code, error) {
	if info.Type != authn.IdentityTypeLoginID {
		panic("verification: expect login ID identity")
	}

	loginIDType := config.LoginIDKeyType(info.Claims[identity.IdentityClaimLoginIDType].(string))

	code := secretcode.OOBOTPSecretCode.Generate()
	codeModel := &Code{
		ID:              id,
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

func (s *Service) GetCode(id string) (*Code, error) {
	return s.CodeStore.Get(id)
}

func (s *Service) VerifyCode(id string, code string) (*Code, error) {
	err := s.RateLimiter.TakeToken(VerifyRateLimitBucket(httputil.GetIP(s.Request, bool(s.TrustProxy))))
	if err != nil {
		return nil, err
	}

	codeModel, err := s.CodeStore.Get(id)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	if !secretcode.OOBOTPSecretCode.Compare(code, codeModel.Code) {
		return nil, ErrInvalidVerificationCode
	}

	if err = s.CodeStore.Delete(id); err != nil {
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

// DeriveStandardAttributes populates email_verified and phone_number_verified,
// if email or phone_number are found in attrs.
func (s *Service) DeriveStandardAttributes(userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	for key, value := range attrs {
		// Copy
		out[key] = value

		// Email
		if key == stdattrs.Email {
			verified := false
			if str, ok := value.(string); ok {
				claims, err := s.ClaimStore.ListByClaimName(userID, stdattrs.Email)
				if err != nil {
					return nil, err
				}
				for _, claim := range claims {
					if claim.Value == str {
						verified = true
					}
				}
			}
			out[stdattrs.EmailVerified] = verified
		}

		// Phone number
		if key == stdattrs.PhoneNumber {
			verified := false
			if str, ok := value.(string); ok {
				claims, err := s.ClaimStore.ListByClaimName(userID, stdattrs.PhoneNumber)
				if err != nil {
					return nil, err
				}
				for _, claim := range claims {
					if claim.Value == str {
						verified = true
					}
				}
			}
			out[stdattrs.PhoneNumberVerified] = verified
		}
	}

	// updated_at
	out["updated_at"] = updatedAt.Unix()

	return out, nil
}
