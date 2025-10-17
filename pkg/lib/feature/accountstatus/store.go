package accountstatus

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Store struct {
	Handle      *globaldb.Handle
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
	Clock       clock.Clock
}

type AppUser struct {
	AppID  string
	UserID string
}

// ListAppUsers returns a list of (appID, userID) pairs that
// it is the time to update account status.
func (s *Store) ListAppUsers(ctx context.Context) (appUsers []AppUser, err error) {
	now := s.Clock.NowUTC()
	err = s.Handle.ReadOnly(ctx, func(ctx context.Context) (err error) {
		q := s.SQLBuilder.
			Select("app_id", "id").
			From(s.SQLBuilder.TableName("_auth_user")).
			Where("account_status_stale_from IS NOT NULL AND account_status_stale_from < ?", now)
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return
		}
		for rows.Next() {
			var appUser AppUser
			err = rows.Scan(
				&appUser.AppID,
				&appUser.UserID,
			)
			if err != nil {
				return
			}
			appUsers = append(appUsers, appUser)
		}
		return
	})
	if err != nil {
		return
	}

	return
}
