package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type MessageStore struct {
	// We use global redis because the whatsapp credential might be shared between different apps
	Redis       *globalredis.Handle
	Credentials *config.WhatsappCloudAPICredentials
}

type WhatsappMessageStatusData struct {
	Status WhatsappMessageStatus `json:"status"`
	Errors []WhatsappStatusError `json:"errors"`
}

func redisMessageStatusKey(phoneNumberID string, messageID string) string {
	hashedPhoneNumberID := crypto.SHA256String(phoneNumberID)
	return fmt.Sprintf("whatsapp:phone-number-id-sha256:%s:message-id:%s", hashedPhoneNumberID, messageID)
}

func (s *MessageStore) UpdateMessageStatus(ctx context.Context, messageID string, status *WhatsappMessageStatusData) error {
	key := redisMessageStatusKey(s.Credentials.PhoneNumberID, messageID)
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			panic(fmt.Errorf("unexpected: failed to marshal WhatsappMessageStatusData"))
		}

		_, err = conn.Set(ctx, key, string(statusBytes), duration.UserInteraction).Result()
		return err
	})
}

func (s *MessageStore) GetMessageStatus(ctx context.Context, messageID string) (*WhatsappMessageStatusData, error) {
	key := redisMessageStatusKey(s.Credentials.PhoneNumberID, messageID)
	var data WhatsappMessageStatusData
	var ok bool = false
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		dataBytes, err := conn.Get(ctx, key).Result()
		if errors.Is(err, goredis.Nil) {
			return nil
		} else if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(dataBytes), &data)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to unmarshal WhatsappMessageStatusData"), err)
		}
		ok = true
		return nil
	})
	if !ok {
		return nil, err
	}
	return &data, err
}
