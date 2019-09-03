package session

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
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
	clientConfigs map[string]config.APIClientConfiguration

	time time.Provider
	rand *rand.Rand
}

func NewProvider(req *http.Request, store Store, eventStore EventStore, authContext auth.ContextGetter, clientConfigs map[string]config.APIClientConfiguration) Provider {
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

func (p *providerImpl) Create(userID string, principalID string) (s *auth.Session, err error) {
	now := p.time.NowUTC()
	clientID := p.authContext.AccessKey().ClientID
	clientConfig := p.clientConfigs[clientID]

	accessEvent := newAccessEvent(now, p.req)
	sess := auth.Session{
		ID:          uuid.New(),
		ClientID:    clientID,
		UserID:      userID,
		PrincipalID: principalID,

		InitialAccess: accessEvent,

		CreatedAt:  now,
		AccessedAt: now,
	}
	if !clientConfig.RefreshTokenDisabled {
		p.generateRefreshToken(&sess)
	}
	p.generateAccessToken(&sess)

	expiry := computeSessionStorageExpiry(&sess, clientConfig)
	err = p.store.Create(&sess, expiry.Sub(now))
	if err != nil {
		return
	}

	err = p.eventStore.AppendAccessEvent(&sess, &accessEvent)
	if err != nil {
		return
	}

	return &sess, nil
}

func (p *providerImpl) GetByToken(token string, kind auth.SessionTokenKind) (*auth.Session, error) {
	id, ok := decodeTokenSessionID(token)
	if !ok {
		return nil, ErrSessionNotFound
	}

	s, err := p.store.Get(id)
	if err != nil {
		return nil, err
	}

	var expectedToken string
	switch kind {
	case auth.SessionTokenKindAccessToken:
		expectedToken = s.AccessToken
	case auth.SessionTokenKindRefreshToken:
		expectedToken = s.RefreshToken
	default:
		return nil, ErrSessionNotFound
	}

	if expectedToken == "" {
		return nil, ErrSessionNotFound
	}

	if subtle.ConstantTimeCompare([]byte(expectedToken), []byte(token)) == 0 {
		return nil, ErrSessionNotFound
	}

	accessKey := p.authContext.AccessKey()
	// microservices may allow no access key, when rendering HTML pages at server
	// check client ID only if client ID is present (i.e. an access key is used)
	if accessKey.ClientID != "" && s.ClientID != accessKey.ClientID {
		return nil, ErrSessionNotFound
	}

	clientConfig, clientExists := p.clientConfigs[s.ClientID]
	// if client does not exist or is disabled, ignore the session
	if !clientExists || clientConfig.Disabled {
		return nil, ErrSessionNotFound
	}
	if checkSessionExpired(s, p.time.NowUTC(), clientConfig, kind) {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *providerImpl) Access(s *auth.Session) error {
	now := p.time.NowUTC()
	accessEvent := newAccessEvent(now, p.req)

	s.AccessedAt = now
	s.LastAccess = accessEvent

	err := p.eventStore.AppendAccessEvent(s, &accessEvent)
	if err != nil {
		return err
	}

	expiry := computeSessionStorageExpiry(s, p.clientConfigs[s.ClientID])
	return p.store.Update(s, expiry.Sub(now))
}

func (p *providerImpl) Invalidate(id string) error {
	return p.store.Delete(id)
}

func (p *providerImpl) Refresh(session *auth.Session) error {
	p.generateAccessToken(session)

	expiry := computeSessionStorageExpiry(session, p.clientConfigs[session.ClientID])
	return p.store.Update(session, expiry.Sub(p.time.NowUTC()))
}

func (p *providerImpl) generateAccessToken(s *auth.Session) {
	accessToken := corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand)
	s.AccessToken = encodeToken(s.ID, accessToken)
	s.AccessTokenCreatedAt = p.time.NowUTC()
	return
}

func (p *providerImpl) generateRefreshToken(s *auth.Session) {
	refreshToken := corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand)
	s.RefreshToken = encodeToken(s.ID, refreshToken)
	return
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

	extraData := []byte(req.Header.Get(corehttp.HeaderSessionExtraInfo))
	extra := auth.SessionAccessEventExtraInfo{}
	if len(extraData) <= extraDataSizeLimit {
		json.Unmarshal(extraData, &extra)
	}

	return auth.SessionAccessEvent{
		Timestamp: timestamp,
		Remote:    remote,
		UserAgent: req.UserAgent(),
		Extra:     extra,
	}
}
