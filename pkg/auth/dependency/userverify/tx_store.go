package userverify

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeStoreImpl struct {
	impl      *storeImpl
	txContext db.SafeTxContext
}

func NewSafeStore(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) Store {
	return &safeStoreImpl{
		impl:      newStore(builder, executor, logger),
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

func (s *safeStoreImpl) GetVerifyCodeByCode(userID string, code string) (VerifyCode, error) {
	s.txContext.EnsureTx()
	return s.impl.GetVerifyCodeByCode(userID, code)
}

var _ Store = &safeStoreImpl{}
