package session

import (
	"crypto/subtle"
	"fmt"
	"math/rand"
	"strings"

	corerand "github.com/skygeario/skygear-server/pkg/core/rand"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

const (
	tokenAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tokenLength   = 32
)

type providerImpl struct {
	store Store

	time time.Provider
	rand *rand.Rand
}

func NewProvider(store Store) Provider {
	return &providerImpl{
		store: store,
		time:  time.NewProvider(),
		rand:  corerand.SecureRand,
	}
}

func (p *providerImpl) Create(userID string, principalID string) (s *Session, err error) {
	now := p.time.NowUTC()
	sess := Session{
		ID:          uuid.New(),
		UserID:      userID,
		PrincipalID: principalID,

		CreatedAt:  now,
		AccessedAt: now,
	}
	p.generateAccessToken(&sess)

	err = p.store.Create(&sess)
	if err != nil {
		return
	}

	return &sess, nil
}

func (p *providerImpl) GetByToken(token string, kind TokenKind) (*Session, error) {
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
	case TokenKindAccessToken:
		expectedToken = s.AccessToken
	default:
		return nil, ErrSessionNotFound
	}

	if subtle.ConstantTimeCompare([]byte(expectedToken), []byte(token)) == 0 {
		return nil, ErrSessionNotFound
	}

	return s, nil
}

func (p *providerImpl) Access(s *Session) error {
	s.AccessedAt = p.time.NowUTC()
	return p.store.Update(s)
}

func (p *providerImpl) Invalidate(id string) error {
	return p.store.Delete(id)
}

func (p *providerImpl) Refresh(session *Session) error {
	p.generateAccessToken(session)
	return p.store.Update(session)
}

func (p *providerImpl) generateAccessToken(s *Session) {
	accessToken := corerand.StringWithAlphabet(tokenLength, tokenAlphabet, p.rand)
	s.AccessToken = encodeToken(s.ID, accessToken)
	s.AccessTokenCreatedAt = p.time.NowUTC()
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
