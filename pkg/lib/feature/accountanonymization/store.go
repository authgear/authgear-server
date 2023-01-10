package accountanonymization

import (
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
// it is the time to execute the scheduled deletion.
func (s *Store) ListAppUsers() (appUsers []AppUser, err error) {
	now := s.Clock.NowUTC()
	err = s.Handle.ReadOnly(func() (err error) {
		q := s.SQLBuilder.
			Select("app_id", "id").
			From(s.SQLBuilder.TableName("_auth_user")).
			Where("anonymize_at < ?", now).
			Where("is_anonymized IS FALSE")
		rows, err := s.SQLExecutor.QueryWith(q)
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
