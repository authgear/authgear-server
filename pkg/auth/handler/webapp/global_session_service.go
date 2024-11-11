package webapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type SessionStore interface {
	Get(ctx context.Context, id string) (session *webapp.Session, err error)
	Update(ctx context.Context, session *webapp.Session) (err error)
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

func (s *GlobalSessionService) GetSession(ctx context.Context, sessionID string) (session *webapp.Session, err error) {
	return s.SessionStore.Get(ctx, sessionID)
}

func (s *GlobalSessionService) UpdateSession(ctx context.Context, session *webapp.Session) error {
	now := s.Clock.NowUTC()
	session.UpdatedAt = now
	err := s.SessionStore.Update(ctx, session)
	if err != nil {
		return err
	}

	msg := &WebsocketMessage{
		Kind: WebsocketMessageKindRefresh,
	}

	err = s.Publisher.Publish(ctx, session, msg)
	if err != nil {
		return err
	}

	return nil
}
