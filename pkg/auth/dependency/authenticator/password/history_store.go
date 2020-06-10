package password

import (
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/skygeario/skygear-server/pkg/core/db"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type HistoryStore interface {
	// CreatePasswordHistory create new password history.
	CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error

	// GetPasswordHistory returns a slice of PasswordHistory of the given user
	//
	// If historySize is greater than 0, the returned slice contains history
	// of that size.
	// If historyDays is greater than 0, the returned slice contains history
	// up to now.
	//
	// If both historySize and historyDays are greater than 0, the returned slice
	// is the longer of the result.
	GetPasswordHistory(userID string, historySize, historyDays int) ([]History, error)

	// RemovePasswordHistory removes old password history.
	// It uses GetPasswordHistory to query active history and then purge old history.
	RemovePasswordHistory(userID string, historySize, historyDays int) error
}

type HistoryStoreImpl struct {
	timeProvider coreTime.Provider
	sqlBuilder   db.SQLBuilder
	sqlExecutor  db.SQLExecutor
}

func NewHistoryStore(timeProvider coreTime.Provider, builder db.SQLBuilder, executor db.SQLExecutor) *HistoryStoreImpl {
	return &HistoryStoreImpl{
		timeProvider: timeProvider,
		sqlBuilder:   builder,
		sqlExecutor:  executor,
	}
}

func (p *HistoryStoreImpl) CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error {
	updateBuilder := p.insertPasswordHistoryBuilder(
		userID,
		hashedPassword,
		loggedAt,
	)
	if _, err := p.sqlExecutor.ExecWith(updateBuilder); err != nil {
		return err
	}
	return nil
}

func (p *HistoryStoreImpl) GetPasswordHistory(userID string, historySize, historyDays int) ([]History, error) {
	var err error
	var sizeHistory, daysHistory []History
	t := p.timeProvider.NowUTC()

	if historySize > 0 {
		sizeBuilder := p.basePasswordHistoryBuilder(userID).Limit(uint64(historySize))
		sizeHistory, err = p.doQueryPasswordHistory(sizeBuilder)
		if err != nil {
			return nil, err
		}
	}

	if historyDays > 0 {
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		since := startOfDay.AddDate(0, 0, -historyDays)
		daysBuilder := p.basePasswordHistoryBuilder(userID).
			Where("logged_at >= ?", since)
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

func (p *HistoryStoreImpl) RemovePasswordHistory(userID string, historySize, historyDays int) error {
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

	builder := p.sqlBuilder.Tenant().
		Delete(p.sqlBuilder.FullTableName("password_history")).
		Where("user_id = ?", userID).
		Where("id NOT IN ("+sq.Placeholders(len(ids))+")", ids...).
		Where("created_at < ?", oldestTime)

	_, err = p.sqlExecutor.ExecWith(builder)
	return err
}

func (p *HistoryStoreImpl) basePasswordHistoryBuilder(userID string) db.SelectBuilder {
	return p.sqlBuilder.Tenant().
		Select("id", "user_id", "password", "created_at").
		From(p.sqlBuilder.FullTableName("password_history")).
		Where("user_id = ?", userID).
		OrderBy("logged_at DESC")
}

func (p *HistoryStoreImpl) insertPasswordHistoryBuilder(userID string, hashedPassword []byte, createdAt time.Time) db.InsertBuilder {
	return p.sqlBuilder.Tenant().
		Insert(p.sqlBuilder.FullTableName("password_history")).
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

func (p *HistoryStoreImpl) doQueryPasswordHistory(builder db.SelectBuilder) ([]History, error) {
	rows, err := p.sqlExecutor.QueryWith(builder)
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
			loggedAt          time.Time
		)
		if err := rows.Scan(&id, &userID, &hashedPasswordStr, &loggedAt); err != nil {
			return nil, err
		}
		passwordHistory := History{
			ID:             id,
			UserID:         userID,
			HashedPassword: []byte(hashedPasswordStr),
			CreatedAt:      loggedAt,
		}
		out = append(out, passwordHistory)
	}
	return out, nil
}
