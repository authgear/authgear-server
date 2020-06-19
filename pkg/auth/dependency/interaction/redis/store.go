package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

func interactionKey(appID, token string) string {
	return fmt.Sprintf("%s:interaction:%s", appID, token)
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

type Store struct {
	Context context.Context
	AppID   string
	Clock   clock.Clock
}

func (s *Store) Create(i *interaction.Interaction) error {
	json, err := json.Marshal(i)
	if err != nil {
		return err
	}

	conn := redis.GetConn(s.Context)
	ttl := i.ExpireAt.Sub(s.Clock.NowUTC())
	key := interactionKey(s.AppID, i.Token)

	_, err = redigo.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "NX"))
	if errors.Is(err, redigo.ErrNil) {
		return fmt.Errorf("duplicated token: %w", err)
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Store) Get(token string) (*interaction.Interaction, error) {
	conn := redis.GetConn(s.Context)
	key := interactionKey(s.AppID, token)
	data, err := redigo.Bytes(conn.Do("GET", key))
	if errors.Is(err, redigo.ErrNil) {
		return nil, interaction.ErrInteractionNotFound
	} else if err != nil {
		return nil, err
	}

	i := &interaction.Interaction{}
	err = json.Unmarshal(data, i)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (s *Store) Update(i *interaction.Interaction) error {
	json, err := json.Marshal(i)
	if err != nil {
		return err
	}

	conn := redis.GetConn(s.Context)
	ttl := i.ExpireAt.Sub(s.Clock.NowUTC())
	key := interactionKey(s.AppID, i.Token)

	_, err = redigo.String(conn.Do("SET", key, json, "PX", toMilliseconds(ttl), "XX"))
	if errors.Is(err, redigo.ErrNil) {
		return interaction.ErrInteractionNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(i *interaction.Interaction) error {
	conn := redis.GetConn(s.Context)
	key := interactionKey(s.AppID, i.Token)

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}
