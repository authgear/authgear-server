package userprofile

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeUserProfileImpl struct {
	impl      *storeImpl
	txContext db.SafeTxContext
}

func NewSafeProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) Store {
	return &safeUserProfileImpl{
		impl:      newUserProfileStore(builder, executor, logger),
		txContext: txContext,
	}
}

func (s *safeUserProfileImpl) CreateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	s.txContext.EnsureTx()
	return s.impl.CreateUserProfile(userID, data)
}

func (s *safeUserProfileImpl) GetUserProfile(userID string) (profile UserProfile, err error) {
	s.txContext.EnsureTx()
	return s.impl.GetUserProfile(userID)
}
