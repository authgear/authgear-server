package pq

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeRoleStore struct {
	impl      *roleStore
	txContext db.SafeTxContext
}

func NewSafeRoleStore(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) role.Store {
	return &safeRoleStore{
		impl:      newRoleStore(builder, executor, logger),
		txContext: txContext,
	}
}

func (s *safeRoleStore) CreateRoles(roles []string) error {
	s.txContext.EnsureTx()
	return s.impl.CreateRoles(roles)
}

func (s *safeRoleStore) QueryRoles(roles []string) ([]role.Role, error) {
	s.txContext.EnsureTx()
	return s.impl.QueryRoles(roles)
}

func (s *safeRoleStore) GetDefaultRoles() ([]string, error) {
	s.txContext.EnsureTx()
	return s.impl.GetDefaultRoles()
}

func (s *safeRoleStore) SetAdminRoles(roles []string) error {
	s.txContext.EnsureTx()
	return s.impl.SetAdminRoles(roles)
}

func (s *safeRoleStore) SetDefaultRoles(roles []string) error {
	s.txContext.EnsureTx()
	return s.impl.SetDefaultRoles(roles)
}
