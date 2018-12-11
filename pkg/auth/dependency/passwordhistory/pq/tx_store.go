package pq

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safePasswordHistoryStore struct {
	impl      *passwordHistoryStore
	txContext db.SafeTxContext
}

func NewSafePasswordHistoryStore(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) passwordhistory.Store {
	return &safePasswordHistoryStore{
		impl:      newPasswordHistoryStore(builder, executor, logger),
		txContext: txContext,
	}
}

func (s *safePasswordHistoryStore) CreatePasswordHistory(userID string, hashedPassword []byte, loggedAt time.Time) error {
	s.txContext.EnsureTx()
	return s.impl.CreatePasswordHistory(userID, hashedPassword, loggedAt)
}

func (s *safePasswordHistoryStore) GetPasswordHistory(userID string, historySize, historyDays int) ([]passwordhistory.PasswordHistory, error) {
	s.txContext.EnsureTx()
	return s.impl.GetPasswordHistory(userID, historySize, historyDays)
}

func (s *safePasswordHistoryStore) RemovePasswordHistory(userID string, historySize, historyDays int) error {
	s.txContext.EnsureTx()
	return s.impl.RemovePasswordHistory(userID, historySize, historyDays)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ passwordhistory.Store = &safePasswordHistoryStore{}
)
