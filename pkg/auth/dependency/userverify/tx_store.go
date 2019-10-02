package userverify

import (
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type safeStoreImpl struct {
	impl      *storeImpl
	txContext db.SafeTxContext
}

func NewSafeStore(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	loggerFactory logging.Factory,
	txContext db.SafeTxContext,
) Store {
	return &safeStoreImpl{
		impl:      newStore(builder, executor, loggerFactory),
		txContext: txContext,
	}
}

func (s *safeStoreImpl) CreateVerifyCode(code *VerifyCode) error {
	s.txContext.EnsureTx()
	return s.impl.CreateVerifyCode(code)
}

func (s *safeStoreImpl) MarkConsumed(codeID string) error {
	s.txContext.EnsureTx()
	return s.impl.MarkConsumed(codeID)
}

func (s *safeStoreImpl) GetVerifyCodeByUser(userID string) (*VerifyCode, error) {
	s.txContext.EnsureTx()
	return s.impl.GetVerifyCodeByUser(userID)
}

var _ Store = &safeStoreImpl{}
