package userprofile

import (
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type safeUserProfileImpl struct {
	impl      *storeImpl
	txContext db.SafeTxContext
}

// NewSafeProvider returns a auth gear user profile store implementation
func NewSafeProvider(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	loggerFactory logging.Factory,
	txContext db.SafeTxContext,
) Store {
	return &safeUserProfileImpl{
		impl:      newUserProfileStore(builder, executor, loggerFactory),
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

func (s *safeUserProfileImpl) UpdateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	s.txContext.EnsureTx()
	return s.impl.UpdateUserProfile(userID, data)
}
