package password

import (
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type HistoryStore struct {
	Clock       clock.Clock
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (p *HistoryStore) CreatePasswordHistory(userID string, hashedPassword []byte, createdAt time.Time) error {
	updateBuilder := p.insertPasswordHistoryBuilder(
		userID,
		hashedPassword,
		createdAt,
	)
	if _, err := p.SQLExecutor.ExecWith(updateBuilder); err != nil {
		return err
	}
	return nil
}

func (p *HistoryStore) GetPasswordHistory(userID string, historySize int, historyDays config.DurationDays) ([]History, error) {
	var err error
	var sizeHistory, daysHistory []History
	t := p.Clock.NowUTC()

	if historySize > 0 {
		sizeBuilder := p.basePasswordHistoryBuilder(userID).Limit(uint64(historySize))
		sizeHistory, err = p.doQueryPasswordHistory(sizeBuilder)
		if err != nil {
			return nil, err
		}
	}

	if historyDays > 0 {
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		since := startOfDay.Add(-historyDays.Duration())
		daysBuilder := p.basePasswordHistoryBuilder(userID).
			Where("created_at >= ?", since)
		daysHistory, err = p.doQueryPasswordHistory(daysBuilder)
		if err != nil {
			return nil, err
		}
	}

	if len(sizeHistory) > len(daysHistory) {
		return sizeHistory, nil
	}

	return daysHistory, nil
}

func (p *HistoryStore) RemovePasswordHistory(userID string, historySize int, historyDays config.DurationDays) error {
	history, err := p.GetPasswordHistory(userID, historySize, historyDays)
	if err != nil {
		return err
	}

	if len(history) <= 0 {
		return nil
	}

	oldestTime := history[len(history)-1].CreatedAt
	ids := []interface{}{}
	for _, h := range history {
		ids = append(ids, h.ID)
	}

	builder := p.SQLBuilder.Tenant().
		Delete(p.SQLBuilder.FullTableName("password_history")).
		Where("user_id = ?", userID).
		Where("id != ALL (?)", pq.Array(ids)).
		Where("created_at < ?", oldestTime)

	_, err = p.SQLExecutor.ExecWith(builder)
	return err
}

func (p *HistoryStore) basePasswordHistoryBuilder(userID string) db.SelectBuilder {
	return p.SQLBuilder.Tenant().
		Select("id", "user_id", "password", "created_at").
		From(p.SQLBuilder.FullTableName("password_history")).
		Where("user_id = ?", userID).
		OrderBy("created_at DESC")
}

func (p *HistoryStore) insertPasswordHistoryBuilder(userID string, hashedPassword []byte, createdAt time.Time) db.InsertBuilder {
	return p.SQLBuilder.Tenant().
		Insert(p.SQLBuilder.FullTableName("password_history")).
		Columns(
			"id",
			"user_id",
			"password",
			"created_at",
		).
		Values(
			uuid.New(),
			userID,
			hashedPassword,
			createdAt,
		)
}

func (p *HistoryStore) doQueryPasswordHistory(builder db.SelectBuilder) ([]History, error) {
	rows, err := p.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []History{}
	for rows.Next() {
		var (
			id                string
			userID            string
			hashedPasswordStr string
			createdAt         time.Time
		)
		if err := rows.Scan(&id, &userID, &hashedPasswordStr, &createdAt); err != nil {
			return nil, err
		}
		passwordHistory := History{
			ID:             id,
			UserID:         userID,
			HashedPassword: []byte(hashedPasswordStr),
			CreatedAt:      createdAt,
		}
		out = append(out, passwordHistory)
	}
	return out, nil
}
