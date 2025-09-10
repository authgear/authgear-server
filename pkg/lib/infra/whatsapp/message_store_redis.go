package whatsapp

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type MessageStore struct {
	// We use global redis because the whatsapp credential might be shared between different apps
	Redis       *globalredis.Handle
	Credentials *config.WhatsappCloudAPICredentials
}

func redisMessageStatusKey(phoneNumberID string, messageID string) string {
	return fmt.Sprintf("whatsapp:phone-number-id:%s:message-id:%s", phoneNumberID, messageID)
}

func (s *MessageStore) UpdateMessageStatus(ctx context.Context, messageID string, status WhatsappMessageStatus) error {
	key := redisMessageStatusKey(s.Credentials.PhoneNumberID, messageID)
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, key, string(status), duration.UserInteraction).Result()
		return err
	})
}

func (s *MessageStore) GetMessageStatus(ctx context.Context, messageID string) (WhatsappMessageStatus, error) {
	key := redisMessageStatusKey(s.Credentials.PhoneNumberID, messageID)
	var status WhatsappMessageStatus
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, key).Result()
		if errors.Is(err, goredis.Nil) {
			return nil
		} else if err != nil {
			return err
		}
		status = WhatsappMessageStatus(data)
		return nil
	})
	return status, err
}
