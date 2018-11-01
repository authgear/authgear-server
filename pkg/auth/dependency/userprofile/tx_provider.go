package userprofile

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeUserProfileImpl struct {
	impl      *userProfileStoreImpl
	txContext db.SafeTxContext
}

func NewSafeProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) UserProfileStore {
	return &safeUserProfileImpl{
		impl:      newUserProfileStore(builder, executor, logger),
		txContext: txContext,
	}
}

func (s *safeUserProfileImpl) CreateUserProfile(userID string, userProfile map[string]interface{}) (err error) {
	s.txContext.EnsureTx()
	return s.impl.CreateUserProfile(userID, userProfile)
}

func (s *safeUserProfileImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	s.txContext.EnsureTx()
	return s.impl.GetUserProfile(userID, userProfile)
}
