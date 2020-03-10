package session

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/model"
	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

const (
	tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tokenLength   = 32
)

const extraDataSizeLimit = 1024

type ProviderImpl struct {
	req           *http.Request
	store         Store
	eventStore    EventStore
	clientConfigs []config.APIClientConfiguration

	time time.Provider
	rand *rand.Rand
}

func NewProvider(req *http.Request, store Store, eventStore EventStore, clientConfigs []config.APIClientConfiguration) *ProviderImpl {
	return &ProviderImpl{
		req:           req,
		store:         store,
		eventStore:    eventStore,
		clientConfigs: clientConfigs,
		time:          time.NewProvider(),
		rand:          corerand.SecureRand,
	}
}

var _ Provider = &ProviderImpl{}

func (p *ProviderImpl) MakeSession(authnSess *auth.AuthnSession) (*Session, string) {
	now := p.time.NowUTC()
	accessEvent := newAccessEvent(now, p.req)
	// NOTE(louis): remember to update the mock provider
	// if session has new fields.
	session := &Session{
		ID:                      uuid.New(),
		ClientID:                authnSess.ClientID,
		UserID:                  authnSess.UserID,
		PrincipalID:             authnSess.PrincipalID,
		PrincipalType:           authn.PrincipalType(authnSess.PrincipalType),
		PrincipalUpdatedAt:      authnSess.PrincipalUpdatedAt,
		AuthenticatorID:         authnSess.AuthenticatorID,
		AuthenticatorType:       authn.AuthenticatorType(authnSess.AuthenticatorType),
		AuthenticatorOOBChannel: authn.AuthenticatorOOBChannel(authnSess.AuthenticatorOOBChannel),
		AuthenticatorUpdatedAt:  authnSess.AuthenticatorUpdatedAt,
		InitialAccess:           accessEvent,
		LastAccess:              accessEvent,
		CreatedAt:               now,
		AccessedAt:              now,
	}
	token := p.generateToken(session)

	return session, token
}

func (p *ProviderImpl) Create(session *Session) error {
	clientConfig, _ := model.GetClientConfig(p.clientConfigs, session.ClientID)
	expiry := computeSessionStorageExpiry(session, *clientConfig)
	err := p.store.Create(session, expiry)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create session")
	}

	err = p.eventStore.AppendAccessEvent(session, &session.InitialAccess)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to access session")
	}

	return nil
}

func (p *ProviderImpl) GetByToken(token string) (*Session, error) {
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

	clientConfig, clientExists := model.GetClientConfig(p.clientConfigs, s.ClientID)
	// if client does not exist, ignore the session
	if !clientExists {
		return nil, ErrSessionNotFound
	}
	if checkSessionExpired(s, p.time.NowUTC(), *clientConfig) {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *ProviderImpl) Get(id string) (*Session, error) {
	session, err := p.store.Get(id)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) {
			err = errors.HandledWithMessage(err, "failed to get session")
		}
		return nil, err
	}

	return session, nil
}

func (p *ProviderImpl) Access(s *Session) error {
	now := p.time.NowUTC()
	accessEvent := newAccessEvent(now, p.req)

	s.AccessedAt = now
	s.LastAccess = accessEvent

	err := p.eventStore.AppendAccessEvent(s, &accessEvent)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to access session")
	}

	clientConfig, _ := model.GetClientConfig(p.clientConfigs, s.ClientID)

	expiry := computeSessionStorageExpiry(s, *clientConfig)
	err = p.store.Update(s, expiry)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update session")
	}
	return nil
}

func (p *ProviderImpl) Invalidate(session *Session) error {
	err := p.store.Delete(session)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (p *ProviderImpl) InvalidateBatch(sessions []*Session) error {
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

func (p *ProviderImpl) List(userID string) (sessions []*Session, err error) {
	storedSessions, err := p.store.List(userID)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to list sessions")
		return
	}

	now := p.time.NowUTC()
	for _, session := range storedSessions {
		clientConfig, clientExists := model.GetClientConfig(p.clientConfigs, session.ClientID)
		// if client does not exist, ignore the session
		if !clientExists {
			continue
		}

		maxExpiry := computeSessionStorageExpiry(session, *clientConfig)
		// ignore expired sessions
		if now.After(maxExpiry) {
			continue
		}

		sessions = append(sessions, session)
	}
	return
}

func (p *ProviderImpl) Update(sess *Session) error {
	clientConfig, _ := model.GetClientConfig(p.clientConfigs, sess.ClientID)
	expiry := computeSessionStorageExpiry(sess, *clientConfig)
	err := p.store.Update(sess, expiry)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to update session")
	}
	return err
}

func (p *ProviderImpl) generateToken(s *Session) string {
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

func newAccessEvent(timestamp gotime.Time, req *http.Request) AccessEvent {
	remote := AccessEventConnInfo{
		RemoteAddr:    req.RemoteAddr,
		XForwardedFor: req.Header.Get("X-Forwarded-For"),
		XRealIP:       req.Header.Get("X-Real-IP"),
		Forwarded:     req.Header.Get("Forwarded"),
	}

	extra := AccessEventExtraInfo{}
	extraData, err := base64.StdEncoding.DecodeString(req.Header.Get(corehttp.HeaderSessionExtraInfo))
	if err == nil && len(extraData) <= extraDataSizeLimit {
		_ = json.Unmarshal(extraData, &extra)
	}

	return AccessEvent{
		Timestamp: timestamp,
		Remote:    remote,
		UserAgent: req.UserAgent(),
		Extra:     extra,
	}
}
