package webapp

import (
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SessionStore interface {
	Get(id string) (session *webapp.Session, err error)
	Update(session *webapp.Session) (err error)
}

type GlobalSessionServiceFactory struct {
	Clock       clock.Clock
	RedisHandle *appredis.Handle
}

func (f *GlobalSessionServiceFactory) NewGlobalSessionService(appID config.AppID) *GlobalSessionService {
	return newGlobalSessionService(appID, f.Clock, f.RedisHandle)
}

type GlobalSessionService struct {
	SessionStore SessionStore
	Publisher    *Publisher
	Clock        clock.Clock
}

func (s *GlobalSessionService) GetSession(sessionID string) (session *webapp.Session, err error) {
	return s.SessionStore.Get(sessionID)
}

func (s *GlobalSessionService) UpdateSession(session *webapp.Session) error {
	now := s.Clock.NowUTC()
	session.UpdatedAt = now
	err := s.SessionStore.Update(session)
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
