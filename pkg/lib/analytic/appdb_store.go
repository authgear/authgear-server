package analytic

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type AppDBStore struct {
	SQLBuilder  *appdb.SQLBuilder
	SQLExecutor *appdb.SQLExecutor
}

type User struct {
	ID    string
	Email string
}

func (s *AppDBStore) GetAllUsers(appID string) ([]*User, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select(
			"id",
			"standard_attributes ->> 'email'",
		).
		From(s.SQLBuilder.TableName("_auth_user"))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		var id string
		var email sql.NullString
		err = rows.Scan(
			&id,
			&email,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &User{
			ID:    id,
			Email: email.String,
		})
	}

	return result, nil
}

func (s *AppDBStore) GetNewUserIDs(appID string, rangeFrom *time.Time, rangeTo *time.Time) ([]string, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select(
			"id",
		).
		From(s.SQLBuilder.TableName("_auth_user")).
		Where("created_at >= ?", rangeFrom).
		Where("created_at < ?", rangeTo)
	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var userID string
		err = rows.Scan(
			&userID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, userID)
	}
	return result, nil
}

// GetUserVerifiedEmails returns userID to email map
func (s *AppDBStore) GetUserVerifiedEmails(appID string, userIDs []string) (result map[string]string, err error) {
	getUserVerifiedEmails := func(appID string, userIDs []string, result map[string]string) error {
		builder := s.SQLBuilder.WithAppID(appID).
			Select(
				"user_id",
				"value",
			).
			From(s.SQLBuilder.TableName("_auth_verified_claim")).
			Where("name = ?", "email").
			Where(sq.Eq{"user_id": userIDs})

		rows, err := s.SQLExecutor.QueryWith(builder)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var userID string
			var email string
			err = rows.Scan(
				&userID,
				&email,
			)
			if err != nil {
				return err
			}
			result[userID] = email
		}
		return nil
	}

	result = map[string]string{}
	batchSize := 50
	for i := 0; i < len(userIDs); i += batchSize {
		j := i + batchSize
		if j > len(userIDs) {
			j = len(userIDs)
		}
		batch := userIDs[i:j]

		err = getUserVerifiedEmails(appID, batch, result)
		if err != nil {
			return
		}
	}

	return
}

func (s *AppDBStore) GetUserCountBeforeTime(appID string, beforeTime *time.Time) (int, error) {
	builder := s.SQLBuilder.WithAppID(appID).
		Select("count(*)").
		From(s.SQLBuilder.TableName("_auth_user")).
		Where("created_at < ?", beforeTime)
	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return 0, err
	}
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
