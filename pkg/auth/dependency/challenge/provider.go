package challenge

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/redis"
)

type Provider struct {
	Redis *redis.Context
	AppID config.AppID
	Clock clock.Clock
}

func (p *Provider) Create(purpose Purpose) (*Challenge, error) {
	now := p.Clock.NowUTC()
	ttl := purpose.ValidityPeriod()
	c := &Challenge{
		Token:     GenerateChallengeToken(),
		Purpose:   purpose,
		CreatedAt: now,
		ExpireAt:  now.Add(ttl),
	}

	key := challengeKey(p.AppID, c.Token)
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	err = p.Redis.WithConn(func(conn redis.Conn) error {
		_, err = redigo.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "NX"))
		if errors.Is(err, redigo.ErrNil) {
			return errors.New("fail to create new challenge")
		} else if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Provider) Consume(token string) (*Purpose, error) {
	key := challengeKey(p.AppID, token)

	c := &Challenge{}

	err := p.Redis.WithConn(func(conn redis.Conn) error {
		data, err := redigo.Bytes(conn.Do("GET", key))
		if errors.Is(err, redigo.ErrNil) {
			return ErrInvalidChallenge
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, c)
		if err != nil {
			return err
		}

		_, err = conn.Do("DEL", key)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &c.Purpose, nil
}

func challengeKey(appID config.AppID, token string) string {
	return fmt.Sprintf("%s:challenge:%s", appID, token)
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}
