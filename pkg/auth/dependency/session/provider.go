package session

import (
	"crypto/subtle"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const (
	tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tokenLength   = 32
)

type AccessEventProvider interface {
	InitStream(s auth.AuthSession) error
}

type Rand *rand.Rand

type Provider struct {
	Request      *http.Request
	Store        Store
	AccessEvents AccessEventProvider
	ServerConfig *config.ServerConfig
	Config       *config.SessionConfig
	Clock        clock.Clock
	Random       Rand
}

func (p *Provider) MakeSession(attrs *authn.Attrs) (*IDPSession, string) {
	now := p.Clock.NowUTC()
	accessEvent := auth.NewAccessEvent(now, p.Request, p.ServerConfig.TrustProxy)
	// NOTE(louis): remember to update the mock provider
	// if session has new fields.
	session := &IDPSession{
		ID:        uuid.New(),
		CreatedAt: now,
		Attrs:     *attrs,
		AccessInfo: auth.AccessInfo{
			InitialAccess: accessEvent,
			LastAccess:    accessEvent,
		},
	}
	token := p.generateToken(session)

	return session, token
}

func (p *Provider) Create(session *IDPSession) error {
	expiry := computeSessionStorageExpiry(session, p.Config)
	err := p.Store.Create(session, expiry)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create session")
	}

	err = p.AccessEvents.InitStream(session)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to access session")
	}

	return nil
}

func (p *Provider) GetByToken(token string) (*IDPSession, error) {
	id, ok := decodeTokenSessionID(token)
	if !ok {
		return nil, ErrSessionNotFound
	}

	s, err := p.Store.Get(id)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) {
			err = errors.HandledWithMessage(err, "failed to get session")
		}
		return nil, err
	}

	if s.TokenHash == "" {
		return nil, ErrSessionNotFound
	}

	if !matchTokenHash(s.TokenHash, token) {
		return nil, ErrSessionNotFound
	}

	if checkSessionExpired(s, p.Clock.NowUTC(), p.Config) {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *Provider) Get(id string) (*IDPSession, error) {
	session, err := p.Store.Get(id)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) {
			err = errors.HandledWithMessage(err, "failed to get session")
		}
		return nil, err
	}

	return session, nil
}

func (p *Provider) Update(sess *IDPSession) error {
	expiry := computeSessionStorageExpiry(sess, p.Config)
	err := p.Store.Update(sess, expiry)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to update session")
	}
	return err
}

func (p *Provider) generateToken(s *IDPSession) string {
	token := encodeToken(s.ID, corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.Random))
	s.TokenHash = crypto.SHA256String(token)
	return token
}

func matchTokenHash(expectedHash, inputToken string) bool {
	inputHash := crypto.SHA256String(inputToken)
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(inputHash)) == 1
}

func encodeToken(id string, token string) string {
	return fmt.Sprintf("%s.%s", id, token)
}

func decodeTokenSessionID(token string) (id string, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return
	}
	id, ok = parts[0], true
	return
}
