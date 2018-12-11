package passwordhistory

import (
	"sort"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type MockPasswordHistoryStoreImpl struct {
	Data        map[string][]PasswordHistory // userID as key
	TimeNowfunc MockTimeNowfunc
}

type MockTimeNowfunc func() time.Time

func NewMockPasswordHistoryStore() *MockPasswordHistoryStoreImpl {
	return NewMockPasswordHistoryStoreWithData(
		map[string][]PasswordHistory{},
		func() time.Time { return time.Now().UTC() },
	)
}

func NewMockPasswordHistoryStoreWithData(data map[string][]PasswordHistory, timeNowFunc MockTimeNowfunc) *MockPasswordHistoryStoreImpl {
	return &MockPasswordHistoryStoreImpl{
		Data:        data,
		TimeNowfunc: timeNowFunc,
	}
}

func (m *MockPasswordHistoryStoreImpl) CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error {
	if _, ok := m.Data[userID]; !ok {
		m.Data[userID] = []PasswordHistory{}
	}
	ph := PasswordHistory{
		ID:             uuid.New(),
		UserID:         userID,
		HashedPassword: hashedPassword,
		LoggedAt:       loggedAt,
	}
	uph := append(m.Data[userID], ph)
	sort.Slice(uph, func(i, j int) bool { return uph[i].LoggedAt.After(uph[j].LoggedAt) })
	m.Data[userID] = uph
	return nil
}

func (m *MockPasswordHistoryStoreImpl) GetPasswordHistory(userID string, historySize, historyDays int) ([]PasswordHistory, error) {
	uph, ok := m.Data[userID]
	if !ok || len(uph) <= 0 {
		return []PasswordHistory{}, nil
	}

	t := m.TimeNowfunc()
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	since := startOfDay.AddDate(0, 0, -historyDays)

	index := 0
	for i, ph := range uph {
		if i >= historySize && ph.LoggedAt.Before(since) {
			break
		}
		index = i
	}

	return uph[:index+1], nil
}

func (m *MockPasswordHistoryStoreImpl) RemovePasswordHistory(userID string, historySize, historyDays int) error {
	uph, err := m.GetPasswordHistory(userID, historySize, historySize)
	if err != nil {
		return err
	}

	m.Data[userID] = uph
	return nil
}
