package siwe

import (
	"crypto/ecdsa"
	"fmt"
	"net/url"
	"strconv"

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
	siwego "github.com/spruceid/siwe-go"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package siwe

// siwe-go library regex does not support underscore so we define a new one for this case
// https://github.com/spruceid/siwe-go/blob/fc1b0374f4ffff68e3455839655e680be7e0f862/regex.go#L17
const Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	Web3Config *config.Web3Config

	Clock       clock.Clock
	NonceStore  NonceStore
	RateLimiter RateLimiter
	Logger      Logger
}

func (s *Service) CreateNewNonce() (*Nonce, error) {
	nonce := rand.StringWithAlphabet(16, Alphabet, rand.SecureRand)
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

func (s *Service) VerifyMessage(msg string, signature string) (*model.SIWEWallet, *ecdsa.PublicKey, error) {
	message, err := siwego.ParseMessage(msg)
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

	publicOrigin, err := url.Parse(s.HTTPConfig.PublicOrigin)
	if err != nil {
		return nil, nil, err
	}
	now := s.Clock.NowUTC()

	pubKey, err := message.Verify(signature, &publicOrigin.Host, &existingNonce.Nonce, &now)
	if err != nil {
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
