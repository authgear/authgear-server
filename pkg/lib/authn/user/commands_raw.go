package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
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

func (c *RawCommands) New(userID string) *User {
	now := c.Clock.NowUTC()
	user := &User{
		ID:            userID,
		Labels:        make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
		LastLoginAt:   nil,
		IsDisabled:    false,
		DisableReason: nil,
	}
	return user
}

func (c *RawCommands) Create(userID string) (*User, error) {
	user := c.New(userID)

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
