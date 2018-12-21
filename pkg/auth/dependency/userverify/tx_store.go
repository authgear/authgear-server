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
		impl: newStore(builder, executor, logger),
	}
}

func (s *safeStoreImpl) CreateVerifyCode(code *VerifyCode) error {
	s.txContext.EnsureTx()
	return s.impl.CreateVerifyCode(code)
}

func (s *safeStoreImpl) UpdateVerifyCode(code *VerifyCode) error {
	s.txContext.EnsureTx()
	return s.impl.UpdateVerifyCode(code)
}

func (s *safeStoreImpl) GetVerifyCodeByCode(code string, vCode *VerifyCode) error {
	s.txContext.EnsureTx()
	return s.impl.GetVerifyCodeByCode(code, vCode)
}
