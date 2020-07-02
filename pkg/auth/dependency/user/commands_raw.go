package user

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
)

type WelcomeMessageProvider interface {
	SendToIdentityInfos(infos []*identity.Info) error
}

type RawCommands struct {
	Store                  store
	Clock                  clock.Clock
	WelcomeMessageProvider WelcomeMessageProvider
	Queries                *Queries
}

func (c *RawCommands) Create(userID string, metadata map[string]interface{}) (*User, error) {
	now := c.Clock.NowUTC()
	user := &User{
		ID:          userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: nil,
		Metadata:    metadata,
	}

	err := c.Store.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *RawCommands) AfterCreate(userModel *model.User, identities []*identity.Info) error {
	err := c.WelcomeMessageProvider.SendToIdentityInfos(identities)
	if err != nil {
		return err
	}

	return nil
}

func (c *RawCommands) UpdateMetadata(user *model.User, metadata map[string]interface{}) error {
	now := c.Clock.NowUTC()
	if err := c.Store.UpdateMetadata(user.ID, metadata, now); err != nil {
		return err
	}

	user.Metadata = metadata
	return nil
}

func (c *RawCommands) UpdateLoginTime(user *model.User, loginAt time.Time) error {
	err := c.Store.UpdateLoginTime(user.ID, loginAt)
	if err != nil {
		return err
	}

	u, err := c.Queries.Get(user.ID)
	if err != nil {
		return err
	}

	*user = *u
	return nil
}
