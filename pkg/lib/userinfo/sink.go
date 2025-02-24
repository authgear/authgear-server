package userinfo

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
)

type Sink struct {
	UserInfoService *UserInfoService
}

func (s *Sink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) error {
	return nil
}

func (s *Sink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error {
	payload := e.Payload.(event.NonBlockingPayload)
	userIDs := []string{}
	userIDs = append(userIDs, payload.RequireReindexUserIDs()...)
	userIDs = append(userIDs, payload.DeletedUserIDs()...)

	for _, userID := range userIDs {
		err := s.UserInfoService.PurgeUserInfo(ctx, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
