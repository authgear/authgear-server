package challenge

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/redis"
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

	conn := p.Redis.Conn()
	key := challengeKey(p.AppID, c.Token)
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	_, err = redigo.String(conn.Do("SET", key, data, "PX", toMilliseconds(ttl), "NX"))
	if errors.Is(err, redigo.ErrNil) {
		return nil, errors.New("fail to create new challenge")
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func (p *Provider) Consume(token string) (*Purpose, error) {
	conn := p.Redis.Conn()
	key := challengeKey(p.AppID, token)
	data, err := redigo.Bytes(conn.Do("GET", key))
	if errors.Is(err, redigo.ErrNil) {
		return nil, ErrInvalidChallenge
	} else if err != nil {
		return nil, err
	}

	c := &Challenge{}
	err = json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	_, err = conn.Do("DEL", key)
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
