package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type WebSessionStore interface {
	Get(id string) (session *Session, err error)
	Update(session *Session) (err error)
}

type GlobalSessionServiceFactory struct {
	Clock       clock.Clock
	RedisHandle *appredis.Handle
}

func (f *GlobalSessionServiceFactory) NewGlobalSessionService(appID config.AppID) *GlobalSessionService {
	return newGlobalSessionService(appID, f.Clock, f.RedisHandle)
}

type GlobalSessionService struct {
	WebSessionStore WebSessionStore
	Publisher       *Publisher
	Clock           clock.Clock
}

func (s *GlobalSessionService) GetSession(sessionID string) (session *Session, err error) {
	return s.WebSessionStore.Get(sessionID)
}

func (s *GlobalSessionService) UpdateSession(session *Session) error {
	now := s.Clock.NowUTC()
	session.UpdatedAt = now
	err := s.WebSessionStore.Update(session)
	if err != nil {
		return err
	}

	msg := &WebsocketMessage{
		Kind: WebsocketMessageKindRefresh,
	}

	err = s.Publisher.Publish(session, msg)
	if err != nil {
		return err
	}

	return nil
}
