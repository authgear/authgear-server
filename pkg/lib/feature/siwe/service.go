package siwe

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/secrets"
	"github.com/lestrrat-go/jwx/jwk"
	siwego "github.com/spruceid/siwe-go"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package siwe

type NonceStore interface {
	Create(nonce *Nonce) error
	Get(nonce *Nonce) (*Nonce, error)
	Delete(nonce *Nonce) error
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("siwe")} }

type Service struct {
	RemoteIP   httputil.RemoteIP
	HTTPConfig *config.HTTPConfig

	Clock       clock.Clock
	NonceStore  NonceStore
	RateLimiter RateLimiter
	Logger      Logger
}

func (s *Service) CreateNewNonce() (*Nonce, error) {
	nonce := secrets.GenerateSecret(16, rand.SecureRand)
	nonceModel := &Nonce{
		Nonce:    nonce,
		ExpireAt: s.Clock.NowUTC().Add(duration.Short),
	}

	err := s.RateLimiter.TakeToken(AntiSpamNonceBucket(string(s.RemoteIP)))
	if err != nil {
		return nil, err
	}

	err = s.NonceStore.Create(nonceModel)
	if err != nil {
		return nil, err
	}

	return nonceModel, nil
}

func (s *Service) VerifyMessage(request model.SIWEVerificationRequest) (*siwego.Message, jwk.Key, error) {
	message, err := siwego.ParseMessage(request.Message)
	if err != nil {
		return nil, nil, err
	}

	messageNonce := message.GetNonce()
	existingNonce, err := s.NonceStore.Get(&Nonce{
		Nonce: messageNonce,
	})
	if err != nil {
		return nil, nil, err
	}

	domain := s.HTTPConfig.PublicOrigin
	now := s.Clock.NowUTC()

	pubKey, err := message.Verify(request.Signature, &domain, &existingNonce.Nonce, &now)
	if err != nil {
		return nil, nil, err
	}

	jwkKey, err := jwk.New(pubKey)
	if err != nil {
		return nil, nil, err
	}

	if err = s.NonceStore.Delete(existingNonce); err != nil {
		s.Logger.WithError(err).Error("failed to delete nonce after verification")
	}

	return message, jwkKey, nil
}
