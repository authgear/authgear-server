package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type RawCommands struct {
	Store store
	Clock clock.Clock
}

func (c *RawCommands) New(userID string) *User {
	now := c.Clock.NowUTC()
	user := &User{
		ID:                  userID,
		CreatedAt:           now,
		UpdatedAt:           now,
		MostRecentLoginAt:   nil,
		LessRecentLoginAt:   nil,
		IsDisabled:          false,
		DisableReason:       nil,
		StandardAttributes:  make(map[string]interface{}),
		RequireReindexAfter: &now,
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
	return nil
}

func (c *RawCommands) UpdateLoginTime(userID string, loginAt time.Time) error {
	return c.Store.UpdateLoginTime(userID, loginAt)
}

func (c *RawCommands) UpdateMFAEnrollment(userID string, endAt *time.Time) error {
	return c.Store.UpdateMFAEnrollment(userID, endAt)
}

func (c *RawCommands) UpdateAccountStatus(userID string, accountStatus AccountStatus) error {
	return c.Store.UpdateAccountStatus(userID, accountStatus)
}

func (c *RawCommands) Delete(userID string) error {
	return c.Store.Delete(userID)
}

func (c *RawCommands) Anonymize(userID string) error {
	return c.Store.Anonymize(userID)
}
