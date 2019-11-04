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

type providerImpl struct {
	req           *http.Request
	store         Store
	eventStore    EventStore
	authContext   auth.ContextGetter
	clientConfigs []config.APIClientConfiguration

	time time.Provider
	rand *rand.Rand
}

func NewProvider(req *http.Request, store Store, eventStore EventStore, authContext auth.ContextGetter, clientConfigs []config.APIClientConfiguration) Provider {
	return &providerImpl{
		req:           req,
		store:         store,
		eventStore:    eventStore,
		authContext:   authContext,
		clientConfigs: clientConfigs,
		time:          time.NewProvider(),
		rand:          corerand.SecureRand,
	}
}

func (p *providerImpl) Create(authnSess *auth.AuthnSession) (*auth.Session, auth.SessionTokens, error) {
	now := p.time.NowUTC()
	clientID := p.authContext.AccessKey().ClientID
	clientConfig, _ := model.GetClientConfig(p.clientConfigs, clientID)
	accessEvent := newAccessEvent(now, p.req)
	// NOTE(louis): remember to update the mock provider
	// if session has new fields.
	sess := auth.Session{
		ID:                      uuid.New(),
		ClientID:                authnSess.ClientID,
		UserID:                  authnSess.UserID,
		PrincipalID:             authnSess.PrincipalID,
		PrincipalType:           authnSess.PrincipalType,
		PrincipalUpdatedAt:      authnSess.PrincipalUpdatedAt,
		AuthenticatorID:         authnSess.AuthenticatorID,
		AuthenticatorType:       authnSess.AuthenticatorType,
		AuthenticatorOOBChannel: authnSess.AuthenticatorOOBChannel,
		AuthenticatorUpdatedAt:  authnSess.AuthenticatorUpdatedAt,
		InitialAccess:           accessEvent,
		LastAccess:              accessEvent,
		CreatedAt:               now,
		AccessedAt:              now,
	}
	tok := auth.SessionTokens{ID: sess.ID}
	if !clientConfig.RefreshTokenDisabled {
		tok.RefreshToken = p.generateRefreshToken(&sess)
	}
	tok.AccessToken = p.generateAccessToken(&sess)

	expiry := computeSessionStorageExpiry(&sess, *clientConfig)
	err := p.store.Create(&sess, expiry)
	if err != nil {
		return nil, tok, errors.HandledWithMessage(err, "failed to create session")
	}

	err = p.eventStore.AppendAccessEvent(&sess, &accessEvent)
	if err != nil {
		return nil, tok, errors.HandledWithMessage(err, "failed to access session")
	}

	return &sess, tok, nil
}

func (p *providerImpl) GetByToken(token string, kind auth.SessionTokenKind) (*auth.Session, error) {
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

	var expectedHash string
	switch kind {
	case auth.SessionTokenKindAccessToken:
		expectedHash = s.AccessTokenHash
	case auth.SessionTokenKindRefreshToken:
		expectedHash = s.RefreshTokenHash
	default:
		panic("session: unexpected token kind: " + kind)
	}

	if expectedHash == "" {
		return nil, ErrSessionNotFound
	}

	if !matchTokenHash(expectedHash, token) {
		return nil, ErrSessionNotFound
	}

	accessKey := p.authContext.AccessKey()
	// microservices may allow no access key, when rendering HTML pages at server
	// check client ID only if client ID is present (i.e. an access key is used)
	if accessKey.ClientID != "" && s.ClientID != accessKey.ClientID {
		return nil, ErrSessionNotFound
	}

	clientConfig, clientExists := model.GetClientConfig(p.clientConfigs, s.ClientID)
	// if client does not exist or is disabled, ignore the session
	if !clientExists || clientConfig.Disabled {
		return nil, ErrSessionNotFound
	}
	if checkSessionExpired(s, p.time.NowUTC(), *clientConfig, kind) {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *providerImpl) Get(id string) (*auth.Session, error) {
	session, err := p.store.Get(id)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) {
			err = errors.HandledWithMessage(err, "failed to get session")
		}
		return nil, err
	}

	currentSession, _ := p.authContext.Session()
	if currentSession != nil && session.ID == currentSession.ID {
		// should use current session data instead
		session = currentSession
	}

	return session, nil
}

func (p *providerImpl) Access(s *auth.Session) error {
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

func (p *providerImpl) Invalidate(session *auth.Session) error {
	err := p.store.Delete(session)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate session")
	}
	return nil
}

func (p *providerImpl) InvalidateBatch(sessions []*auth.Session) error {
	err := p.store.DeleteBatch(sessions)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate sessions")
	}
	return nil
}

func (p *providerImpl) InvalidateAll(userID string, sessionID string) error {
	err := p.store.DeleteAll(userID, sessionID)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to invalidate sessions")
	}
	return nil
}

func (p *providerImpl) List(userID string) (sessions []*auth.Session, err error) {
	storedSessions, err := p.store.List(userID)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to list sessions")
		return
	}

	now := p.time.NowUTC()
	currentSession, _ := p.authContext.Session()
	for _, session := range storedSessions {
		clientConfig, clientExists := model.GetClientConfig(p.clientConfigs, session.ClientID)
		// if client does not exist or is disabled, ignore the session
		if !clientExists || clientConfig.Disabled {
			continue
		}

		maxExpiry := computeSessionStorageExpiry(session, *clientConfig)
		// ignore expired sessions
		if now.After(maxExpiry) {
			continue
		}

		if currentSession != nil && session.ID == currentSession.ID {
			// should use current session data instead
			session = currentSession
		}

		sessions = append(sessions, session)
	}
	return
}

func (p *providerImpl) Refresh(session *auth.Session) (string, error) {
	accessToken := p.generateAccessToken(session)
	clientConfig, _ := model.GetClientConfig(p.clientConfigs, session.ClientID)

	expiry := computeSessionStorageExpiry(session, *clientConfig)
	err := p.store.Update(session, expiry)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to refresh session")
	}
	return accessToken, err
}

func (p *providerImpl) UpdateMFA(sess *auth.Session, opts auth.AuthnSessionStepMFAOptions) error {
	now := p.time.NowUTC()
	sess.AuthenticatorID = opts.AuthenticatorID
	sess.AuthenticatorType = opts.AuthenticatorType
	sess.AuthenticatorOOBChannel = opts.AuthenticatorOOBChannel
	sess.AuthenticatorUpdatedAt = &now
	_, err := p.Refresh(sess)
	return err
}

func (p *providerImpl) generateAccessToken(s *auth.Session) string {
	accessToken := encodeToken(s.ID, corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand))
	s.AccessTokenHash = crypto.SHA256String(accessToken)
	s.AccessTokenCreatedAt = p.time.NowUTC()
	return accessToken
}

func (p *providerImpl) generateRefreshToken(s *auth.Session) string {
	refreshToken := encodeToken(s.ID, corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand))
	s.RefreshTokenHash = crypto.SHA256String(refreshToken)
	return refreshToken
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

func newAccessEvent(timestamp gotime.Time, req *http.Request) auth.SessionAccessEvent {
	remote := auth.SessionAccessEventConnInfo{
		RemoteAddr:    req.RemoteAddr,
		XForwardedFor: req.Header.Get("X-Forwarded-For"),
		XRealIP:       req.Header.Get("X-Real-IP"),
		Forwarded:     req.Header.Get("Forwarded"),
	}

	extra := auth.SessionAccessEventExtraInfo{}
	extraData, err := base64.StdEncoding.DecodeString(req.Header.Get(corehttp.HeaderSessionExtraInfo))
	if err == nil && len(extraData) <= extraDataSizeLimit {
		_ = json.Unmarshal(extraData, &extra)
	}

	return auth.SessionAccessEvent{
		Timestamp: timestamp,
		Remote:    remote,
		UserAgent: req.UserAgent(),
		Extra:     extra,
	}
}
