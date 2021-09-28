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
}

func (c *RawCommands) New(userID string) *User {
	now := c.Clock.NowUTC()
	user := &User{
		ID:                userID,
		Labels:            make(map[string]interface{}),
		CreatedAt:         now,
		UpdatedAt:         now,
		MostRecentLoginAt: nil,
		LessRecentLoginAt: nil,
		IsDisabled:        false,
		DisableReason:     nil,
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

func (c *RawCommands) UpdateLoginTime(userID string, loginAt time.Time) error {
	return c.Store.UpdateLoginTime(userID, loginAt)
}

func (c *RawCommands) UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error {
	return c.Store.UpdateDisabledStatus(userID, isDisabled, reason)
}

func (c *RawCommands) Delete(userID string) error {
	return c.Store.Delete(userID)
}
