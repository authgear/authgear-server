package session

import (
	"crypto/subtle"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

const (
	tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tokenLength   = 32
)

type AccessEventProvider interface {
	InitStream(s auth.AuthSession) error
}

type ProviderImpl struct {
	req          *http.Request
	store        Store
	accessEvents AccessEventProvider
	config       config.SessionConfiguration

	time time.Provider
	rand *rand.Rand
}

func NewProvider(
	req *http.Request,
	store Store,
	accessEvents AccessEventProvider,
	sessionConfig config.SessionConfiguration,
) *ProviderImpl {
	return &ProviderImpl{
		req:          req,
		store:        store,
		accessEvents: accessEvents,
		config:       sessionConfig,
		time:         time.NewProvider(),
		rand:         corerand.SecureRand,
	}
}

var _ Provider = &ProviderImpl{}

func (p *ProviderImpl) MakeSession(attrs *authn.Attrs) (*IDPSession, string) {
	now := p.time.NowUTC()
	accessEvent := auth.NewAccessEvent(now, p.req)
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

func (p *ProviderImpl) Create(session *IDPSession) error {
	expiry := computeSessionStorageExpiry(session, p.config)
	err := p.store.Create(session, expiry)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create session")
	}

	err = p.accessEvents.InitStream(session)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to access session")
	}

	return nil
}

func (p *ProviderImpl) GetByToken(token string) (*IDPSession, error) {
	id, ok := decodeTokenSessionID(token)
	if !ok {
		return nil, ErrSessionNotFound
	}

	s, err := p.store.Get(id)
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

	if checkSessionExpired(s, p.time.NowUTC(), p.config) {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *ProviderImpl) Get(id string) (*IDPSession, error) {
	session, err := p.store.Get(id)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) {
			err = errors.HandledWithMessage(err, "failed to get session")
		}
		return nil, err
	}

	return session, nil
}

func (p *ProviderImpl) Invalidate(session *IDPSession) error {
	err := p.store.Delete(session)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (p *ProviderImpl) InvalidateBatch(sessions []*IDPSession) error {
	err := p.store.DeleteBatch(sessions)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate sessions")
	}
	return nil
}

func (p *ProviderImpl) InvalidateAll(userID string, sessionID string) error {
	err := p.store.DeleteAll(userID, sessionID)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate sessions")
	}
	return nil
}

func (p *ProviderImpl) List(userID string) (sessions []*IDPSession, err error) {
	storedSessions, err := p.store.List(userID)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to list sessions")
		return
	}

	now := p.time.NowUTC()
	for _, session := range storedSessions {
		maxExpiry := computeSessionStorageExpiry(session, p.config)
		// ignore expired sessions
		if now.After(maxExpiry) {
			continue
		}

		sessions = append(sessions, session)
	}
	return
}

func (p *ProviderImpl) Update(sess *IDPSession) error {
	expiry := computeSessionStorageExpiry(sess, p.config)
	err := p.store.Update(sess, expiry)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to update session")
	}
	return err
}

func (p *ProviderImpl) generateToken(s *IDPSession) string {
	token := encodeToken(s.ID, corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand))
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
