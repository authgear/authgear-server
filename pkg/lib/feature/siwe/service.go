package siwe

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	siwego "github.com/spruceid/siwe-go"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/web3"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package siwe

// siwe-go library regex does not support underscore so we define a new one for this case
// https://github.com/spruceid/siwe-go/blob/fc1b0374f4ffff68e3455839655e680be7e0f862/regex.go#L17
const Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	SIWENoncePerIP  ratelimit.BucketName = "SIWENoncePerIP"
	SIWEVerifyPerIP ratelimit.BucketName = "SIWEVerifyPerIP"
)

type NonceStore interface {
	Create(ctx context.Context, nonce *Nonce) error
	Get(ctx context.Context, nonce *Nonce) (*Nonce, error)
	Delete(ctx context.Context, nonce *Nonce) error
}

type RateLimiter interface {
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("siwe")} }

type Service struct {
	RemoteIP             httputil.RemoteIP
	HTTPOrigin           httputil.HTTPOrigin
	Web3Config           *config.Web3Config
	AuthenticationConfig *config.AuthenticationConfig

	Clock       clock.Clock
	NonceStore  NonceStore
	RateLimiter RateLimiter
	Logger      Logger
}

func (s *Service) rateLimitGenerateNonce() ratelimit.BucketSpec {
	enabled := true
	return ratelimit.NewBucketSpec(&config.RateLimitConfig{
		Enabled: &enabled,
		Period:  config.DurationString(time.Minute.String()),
		Burst:   100,
	}, SIWENoncePerIP, string(s.RemoteIP))
}

func (s *Service) rateLimitVerifyMessage() ratelimit.BucketSpec {
	return ratelimit.NewBucketSpec(
		s.AuthenticationConfig.RateLimits.SIWE.PerIP, SIWEVerifyPerIP,
		string(s.RemoteIP),
	)
}

func (s *Service) CreateNewNonce(ctx context.Context) (*Nonce, error) {
	nonce := rand.StringWithAlphabet(16, Alphabet, rand.SecureRand)
	nonceModel := &Nonce{
		Nonce:    nonce,
		ExpireAt: s.Clock.NowUTC().Add(duration.Short),
	}

	failed, err := s.RateLimiter.Allow(ctx, s.rateLimitGenerateNonce())
	if err != nil {
		return nil, err
	}
	if err := failed.Error(); err != nil {
		return nil, err
	}

	err = s.NonceStore.Create(ctx, nonceModel)
	if err != nil {
		return nil, err
	}

	return nonceModel, nil
}

func (s *Service) VerifyMessage(ctx context.Context, msg string, signature string) (*model.SIWEWallet, *ecdsa.PublicKey, error) {
	message, err := siwego.ParseMessage(msg)
	if err != nil {
		return nil, nil, err
	}

	chainID := message.GetChainID()

	var expectedNetworkID *web3.ContractID
	mismatchNetwork := true
	for _, networkURL := range s.Web3Config.SIWE.Networks {
		expectedNetworkID, err = web3.ParseContractID(networkURL)
		if err != nil {
			return nil, nil, err
		}

		if expectedNetworkID.Network == strconv.Itoa(chainID) {
			mismatchNetwork = false
			break
		}

	}

	if mismatchNetwork {
		return nil, nil, InvalidNetwork.NewWithInfo("network does not match expected network", apierrors.Details{"expected_chain_id": fmt.Sprintf("_%s", expectedNetworkID.Network)})
	}

	reservation, failed, err := s.RateLimiter.Reserve(ctx, s.rateLimitVerifyMessage())
	if err != nil {
		return nil, nil, err
	}
	if err := failed.Error(); err != nil {
		return nil, nil, err
	}
	defer s.RateLimiter.Cancel(ctx, reservation)

	messageNonce := message.GetNonce()
	existingNonce, err := s.NonceStore.Get(ctx, &Nonce{
		Nonce: messageNonce,
	})
	if errors.Is(err, ErrNonceNotFound) {
		reservation.PreventCancel()
		return nil, nil, err
	} else if err != nil {
		return nil, nil, err
	}

	publicOrigin, err := url.Parse(string(s.HTTPOrigin))
	if err != nil {
		return nil, nil, err
	}

	now := s.Clock.NowUTC()
	pubKey, err := message.Verify(signature, &publicOrigin.Host, &existingNonce.Nonce, &now)
	if err != nil {
		reservation.PreventCancel()
		return nil, nil, err
	}

	eip55, err := web3.NewEIP55(message.GetAddress().Hex())
	if err != nil {
		return nil, nil, err
	}

	wallet := &model.SIWEWallet{
		Address: eip55,
		ChainID: message.GetChainID(),
	}

	return wallet, pubKey, nil
}
