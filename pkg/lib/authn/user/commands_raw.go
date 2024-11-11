package user

import (
	"context"
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

func (c *RawCommands) Create(ctx context.Context, userID string) (*User, error) {
	user := c.New(userID)

	err := c.Store.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *RawCommands) AfterCreate(userModel *model.User, identities []*identity.Info) error {
	return nil
}

func (c *RawCommands) UpdateLoginTime(ctx context.Context, userID string, loginAt time.Time) error {
	return c.Store.UpdateLoginTime(ctx, userID, loginAt)
}

func (c *RawCommands) UpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error {
	return c.Store.UpdateMFAEnrollment(ctx, userID, endAt)
}

func (c *RawCommands) UpdateAccountStatus(ctx context.Context, userID string, accountStatus AccountStatus) error {
	return c.Store.UpdateAccountStatus(ctx, userID, accountStatus)
}

func (c *RawCommands) Delete(ctx context.Context, userID string) error {
	return c.Store.Delete(ctx, userID)
}

func (c *RawCommands) Anonymize(ctx context.Context, userID string) error {
	return c.Store.Anonymize(ctx, userID)
}
