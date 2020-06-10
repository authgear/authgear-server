package password

import (
	"sort"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type mockPasswordHistoryStoreImpl struct {
	Data    map[string][]History // userID as key
	TimeNow func() time.Time
}

func (m *mockPasswordHistoryStoreImpl) CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error {
	if _, ok := m.Data[userID]; !ok {
		m.Data[userID] = []History{}
	}
	ph := History{
		ID:             uuid.New(),
		UserID:         userID,
		HashedPassword: hashedPassword,
		CreatedAt:      loggedAt,
	}
	uph := append(m.Data[userID], ph)
	sort.Slice(uph, func(i, j int) bool { return uph[i].CreatedAt.After(uph[j].CreatedAt) })
	m.Data[userID] = uph
	return nil
}

func (m *mockPasswordHistoryStoreImpl) GetPasswordHistory(userID string, historySize, historyDays int) ([]History, error) {
	uph, ok := m.Data[userID]
	if !ok || len(uph) <= 0 {
		return []History{}, nil
	}

	t := m.TimeNow()
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	since := startOfDay.AddDate(0, 0, -historyDays)

	index := 0
	for i, ph := range uph {
		if i >= historySize && ph.CreatedAt.Before(since) {
			break
		}
		index = i
	}

	return uph[:index+1], nil
}

func (m *mockPasswordHistoryStoreImpl) RemovePasswordHistory(userID string, historySize, historyDays int) error {
	uph, err := m.GetPasswordHistory(userID, historySize, historySize)
	if err != nil {
		return err
	}

	m.Data[userID] = uph
	return nil
}
